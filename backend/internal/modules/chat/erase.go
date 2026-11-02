package chat

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/realtime"
)

// Moments live for 24 hours. After that they are hidden at once (the lists leave them out) and erased for good by the daily sweep
// (00:02 Nairobi time) together with their pictures and videos.
const MomentLife = 24 * time.Hour

// linkFiles records which uploaded files a message carries (only the sender's own, still-live files can be linked).
func linkFiles(ctx context.Context, tx pgx.Tx, org, msg, person uuid.UUID, files []uuid.UUID) error {
	if len(files) == 0 {
		return nil
	}
	if len(files) > 20 {
		return httpx.Bad("A message can carry up to 20 files.")
	}
	_, err := tx.Exec(ctx, `insert into hub.message_files (org_id, message_id, file_id)
		select $1, $2, f.id from hub.files f where f.org_id=$1 and f.id = any($3) and f.uploader=$4 and f.status <> 'deleted'
		on conflict do nothing`, org, msg, files, person)
	return err
}

// eraseAround removes what hangs on the given messages, inside the caller's transaction: the comments, likes and edits written
// on them lose their text, and every file they carry is marked deleted with its bytes queued for erasure (hub.blob_purge).
// Call purge afterwards (once the transaction has committed) to erase the bytes right away.
func (s *Service) eraseAround(ctx context.Context, tx pgx.Tx, org uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	// everything that hangs on the message goes with it, however deep: comments, replies to comments, likes, votes, edits
	all := append([]uuid.UUID{}, ids...)
	frontier := ids
	for depth := 0; depth < 8 && len(frontier) > 0; depth++ {
		rs, err := tx.Query(ctx, `update hub.messages set ciphertext='\x00'::bytea, deleted_at=now(), `+integrity.Bump+`
			where org_id=$1 and reply_to = any($2) and content_type in ('comment','reaction','edit','vote') and deleted_at is null returning id`, org, frontier)
		if err != nil {
			return err
		}
		frontier = nil
		for rs.Next() {
			var k uuid.UUID
			if err := rs.Scan(&k); err != nil {
				rs.Close()
				return err
			}
			all = append(all, k)
			frontier = append(frontier, k)
		}
		rs.Close()
		if err := rs.Err(); err != nil {
			return err
		}
	}
	if err := s.A.Sign.Seal(ctx, tx, "messages", "org_id=$1 and id = any($2)", org, all); err != nil {
		return err
	}
	type gone struct {
		id  uuid.UUID
		key *string
	}
	var files []gone
	fr, err := tx.Query(ctx, `select f.id, f.blob_key from hub.files f where f.org_id=$1 and f.status <> 'deleted'
		and f.id in (select file_id from hub.message_files where org_id=$1 and message_id = any($2)) for update`, org, all)
	if err != nil {
		return err
	}
	var fids []uuid.UUID
	for fr.Next() {
		var g gone
		if err := fr.Scan(&g.id, &g.key); err != nil {
			fr.Close()
			return err
		}
		files = append(files, g)
		fids = append(fids, g.id)
	}
	fr.Close()
	if err := fr.Err(); err != nil {
		return err
	}
	if len(fids) > 0 {
		if _, err := tx.Exec(ctx, `update hub.files set status='deleted', envelope='\x00'::bytea, blob_key=null, updated_at=now(), `+integrity.Bump+` where org_id=$1 and id = any($2)`, org, fids); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "files", "org_id=$1 and id = any($2)", org, fids); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update hub.shares set status='Revoked', `+integrity.Bump+` where file_id = any($1) and status <> 'Revoked'`, fids); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "shares", "file_id = any($1)", fids); err != nil {
			return err
		}
		for _, g := range files {
			if g.key != nil {
				if _, err := tx.Exec(ctx, `insert into hub.blob_purge (org_id, ref) values ($1,$2)`, org, *g.key); err != nil {
					return err
				}
			}
		}
	}
	_, err = tx.Exec(ctx, `delete from hub.message_files where org_id=$1 and message_id = any($2)`, org, all)
	return err
}

// touchChanged marks a conversation as changed now, so devices that were offline know to catch up on it.
func touchChanged(ctx context.Context, tx pgx.Tx, org, conv uuid.UUID) error {
	_, err := tx.Exec(ctx, `update hub.conversation_stats set changed_at=now() where org_id=$1 and conversation_id=$2`, org, conv)
	return err
}

// purge erases queued file contents from storage now (when this node can; the worker retries whatever is left).
func (s *Service) purge(ctx context.Context) {
	if s.Purge != nil {
		s.Purge(context.WithoutCancel(ctx))
	}
}

// ExpireMoments erases every moment older than 24 hours, with its pictures and videos, in every organization.
func (s *Service) ExpireMoments(ctx context.Context) (int, error) {
	var orgs []uuid.UUID
	if err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select id from hub.orgs where status in ('active','suspended','closing')`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var id uuid.UUID
			if err := rs.Scan(&id); err != nil {
				return err
			}
			orgs = append(orgs, id)
		}
		return rs.Err()
	}); err != nil {
		return 0, err
	}
	total := 0
	for _, org := range orgs {
		for {
			type mo struct{ id, conv uuid.UUID }
			var batch []mo
			err := s.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
				rs, err := tx.Query(ctx, `select id, conversation_id from hub.messages where org_id=$1 and content_type='moment' and deleted_at is null
					and created_at < now() - make_interval(secs => $2) order by id limit 200 for update skip locked`, org, MomentLife.Seconds())
				if err != nil {
					return err
				}
				var ids []uuid.UUID
				for rs.Next() {
					var m mo
					if err := rs.Scan(&m.id, &m.conv); err != nil {
						rs.Close()
						return err
					}
					batch = append(batch, m)
					ids = append(ids, m.id)
				}
				rs.Close()
				if err := rs.Err(); err != nil || len(ids) == 0 {
					return err
				}
				if _, err := tx.Exec(ctx, `update hub.messages set ciphertext='\x00'::bytea, deleted_at=now(), `+integrity.Bump+` where org_id=$1 and id = any($2)`, org, ids); err != nil {
					return err
				}
				for _, m := range batch {
					if err := touchChanged(ctx, tx, org, m.conv); err != nil {
						return err
					}
				}
				return s.eraseAround(ctx, tx, org, ids)
			})
			if err != nil {
				slog.Error("moment sweep", "org", org, "err", err)
				break
			}
			for _, m := range batch {
				s.A.Publish(ctx, realtime.Event{Type: "message.deleted", Channel: realtime.ConvChannel(m.conv), Org: org.String(), Data: []byte(`{"id":"` + m.id.String() + `","conversation":"` + m.conv.String() + `"}`)})
			}
			total += len(batch)
			if len(batch) < 200 {
				break
			}
		}
	}
	s.purge(ctx)
	return total, nil
}
