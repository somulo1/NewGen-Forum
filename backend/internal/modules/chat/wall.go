package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/blob"
	"hub/internal/httpx"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

// The community wall's likes and votes: rows the server keeps and counts (one reaction per person per post, comment or moment;
// one final vote per person per poll), with running totals in wall_counts so a page reads its counts in one small query.

// wallItem checks that a message is a live post, comment or moment on a wall the caller belongs to, and returns its wall and type.
func (s *Service) wallItem(ctx context.Context, tx pgx.Tx, p *tenant.Principal, id uuid.UUID) (conv uuid.UUID, typ string, body json.RawMessage, err error) {
	var ct []byte
	var gone bool
	err = tx.QueryRow(ctx, `select m.conversation_id, m.content_type, m.ciphertext, m.deleted_at is not null from hub.messages m
		join hub.conversations c on c.id = m.conversation_id and c.kind = 'wall'
		join hub.conversation_members cm on cm.conversation_id = c.id and cm.account_id = $3 and cm.left_at is null
		where m.org_id=$1 and m.id=$2`, p.Org, id, p.Account).Scan(&conv, &typ, &ct, &gone)
	if errors.Is(err, pgx.ErrNoRows) || gone {
		return conv, typ, nil, httpx.ErrNotFound
	}
	if err != nil {
		return
	}
	if err = s.checkBlocked(ctx, tx, p.Org, "wall"); err != nil {
		return
	}
	if IsPlain(ct) {
		dek, _, e := s.A.OrgDEK(ctx, tx, p.Org)
		if e != nil {
			return conv, typ, nil, e
		}
		body, _ = OpenPlain(dek, p.Org, id, ct)
	}
	return
}

func bumpCount(ctx context.Context, tx pgx.Tx, org, item uuid.UUID, key string, d int) error {
	if key == "" || d == 0 {
		return nil
	}
	_, err := tx.Exec(ctx, `insert into hub.wall_counts as w (org_id, item_id, key, n) values ($1,$2,$3,greatest($4,0))
		on conflict (org_id, item_id, key) do update set n = greatest(w.n + $4, 0)`, org, item, key, d)
	return err
}

// React sets the caller's reaction on a wall item ("" takes it back). One per person: a new kind replaces the old one.
func (s *Service) React(ctx context.Context, p *tenant.Principal, id uuid.UUID, kind string) error {
	if kind != "" && kind != "like" && kind != "dislike" && kind != "love" {
		return httpx.Bad("Choose like, dislike or love.")
	}
	var conv uuid.UUID
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		c, typ, _, err := s.wallItem(ctx, tx, p, id)
		if err != nil {
			return err
		}
		conv = c
		if typ == "reaction" || typ == "vote" || typ == "share" {
			return httpx.Bad("You cannot react to that.")
		}
		if err := authz.Require(ctx, tx, p.Person, "wall", "View"); err != nil {
			return err
		}
		var old string
		_ = tx.QueryRow(ctx, `select kind from hub.wall_reactions where item_id=$1 and person_id=$2 for update`, id, p.Person).Scan(&old)
		if old == kind {
			return nil
		}
		if old != "" {
			if _, err := tx.Exec(ctx, `delete from hub.wall_reactions where item_id=$1 and person_id=$2`, id, p.Person); err != nil {
				return err
			}
			if err := bumpCount(ctx, tx, p.Org, id, old, -1); err != nil {
				return err
			}
		}
		if kind != "" {
			if _, err := tx.Exec(ctx, `insert into hub.wall_reactions (org_id, item_id, person_id, kind) values ($1,$2,$3,$4)`, p.Org, id, p.Person, kind); err != nil {
				return err
			}
			if err := bumpCount(ctx, tx, p.Org, id, kind, 1); err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx, `update hub.conversation_stats set changed_at=now() where org_id=$1 and conversation_id=$2`, p.Org, conv)
		return err
	})
	if err == nil {
		s.wallChanged(ctx, p, conv, id)
	}
	return err
}

