package comms

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/tenant"
)

func (s *Service) Private(r chi.Router) {
	org := func(h func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error) http.HandlerFunc {
		return httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
			p, err := tenant.MustOrg(r.Context())
			if err != nil {
				return err
			}
			return h(w, r, p)
		})
	}
	r.Post("/comms", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in SendIn
		if err := httpx.Decode(r, &in, 600<<10); err != nil {
			return err
		}
		c, req, err := s.Send(r.Context(), p, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, map[string]any{"comm": c, "approval": req})
		return nil
	}))
	r.Get("/comms", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		l, err := s.List(r.Context(), p, r.URL.Query().Get("type"))
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"comms": l})
		return nil
	}))
	// The sealed content of one communication the caller may see, for a client to open with the audience's key.
	r.Get("/comms/{id}/content", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		out, err := s.Content(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, out)
		return nil
	}))
	r.Post("/comms/contents", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in struct {
			IDs []uuid.UUID `json:"ids"`
		}
		if err := httpx.Decode(r, &in, 64<<10); err != nil {
			return err
		}
		if len(in.IDs) > 200 {
			return httpx.Bad("Ask for up to 200 at a time.")
		}
		l, err := s.Contents(r.Context(), p, in.IDs)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"contents": l})
		return nil
	}))
	r.Post("/comms/{id}/read", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		if err := s.MarkRead(r.Context(), p, id); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Post("/comms/{id}/acknowledge", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		if err := s.Acknowledge(r.Context(), p, id); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/comms/{id}/places", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		l, err := s.RepostPlaces(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"places": l})
		return nil
	}))
	r.Get("/comms/{id}/insights", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		l, err := s.Insights(r.Context(), p, id) // checks the right to see insights
		if err != nil {
			return err
		}
		t, err := s.GroupTotals(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"places": l, "totals": t})
		return nil
	}))
	r.Post("/comms/seen", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in struct {
			Conversation uuid.UUID   `json:"conversation"`
			Messages     []uuid.UUID `json:"messages"`
		}
		if err := httpx.Decode(r, &in, 16<<10); err != nil {
			return err
		}
		n, err := s.MarkSeen(r.Context(), p, in.Conversation, in.Messages)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"recorded": n})
		return nil
	}))
	r.Get("/comms/engagement", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		l, ppl, err := s.Engagement(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"comms": l, "people": ppl})
		return nil
	}))
	r.Post("/comms/{id}/remind", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		n, err := s.Remind(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"reminded": n})
		return nil
	}))
	r.Get("/comms/rules", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		m, err := s.Rules(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"rules": m})
		return nil
	}))
	r.Put("/comms/rules/{level}/{type}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in Rule
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		if err := s.SetRule(r.Context(), p, chi.URLParam(r, "level"), chi.URLParam(r, "type"), in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Post("/awards", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in struct {
			Type     string      `json:"type"`
			People   []uuid.UUID `json:"people"`
			Citation string      `json:"citation"`
			Targets  []string    `json:"targets"`
		}
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		a, req, err := s.Nominate(r.Context(), p, in.Type, in.People, in.Citation, in.Targets)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, map[string]any{"award": a, "approval": req})
		return nil
	}))
	r.Post("/awards/{id}/claim", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		var in struct {
			Target string `json:"target"`
		}
		if err := httpx.Decode(r, &in, 1<<10); err != nil {
			return err
		}
		if err := s.ClaimPost(r.Context(), p, id, in.Target); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/awards", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		l, cyc, err := s.AwardsCycle(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"awards": l, "cycle": cyc})
		return nil
	}))
}

// ---- repost ----

type Place struct {
	Conversation uuid.UUID `json:"conversation"`
	Kind         string    `json:"kind"` // wall | forum
	Name         string    `json:"name"`
	Member       bool      `json:"member"` // the caller can post there
}

