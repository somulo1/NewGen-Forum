package chat

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
	"hub/internal/tenant"
)

var neverCleared = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

// clearedAt is the time before which this person sees nothing in a conversation: when they deleted it for themselves, or, in a
// forum, when they joined (someone added later never sees what was said before, and is not told it exists). It is a time long
// ago when neither applies.
func (s *Service) clearedAt(ctx context.Context, tx pgx.Tx, conv, account uuid.UUID) time.Time {
	var t time.Time
	if err := tx.QueryRow(ctx, `select greatest(
			coalesce((select cleared_at from hub.conversation_clears where conversation_id=$1 and account_id=$2), '-infinity'::timestamptz),
			coalesce((select m.joined_at from hub.conversation_members m join hub.conversations c on c.id = m.conversation_id
				where m.conversation_id=$1 and m.account_id=$2 and c.kind='forum'), '-infinity'::timestamptz))`, conv, account).Scan(&t); err != nil || t.Before(neverCleared) {
		return neverCleared
	}
	return t
}

// Clear deletes a private chat or a forum for the caller only: their copy of the history up to now is gone for good (it can
// no longer be read, listed or fetched), and the conversation leaves their list until someone writes there again. Other
// members keep everything, and the caller stays a member, so what is written next arrives as usual.
func (s *Service) Clear(ctx context.Context, p *tenant.Principal, conv uuid.UUID) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		kind, _, _, _, err := s.lockConvNoLock(ctx, tx, p, conv)
		if err != nil {
			return err
		}
		if kind != "dm" && kind != "forum" {
			return httpx.Bad("Only private chats and forums can be deleted from your list.")
		}
		if _, err := tx.Exec(ctx, `insert into hub.conversation_clears (org_id, conversation_id, account_id, cleared_at) values ($1,$2,$3, clock_timestamp())
			on conflict (org_id, conversation_id, account_id) do update set cleared_at = excluded.cleared_at`, p.Org, conv, p.Account); err != nil {
			return err
		}
		// what they had not read is read now: nothing of the deleted history should count as unread
		_, err = tx.Exec(ctx, `delete from hub.read_cursors where conversation_id=$1 and account_id=$2`, conv, p.Account)
		return err
	})
}

func (s *Service) clearRoutes(w http.ResponseWriter, r *http.Request, p *tenant.Principal, id uuid.UUID) error {
	if err := s.Clear(r.Context(), p, id); err != nil {
		return err
	}
	w.WriteHeader(204)
	return nil
}

// Cleared lists the private chats and forums this person deleted from their list, so the app can still hear when someone
// writes in one of them.
func (s *Service) Cleared(ctx context.Context, p *tenant.Principal) ([]uuid.UUID, error) {
	out := []uuid.UUID{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select cl.conversation_id from hub.conversation_clears cl join hub.conversations c on c.id = cl.conversation_id
			join hub.conversation_members m on m.conversation_id = c.id and m.account_id = cl.account_id and m.left_at is null
			where cl.account_id=$1 and c.kind in ('dm','forum')`, p.Account)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var id uuid.UUID
			if err := rs.Scan(&id); err != nil {
				return err
			}
			out = append(out, id)
		}
		return rs.Err()
	})
	return out, err
}