// Vote records the caller's vote on a poll. A vote is final: the first one counts.
func (s *Service) Vote(ctx context.Context, p *tenant.Principal, id uuid.UUID, option string) error {
	var conv uuid.UUID
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		c, _, body, err := s.wallItem(ctx, tx, p, id)
		if err != nil {
			return err
		}
		conv = c
		if err := authz.Require(ctx, tx, p.Person, "wall", "View"); err != nil {
			return err
		}
		var pl struct {
			Poll []struct {
				ID string `json:"id"`
			} `json:"poll"`
		}
		_ = json.Unmarshal(body, &pl)
		ok := false
		for _, o := range pl.Poll {
			if o.ID == option && option != "" {
				ok = true
			}
		}
		if !ok {
			return httpx.Bad("That is not one of the poll's choices.")
		}
		tag, err := tx.Exec(ctx, `insert into hub.wall_votes (org_id, post_id, person_id, option) values ($1,$2,$3,$4) on conflict do nothing`, p.Org, id, p.Person, option)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.Err(409, "already_voted", "You have already voted on this poll. Your first vote is final.")
		}
		if err := bumpCount(ctx, tx, p.Org, id, "vote:"+option, 1); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `update hub.conversation_stats set changed_at=now() where org_id=$1 and conversation_id=$2`, p.Org, conv)
		return err
	})
	if err == nil {
		s.wallChanged(ctx, p, conv, id)
	}
	return err
}

func (s *Service) wallChanged(ctx context.Context, p *tenant.Principal, conv, item uuid.UUID) {
	d, _ := json.Marshal(map[string]any{"item": item})
	s.A.Publish(ctx, realtime.Event{Type: "wall.changed", Channel: realtime.ConvChannel(conv), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
}

// wallExtras adds counts, my reaction and vote, and the first names per reaction to a page of wall messages: four queries for the
// page, whatever its size.
func (s *Service) wallExtras(ctx context.Context, tx pgx.Tx, p *tenant.Principal, list []Message) error {
	if len(list) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(list))
	at := map[uuid.UUID]int{}
	for i, m := range list {
		if !m.Deleted {
			ids = append(ids, m.ID)
			at[m.ID] = i
		}
	}
	if len(ids) == 0 {
		return nil
	}
	rs, err := tx.Query(ctx, `select item_id, key, n from hub.wall_counts where item_id = any($1) and n > 0`, ids)
	if err != nil {
		return err
	}
	for rs.Next() {
		var id uuid.UUID
		var k string
		var n int
		if err := rs.Scan(&id, &k, &n); err != nil {
			rs.Close()
			return err
		}
		m := &list[at[id]]
		if m.Counts == nil {
			m.Counts = map[string]int{}
		}
		m.Counts[k] = n
	}
	rs.Close()
	rs, err = tx.Query(ctx, `select item_id, kind from hub.wall_reactions where item_id = any($1) and person_id=$2`, ids, p.Person)
	if err != nil {
		return err
	}
	for rs.Next() {
		var id uuid.UUID
		var k string
		if err := rs.Scan(&id, &k); err != nil {
			rs.Close()
			return err
		}
		list[at[id]].MyReaction = k
	}
	rs.Close()
	rs, err = tx.Query(ctx, `select post_id, option from hub.wall_votes where post_id = any($1) and person_id=$2`, ids, p.Person)
	if err != nil {
		return err
	}
	for rs.Next() {
		var id uuid.UUID
		var o string
		if err := rs.Scan(&id, &o); err != nil {
			rs.Close()
			return err
		}
		list[at[id]].MyVote = o
	}
	rs.Close()
	// the most recent 20 names per item and reaction (who liked, who loved), for "liked by" lines
	rs, err = tx.Query(ctx, `select item_id, kind, name from (
			select r.item_id, r.kind, pe.first_name || ' ' || pe.last_name as name, row_number() over (partition by r.item_id, r.kind order by r.at desc) as rn
			from hub.wall_reactions r join hub.people pe on pe.id = r.person_id where r.item_id = any($1)) x where rn <= 20`, ids)
	if err != nil {
		return err
	}
	defer rs.Close()
	for rs.Next() {
		var id uuid.UUID
		var k, n string
		if err := rs.Scan(&id, &k, &n); err != nil {
			return err
		}
		m := &list[at[id]]
		if m.Names == nil {
			m.Names = map[string][]string{}
		}
		m.Names[k] = append(m.Names[k], n)
	}
	return rs.Err()
}

// ---- plain attachments (the wall) ----

type PlainFile struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

var safeFileType = regexp.MustCompile(`^(image/(png|jpeg|gif|webp|avif|bmp)|video/(mp4|webm|ogg|quicktime)|audio/(mpeg|ogg|wav|webm|mp4)|application/pdf|text/(plain|csv))$`)

func cleanFileType(t string) string {
	t = strings.ToLower(strings.TrimSpace(strings.SplitN(t, ";", 2)[0]))
	if safeFileType.MatchString(t) {
		return t
	}
	return "application/octet-stream"
}

// UploadPlain stores an attachment for a server-readable conversation (the wall).
func (s *Service) UploadPlain(ctx context.Context, p *tenant.Principal, conv uuid.UUID, name, typ string, body io.Reader, declared int64) (*PlainFile, error) {
	if s.Store == nil {
		return nil, httpx.Err(503, "no_storage", "Files cannot be stored right now.")
	}
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 255 || strings.ContainsAny(name, "/\\\x00") {
		return nil, httpx.Bad("Give the file a name.")
	}
	if declared <= 0 || declared > s.A.Cfg.MaxUpload {
		return nil, httpx.Err(413, "too_large", fmt.Sprintf("Files can be up to %d MB.", s.A.Cfg.MaxUpload>>20))
	}
	if err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var kind string
		if err := tx.QueryRow(ctx, `select c.kind from hub.conversations c join hub.conversation_members m on m.conversation_id=c.id and m.account_id=$2 and m.left_at is null where c.id=$1`, conv, p.Account).Scan(&kind); err != nil {
			return httpx.ErrNotFound
		}
		if kind != "wall" {
			return httpx.Bad("Only the community wall takes plain attachments.")
		}
		if err := s.checkBlocked(ctx, tx, p.Org, kind); err != nil {
			return err
		}
		return authz.Require(ctx, tx, p.Person, "wall", "Post")
	}); err != nil {
		return nil, err
	}
	id := app.NewID()
	key, n, _, err := s.Store.Put(ctx, "w/"+p.Org.String()+"/"+id.String(), body, declared)
	if errors.Is(err, blob.ErrTooLarge) {
		return nil, httpx.Err(413, "too_large", "That file is larger than you said it was.")
	}
	if err != nil {
		return nil, err
	}
	if n != declared {
		_ = s.Store.Delete(ctx, key)
		return nil, httpx.Bad("The upload was cut short. Try again.")
	}
	out := &PlainFile{ID: id, Name: name, Type: cleanFileType(typ), Size: n}
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `insert into hub.plain_files (org_id, id, conversation_id, uploader, name, type, size, blob_key) values ($1,$2,$3,$4,$5,$6,$7,$8) returning created_at`,
			p.Org, id, conv, p.Person, name, out.Type, n, key).Scan(&out.CreatedAt)
	})
	if err != nil {
		_ = s.Store.Delete(ctx, key)
		return nil, err
	}
	return out, nil
}

