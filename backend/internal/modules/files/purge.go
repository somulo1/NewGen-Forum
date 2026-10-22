package files

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/integrity"
)

// Erasing file contents. Deleting a file must free its space for real, on whichever file server holds it. The erase
// runs straight after the delete commits; when the server cannot be reached the reference stays in hub.blob_purge and
// the worker retries (PurgeDue) with growing pauses until the bytes are gone.

// purgeOne erases a queued reference now and clears it from the queue when that worked.
func (s *Service) purgeOne(ctx context.Context, id uuid.UUID, ref string) {
	err := s.Store.Delete(ctx, ref)
	_ = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		if err == nil {
			_, e := tx.Exec(ctx, `delete from hub.blob_purge where id=$1`, id)
			return e
		}
		_, e := tx.Exec(ctx, `update hub.blob_purge set attempts=attempts+1, last_error=$2, next_at=now() + interval '1 minute' where id=$1`, id, err.Error())
		return e
	})
	if err != nil {
		slog.Warn("could not erase file contents yet; the worker will retry", "ref", ref, "err", err)
	}
}

// erase removes bytes that never became a file (a failed or cut-short upload); if that fails they are queued.
func (s *Service) erase(ctx context.Context, org uuid.UUID, ref string) {
	if err := s.Store.Delete(ctx, ref); err == nil {
		return
	}
	var id uuid.UUID
	if err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `insert into hub.blob_purge (org_id, ref, next_at) values ($1,$2, now() + interval '1 minute') returning id`, org, ref).Scan(&id)
	}); err != nil {
		slog.Error("could not queue file contents for erasure", "ref", ref, "err", err)
	}
}

// PurgeDue erases queued contents that are due, up to `limit` at a time. Safe on several workers at once.
func (s *Service) PurgeDue(ctx context.Context, limit int) (done, failed int, err error) {
	type item struct {
		id       uuid.UUID
		ref      string
		attempts int
	}
	var due []item
	err = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		// claim the rows by pushing their next try out, so another worker does not take them too
		rs, err := tx.Query(ctx, `update hub.blob_purge set next_at = now() + interval '10 minutes'
			where id in (select id from hub.blob_purge where next_at <= now() order by next_at limit $1 for update skip locked)
			returning id, ref, attempts`, limit)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var it item
			if err := rs.Scan(&it.id, &it.ref, &it.attempts); err != nil {
				return err
			}
			due = append(due, it)
		}
		return rs.Err()
	})
	if err != nil {
		return 0, 0, err
	}
	for _, it := range due {
		derr := s.Store.Delete(ctx, it.ref)
		_ = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
			if derr == nil {
				_, e := tx.Exec(ctx, `delete from hub.blob_purge where id=$1`, it.id)
				return e
			}
			// 1, 2, 4 ... minutes, at most a day apart
			wait := time.Duration(1<<min(it.attempts, 11)) * time.Minute
			_, e := tx.Exec(ctx, `update hub.blob_purge set attempts=attempts+1, last_error=$2, next_at=now() + make_interval(secs => $3) where id=$1`, it.id, derr.Error(), wait.Seconds())
			return e
		})
		if derr == nil {
			done++
		} else {
			failed++
			slog.Warn("file contents not erased yet", "ref", it.ref, "attempts", it.attempts+1, "err", derr)
		}
	}
	return done, failed, nil
}

// EraseOrg erases every stored file of an organization from every file server, then all of its records. Contents that
// cannot be erased now stop the deletion (it is retried on the next run), so nothing is left behind on storage.
func (s *Service) EraseOrg(ctx context.Context, org uuid.UUID, records interface {
	FileRefs(context.Context, uuid.UUID) ([]string, error)
	EraseOrg(context.Context, uuid.UUID) error
}) error {
	refs, err := records.FileRefs(ctx, org)
	if err != nil {
		return err
	}
	for _, ref := range refs {
		if err := s.Store.Delete(ctx, ref); err != nil {
			return err
		}
	}
	return records.EraseOrg(ctx, org)
}

// SweepUploads closes uploads that were started but never finished (the browser closed, the connection dropped): after a day
// they are marked deleted so they stop counting as open uploads. No bytes exist for them, so nothing is erased.
func (s *Service) SweepUploads(ctx context.Context) (int, error) {
	var orgs []uuid.UUID
	if err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select id from hub.orgs where status='active'`)
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
		err := s.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
			rs, err := tx.Query(ctx, `update hub.files set status='deleted', envelope='\x00'::bytea, blob_key=null, updated_at=now(), `+integrity.Bump+`
				where status='uploading' and created_at < now() - interval '24 hours' returning id`)
			if err != nil {
				return err
			}
			var ids []uuid.UUID
			for rs.Next() {
				var id uuid.UUID
				if err := rs.Scan(&id); err != nil {
					rs.Close()
					return err
				}
				ids = append(ids, id)
			}
			rs.Close()
			if err := rs.Err(); err != nil {
				return err
			}
			for _, id := range ids {
				if err := s.A.Sign.Seal(ctx, tx, "files", "id=$1 and org_id=$2", id, org); err != nil {
					return err
				}
			}
			total += len(ids)
			return nil
		})
		if err != nil {
			return total, err
		}
	}
	return total, nil
}