// RepostPlaces lists where a published communication (and the copies sent with it) went, for someone who may repost it. The
// names of walls and forums are sealed, so the screen names them; this gives their ids and whether the caller is in them.
func (s *Service) RepostPlaces(ctx context.Context, p *tenant.Principal, id uuid.UUID) ([]Place, error) {
	out := []Place{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var typ, status string
		var batch *uuid.UUID
		if err := tx.QueryRow(ctx, `select type, status, batch from hub.comms where id=$1`, id).Scan(&typ, &status, &batch); err != nil {
			return httpx.ErrNotFound
		}
		r := insightRight[typ]
		if err := authz.Require(ctx, tx, p.Person, r[0], "Repost"); err != nil {
			return err
		}
		if status != "Published" {
			return httpx.Err(409, "not_published", "Only a published item can be reposted.")
		}
		rs, err := tx.Query(ctx, `select distinct on (c.conversation_id) c.conversation_id, cv.kind, case when cv.kind = 'wall' and cv.audience = 'org' then 'Community wall' when cv.kind = 'wall' then initcap(coalesce(cv.audience, '')) || ' wall' else '' end,
				exists (select 1 from hub.conversation_members m where m.conversation_id = c.conversation_id and m.account_id = $3 and m.left_at is null)
			from hub.comms c join hub.conversations cv on cv.id = c.conversation_id
			where (c.id = $1 or ($2::uuid is not null and c.batch = $2)) and c.status = 'Published' order by c.conversation_id`, id, batch, p.Account)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var x Place
			if err := rs.Scan(&x.Conversation, &x.Kind, &x.Name, &x.Member); err != nil {
				return err
			}
			out = append(out, x)
		}
		return rs.Err()
	})
	return out, err
}

// ---- interaction insights ----

// Spot is how one posted copy of a communication was received in the place it went to.
type Spot struct {
	Comm     uuid.UUID `json:"comm"`
	Place    string    `json:"place"`
	Conv     uuid.UUID `json:"conversation"`
	Kind     string    `json:"kind"` // wall | forum
	Status   string    `json:"status"`
	Audience int       `json:"audience"`
	Read     int       `json:"read"`
	Likes    int       `json:"likes"`
	Replies  int       `json:"replies"`
	Acked    int       `json:"acknowledged"`
	AckedBy  []string  `json:"acknowledged_by"` // who acknowledged it there, earliest first
	// how fast and how fully it landed: the measures WhatsApp can never give
	ReadIn48h  int       `json:"read_in_48h"`      // read within 48 hours of going out
	MedianMins *int      `json:"median_read_mins"` // half of the readers had read it by this many minutes
	Unread     []Pending `json:"unread"`           // who has not read it yet, for a targeted reminder (sender excluded)
	ByDept     []DeptRow `json:"by_department"`    // reach, reads and acknowledgements per department
}

// Pending is one person a communication has not reached yet.
type Pending struct {
	Person     uuid.UUID `json:"person"`
	Name       string    `json:"name"`
	Department string    `json:"department"`
}

// DeptRow is how a communication landed in one department.
type DeptRow struct {
	Department string `json:"department"`
	Audience   int    `json:"audience"`
	Read       int    `json:"read"`
	Acked      int    `json:"acknowledged"`
}

var insightRight = map[string][2]string{"Memo": {"memos", "View insights"}, "Announcement": {"announcements", "View insights"}, "Newsletter": {"newsletter", "View insights"}, "Job post": {"jobposts", "View insights"}}

