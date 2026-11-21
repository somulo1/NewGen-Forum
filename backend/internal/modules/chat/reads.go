package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

// maxReaders bounds one answer for a very large forum.
const maxReaders = 2000

// Reader is one member's reading place in a conversation: At is when the last message they have read was sent (nil when they
// have read nothing yet). A message is read by someone when it was sent at or before that time.
type Reader struct {
	Person uuid.UUID  `json:"person"`
	At     *time.Time `json:"at"`
}

// MarkRead moves the caller's own reading place forward to a message of this conversation. It never moves back, never takes a
// message from another conversation, and only ever changes the caller's own row: the person comes from the signed-in account.
func (s *Service) MarkRead(ctx context.Context, p *tenant.Principal, id, upTo uuid.UUID) error {
	var at *time.Time
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if _, _, _, _, err := s.lockConvNoLock(ctx, tx, p, id); err != nil {
			return err
		}
		var sent time.Time
		if err := tx.QueryRow(ctx, `select created_at from hub.messages where org_id=$1 and id=$2 and conversation_id=$3`, p.Org, upTo, id).Scan(&sent); err != nil {
			return nil // not a message of this conversation: nothing to mark
		}
		tag, err := tx.Exec(ctx, `insert into hub.read_cursors as rc (org_id, conversation_id, account_id, last_read) values ($1,$2,$3,$4)
			on conflict (org_id, conversation_id, account_id) do update set last_read = excluded.last_read, updated_at=now()
			where rc.last_read is null or exists (select 1 from hub.messages o where o.org_id=rc.org_id and o.id=rc.last_read and o.created_at < $5)
				or not exists (select 1 from hub.messages o where o.org_id=rc.org_id and o.id=rc.last_read)`, p.Org, id, p.Account, upTo, sent)
		if err == nil && tag.RowsAffected() > 0 {
			at = &sent
		}
		return err
	})
	if err == nil && at != nil {
		d, _ := json.Marshal(map[string]any{"conversation": id, "person": p.Person, "at": at})
		s.A.Publish(ctx, realtime.Event{Type: "conversation.read", Channel: realtime.ConvChannel(id), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
	}
	return err
}

// Reads says how far each current member has read. Only members of the conversation can ask.
func (s *Service) Reads(ctx context.Context, p *tenant.Principal, id uuid.UUID) ([]Reader, error) {
	out := []Reader{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if _, _, _, _, err := s.lockConvNoLock(ctx, tx, p, id); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select pe.id, m.created_at from (select distinct account_id from hub.conversation_members where conversation_id=$1 and left_at is null) cm
			join hub.people pe on pe.account_id = cm.account_id
			left join hub.read_cursors rc on rc.conversation_id=$1 and rc.account_id = cm.account_id
			left join hub.messages m on m.org_id=rc.org_id and m.id = rc.last_read and m.conversation_id=$1
			limit $2`, id, maxReaders)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var r Reader
			if err := rs.Scan(&r.Person, &r.At); err != nil {
				return err
			}
			out = append(out, r)
		}
		return rs.Err()
	})
	return out, err
}

func (s *Service) readsRoute(r chi.Router) {
	r.Get("/conversations/{id}/reads", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		list, err := s.Reads(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"readers": list})
		return nil
	}))
}