// ServePlain sends a plain attachment to a member of its conversation.
func (s *Service) ServePlain(ctx context.Context, p *tenant.Principal, id uuid.UUID, w http.ResponseWriter, r *http.Request) error {
	if s.Store == nil {
		return httpx.ErrNotFound
	}
	var key, typ string
	if err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `select f.blob_key, f.type from hub.plain_files f join hub.conversation_members m on m.conversation_id = f.conversation_id and m.account_id=$2 and m.left_at is null
			where f.id=$1`, id, p.Account).Scan(&key, &typ)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		return err
	}); err != nil {
		return err
	}
	f, _, err := s.Store.Get(ctx, key)
	if errors.Is(err, blob.ErrNotFound) {
		return httpx.ErrNotFound
	}
	if err != nil {
		return err
	}
	defer f.Close()
	w.Header().Set("Content-Type", cleanFileType(typ))
	w.Header().Set("Cache-Control", "private, max-age=86400, immutable")
	w.Header().Set("Content-Disposition", "attachment")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, "", time.Time{}, f)
	return nil
}

// WallCounts returns the counts, my reaction and vote, and names for some wall items (a like on an older post updates just it).
func (s *Service) WallCounts(ctx context.Context, p *tenant.Principal, ids []uuid.UUID) ([]Message, error) {
	if len(ids) > 100 {
		ids = ids[:100]
	}
	out := []Message{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		// only items on walls the caller belongs to
		rs, err := tx.Query(ctx, `select m.id, m.conversation_id from hub.messages m join hub.conversations c on c.id = m.conversation_id and c.kind = 'wall'
			join hub.conversation_members cm on cm.conversation_id = c.id and cm.account_id = $3 and cm.left_at is null
			where m.org_id=$1 and m.id = any($2) and m.deleted_at is null`, p.Org, ids, p.Account)
		if err != nil {
			return err
		}
		for rs.Next() {
			var m Message
			if err := rs.Scan(&m.ID, &m.Conversation); err != nil {
				rs.Close()
				return err
			}
			out = append(out, m)
		}
		rs.Close()
		return s.wallExtras(ctx, tx, p, out)
	})
	return out, err
}