// Insights counts, for every place a communication (and the copies sent with it) was posted, who read it and how many
// reacted or replied. Only people with the "View insights" right of that communication's page may see it. Reactions and
// replies are messages answering the posted one; their content stays sealed, only how many there are is counted.
func (s *Service) Insights(ctx context.Context, p *tenant.Principal, id uuid.UUID) ([]Spot, error) {
	out := []Spot{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var typ string
		var batch *uuid.UUID
		if err := tx.QueryRow(ctx, `select type, batch from hub.comms where id=$1`, id).Scan(&typ, &batch); err != nil {
			return httpx.ErrNotFound
		}
		r := insightRight[typ]
		if err := authz.Require(ctx, tx, p.Person, r[0], r[1]); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select c.id, case when cv.kind = 'wall' and cv.audience = 'org' then 'Community wall' when cv.kind = 'wall' then initcap(coalesce(cv.audience, '')) || ' wall' else '' end, c.conversation_id, cv.kind, c.status,
				(select count(*) from hub.conversation_members m join hub.people pe on pe.account_id = m.account_id and pe.org_id = m.org_id
					where m.conversation_id = c.conversation_id and m.left_at is null and pe.id <> c.by_person),
				(select count(*) from hub.comm_reads x where x.comm_id = c.id and `+inAudience("x.person_id")+`),
				(select count(*) from hub.messages m where m.org_id = c.org_id and m.conversation_id = c.conversation_id and m.reply_to = c.message_id and m.deleted_at is null and m.content_type = 'reaction'),
				(select count(*) from hub.messages m where m.org_id = c.org_id and m.conversation_id = c.conversation_id and m.reply_to = c.message_id and m.deleted_at is null and m.content_type not in ('reaction','vote','poll_lock'))
			from hub.comms c join hub.conversations cv on cv.id = c.conversation_id
			where c.id = $1 or ($2::uuid is not null and c.batch = $2)
			order by c.created_at, c.id`, id, batch)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var x Spot
			if err := rs.Scan(&x.Comm, &x.Place, &x.Conv, &x.Kind, &x.Status, &x.Audience, &x.Read, &x.Likes, &x.Replies); err != nil {
				return err
			}
			x.AckedBy = []string{}
			out = append(out, x)
		}
		if err := rs.Err(); err != nil {
			return err
		}
		rs.Close()
		for i := range out {
			as, err := tx.Query(ctx, `select pe.first_name || ' ' || pe.last_name from hub.comm_acks a join hub.people pe on pe.id = a.person_id
				join hub.comms c on c.id = a.comm_id
				where a.comm_id = $1 and `+inAudience("a.person_id")+` order by a.at`, out[i].Comm)
			if err != nil {
				return err
			}
			for as.Next() {
				var n string
				if err := as.Scan(&n); err != nil {
					as.Close()
					return err
				}
				out[i].AckedBy = append(out[i].AckedBy, n)
			}
			as.Close()
			out[i].Acked = len(out[i].AckedBy)
			if err := landing(ctx, tx, &out[i]); err != nil {
				return err
			}
		}
		return nil
	})
	return out, err
}

// A communication sent to several places at once is one group (same batch). Every count below is per PERSON across the
// group: someone who is in two of the places, or who met it on the wall and again in a forum, is reached once, has read it
// once (their first read is what counts for speed) and has acknowledged it once. Seeing it again elsewhere is fine and is
// not counted again. Expects the communication as alias c.
const grpComms = `select c2.id from hub.comms c2 where c2.id = c.id or (c.batch is not null and c2.batch = c.batch and c2.status = 'Published')`
const grpAudience = `select distinct ap.id from hub.comms c2
	join hub.conversation_members am on am.conversation_id = c2.conversation_id and am.left_at is null
	join hub.people ap on ap.account_id = am.account_id and ap.org_id = am.org_id
	where (c2.id = c.id or (c.batch is not null and c2.batch = c.batch and c2.status = 'Published')) and ap.id <> c2.by_person`

// inAudience limits a count to the communication's real audience: people who are members of the conversation it was posted
// in now, and not its author. So "read" can never exceed "reached", and someone who left is not counted on either side.
// It expects the communication as alias c.
func inAudience(col string) string {
	return col + ` <> c.by_person and exists (select 1 from hub.conversation_members am join hub.people ap on ap.account_id = am.account_id and ap.org_id = am.org_id
		where am.conversation_id = c.conversation_id and am.left_at is null and ap.id = ` + col + `)`
}

// Totals is how a communication landed for its whole group, one person counted once, whatever places they met it in.
type Totals struct {
	Audience   int       `json:"audience"`
	Read       int       `json:"read"`
	ReadIn48h  int       `json:"read_in_48h"`
	Acked      int       `json:"acknowledged"`
	AckedBy    []string  `json:"acknowledged_by"`
	MedianMins *int      `json:"median_read_mins"`
	Unread     []Pending `json:"unread"`
	ByDept     []DeptRow `json:"by_department"`
}

// GroupTotals measures a communication across all the places it went to, per person (see grpAudience).
func (s *Service) GroupTotals(ctx context.Context, p *tenant.Principal, id uuid.UUID) (*Totals, error) {
	key := "comm:tot:" + p.Org.String() + ":" + id.String() + ":" + s.countsVer(ctx, p.Org)
	var hit Totals
	if s.cacheGet(ctx, key, &hit) {
		return &hit, nil
	}
	t, err := s.groupTotals(ctx, p, id)
	if err == nil {
		s.cachePut(ctx, key, t)
	}
	return t, err
}
