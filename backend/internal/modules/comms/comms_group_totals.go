package comms

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/notify"
	"hub/internal/tenant"
)

func (s *Service) groupTotals(ctx context.Context, p *tenant.Principal, id uuid.UUID) (*Totals, error) {
	t := &Totals{AckedBy: []string{}, Unread: []Pending{}, ByDept: []DeptRow{}}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		cte := `with c as (select * from hub.comms where id = $1),
			aud as (` + strings.Replace(grpAudience, "from hub.comms c2", "from c, hub.comms c2", 1) + `),
			t0 as (select min(coalesce(c2.published_at, c2.created_at)) at from hub.comms c2, c where c2.id in (` + grpComms + `)),
			fr as (select r.person_id pid, min(r.at) at from hub.comm_reads r, c where r.comm_id in (` + grpComms + `) and r.person_id in (select * from aud) group by r.person_id),
			fa as (select a.person_id pid, min(a.at) at from hub.comm_acks a, c where a.comm_id in (` + grpComms + `) and a.person_id in (select * from aud) group by a.person_id)`
		var med *float64
		if err := tx.QueryRow(ctx, cte+` select (select count(*) from aud), (select count(*) from fr),
				(select count(*) from fr, t0 where fr.at <= t0.at + interval '48 hours'), (select count(*) from fa),
				(select percentile_cont(0.5) within group (order by extract(epoch from (fr.at - t0.at)) / 60) from fr, t0)`, id).Scan(&t.Audience, &t.Read, &t.ReadIn48h, &t.Acked, &med); err != nil {
			return err
		}
		if med != nil {
			m := int(*med + 0.5)
			if m < 0 {
				m = 0
			}
			t.MedianMins = &m
		}
		rs, err := tx.Query(ctx, cte+` select pe.id, pe.first_name || ' ' || pe.last_name, coalesce(u.name, 'No department'),
				exists (select 1 from fr where fr.pid = pe.id), (select at from fa where fa.pid = pe.id)
			from hub.people pe left join hub.org_units u on u.id = pe.department_id
			where pe.id in (select * from aud) order by 3, 2 limit 5000`, id)
		if err != nil {
			return err
		}
		defer rs.Close()
		at := map[string]int{}
		type ack struct {
			name string
			at   time.Time
		}
		acks := []ack{}
		for rs.Next() {
			var pid uuid.UUID
			var name, dept string
			var read bool
			var ackAt *time.Time
			if err := rs.Scan(&pid, &name, &dept, &read, &ackAt); err != nil {
				return err
			}
			i, ok := at[dept]
			if !ok {
				i = len(t.ByDept)
				at[dept] = i
				t.ByDept = append(t.ByDept, DeptRow{Department: dept})
			}
			d := &t.ByDept[i]
			d.Audience++
			if read {
				d.Read++
			} else if len(t.Unread) < 300 {
				t.Unread = append(t.Unread, Pending{Person: pid, Name: name, Department: dept})
			}
			if ackAt != nil {
				d.Acked++
				acks = append(acks, ack{name, *ackAt})
			}
		}
		sort.Slice(acks, func(i, j int) bool { return acks[i].at.Before(acks[j].at) })
		for _, a := range acks {
			t.AckedBy = append(t.AckedBy, a.name)
		}
		return rs.Err()
	})
	return t, err
}

// landing fills in the speed of reading, who has not read it yet, and the per-department picture for one place.
func landing(ctx context.Context, tx pgx.Tx, x *Spot) error {
	var med *float64
	if err := tx.QueryRow(ctx, `select count(r.person_id) filter (where r.at <= coalesce(c.published_at, c.created_at) + interval '48 hours'),
			percentile_cont(0.5) within group (order by extract(epoch from (r.at - coalesce(c.published_at, c.created_at))) / 60)
		from hub.comms c left join hub.comm_reads r on r.comm_id = c.id and `+inAudience("r.person_id")+` where c.id = $1`, x.Comm).Scan(&x.ReadIn48h, &med); err != nil {
		return err
	}
	if med != nil {
		m := int(*med + 0.5)
		if m < 0 {
			m = 0
		}
		x.MedianMins = &m
	}
	rs, err := tx.Query(ctx, `select p.id, p.first_name || ' ' || p.last_name, coalesce(u.name, 'No department'),
			exists (select 1 from hub.comm_reads r where r.comm_id = $2 and r.person_id = p.id),
			exists (select 1 from hub.comm_acks a where a.comm_id = $2 and a.person_id = p.id),
			p.id = (select by_person from hub.comms where id = $2)
		from hub.conversation_members m
		join hub.people p on p.account_id = m.account_id and p.org_id = m.org_id
		left join hub.org_units u on u.id = p.department_id
		where m.conversation_id = $1 and m.left_at is null
		order by 3, 2 limit 5000`, x.Conv, x.Comm)
	if err != nil {
		return err
	}
	defer rs.Close()
	x.Unread, x.ByDept = []Pending{}, []DeptRow{}
	at := map[string]int{}
	for rs.Next() {
		var pid uuid.UUID
		var name, dept string
		var read, acked, sender bool
		if err := rs.Scan(&pid, &name, &dept, &read, &acked, &sender); err != nil {
			return err
		}
		if sender {
			continue
		}
		i, ok := at[dept]
		if !ok {
			i = len(x.ByDept)
			at[dept] = i
			x.ByDept = append(x.ByDept, DeptRow{Department: dept})
		}
		d := &x.ByDept[i]
		d.Audience++
		if read {
			d.Read++
		} else if len(x.Unread) < 300 {
			x.Unread = append(x.Unread, Pending{Person: pid, Name: name, Department: dept})
		}
		if acked {
			d.Acked++
		}
	}
	return rs.Err()
}

// EngRow is how one published communication (one place) landed, for the engagement overview.
type EngRow struct {
	Comm       uuid.UUID  `json:"comm"`
	Batch      *uuid.UUID `json:"batch,omitempty"` // the group: one row per communication, however many places it went to
	Type       string     `json:"type"`
	Ref        string     `json:"ref"`
	At         time.Time  `json:"at"`
	Audience   int        `json:"audience"`
	Read       int        `json:"read"`
	ReadIn48h  int        `json:"read_in_48h"`
	Acked      int        `json:"acknowledged"`
	MedianMins *int       `json:"median_read_mins"`
}

// engCTE: the latest published communications (heads, copies sent together counted once), who each reached (aud), who read
// (fr) and who acknowledged (fa). $1 is the types the caller may see.
const engCTE = `with cc as (
				select id, coalesce(batch, id) k, conversation_id, by_person, coalesce(published_at, created_at) t0, type, ref
				from hub.comms where status = 'Published' and type = any($1)),
			heads as (select k, min(t0) t0, (array_agg(id order by t0, id))[1] comm, (array_agg(type order by t0, id))[1] type, (array_agg(ref order by t0, id))[1] ref
				from cc group by k order by min(t0) desc limit 300),
			aud as (select distinct cc.k, pe.id pid from cc join heads h on h.k = cc.k
				join hub.conversation_members m on m.conversation_id = cc.conversation_id and m.left_at is null
				join hub.people pe on pe.account_id = m.account_id and pe.org_id = m.org_id where pe.id <> cc.by_person),
			fr as (select cc.k, r.person_id pid, min(r.at) at from cc join heads h on h.k = cc.k join hub.comm_reads r on r.comm_id = cc.id
				join aud on aud.k = cc.k and aud.pid = r.person_id group by cc.k, r.person_id),
			fa as (select distinct cc.k, a.person_id pid from cc join heads h on h.k = cc.k join hub.comm_acks a on a.comm_id = cc.id
				join aud on aud.k = cc.k and aud.pid = a.person_id)`

// People counts distinct people across every communication measured: each person once, however many they received.
type People struct {
	Reached int `json:"reached"`
	Behind  int `json:"behind"` // people with at least one communication still unread
	// every person-and-communication pair measured (a person who got 3 communications is 3 deliveries), so the rates below are
	// exact counts, not averages of percentages
	Deliveries int  `json:"deliveries"`
	Read       int  `json:"read"`
	ReadIn48h  int  `json:"read_in_48h"`
	Acked      int  `json:"acknowledged"`
	MedianMins *int `json:"median_read_mins"` // over every first read, not a median of medians
}

// Engagement measures the latest published communications (up to 300) of the types the person may see insights for,
// in one query, so the overview does not ask the server once per communication.
func (s *Service) Engagement(ctx context.Context, p *tenant.Principal) ([]EngRow, *People, error) {
	out := []EngRow{}
	ppl := &People{}
	engKey := ""
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		types := []string{}
		for t, r := range insightRight {
			if ok, err := authz.Can(ctx, tx, p.Person, r[0], r[1]); err == nil && ok {
				types = append(types, t)
			}
		}
		if len(types) == 0 {
			return httpx.ErrForbidden
		}
		sort.Strings(types)
		key := "comm:eng:" + p.Org.String() + ":" + strings.Join(types, ",") + ":" + s.countsVer(ctx, p.Org)
		var hit struct {
			Rows   []EngRow `json:"rows"`
			People *People  `json:"people"`
		}
		if s.cacheGet(ctx, key, &hit) {
			out, ppl = hit.Rows, hit.People
			return errCached
		}
		engKey = key
		rs, err := tx.Query(ctx, engCTE+`
			select h.comm, h.k, h.type, h.ref, h.t0,
				(select count(*) from aud where aud.k = h.k), (select count(*) from fr where fr.k = h.k),
				(select count(*) from fr where fr.k = h.k and fr.at <= h.t0 + interval '48 hours'), (select count(*) from fa where fa.k = h.k),
				(select percentile_cont(0.5) within group (order by extract(epoch from (fr.at - h.t0)) / 60) from fr where fr.k = h.k)
			from heads h order by h.t0 desc`, types)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var e EngRow
			var med *float64
			if err := rs.Scan(&e.Comm, &e.Batch, &e.Type, &e.Ref, &e.At, &e.Audience, &e.Read, &e.ReadIn48h, &e.Acked, &med); err != nil {
				return err
			}
			if med != nil {
				m := int(*med + 0.5)
				if m < 0 {
					m = 0
				}
				e.MedianMins = &m
			}
			out = append(out, e)
		}
		rs.Close()
		if err := rs.Err(); err != nil {
			return err
		}
		return tx.QueryRow(ctx, engCTE+`
			select count(distinct aud.pid), count(distinct aud.pid) filter (where not exists (select 1 from fr where fr.k = aud.k and fr.pid = aud.pid)),
				(select count(*) from aud), (select count(*) from fr), (select count(*) from fr join heads h on h.k = fr.k where fr.at <= h.t0 + interval '48 hours'),
				(select count(*) from fa), (select round(percentile_cont(0.5) within group (order by extract(epoch from (fr.at - h.t0)) / 60))::int from fr join heads h on h.k = fr.k)
			from aud`, types).Scan(&ppl.Reached, &ppl.Behind, &ppl.Deliveries, &ppl.Read, &ppl.ReadIn48h, &ppl.Acked, &ppl.MedianMins)
	})
	if errors.Is(err, errCached) {
		return out, ppl, nil
	}
	if err == nil && engKey != "" {
		s.cachePut(ctx, engKey, struct {
			Rows   []EngRow `json:"rows"`
			People *People  `json:"people"`
		}{out, ppl})
	}
	return out, ppl, err
}

var errCached = errors.New("cached")

// Remind sends a reminder only to the people a communication has not reached yet (in every place it went), instead of
// re-posting it to everyone. At most once every 6 hours per communication, so nobody is nagged.
func (s *Service) Remind(ctx context.Context, p *tenant.Principal, id uuid.UUID) (int, error) {
	if _, err := s.Insights(ctx, p, id); err != nil { // the right to see insights, before anything is recorded
		return 0, err
	}
	t, err := s.GroupTotals(ctx, p, id)
	if err != nil {
		return 0, err
	}
	if s.A.Rdb != nil {
		ok, err := s.A.Rdb.SetNX(ctx, "comm:remind:"+p.Org.String()+":"+id.String(), "1", 6*time.Hour).Result()
		if err == nil && !ok {
			return 0, httpx.Err(429, "reminded_recently", "A reminder for this was sent in the last 6 hours. Try again later.")
		}
	}
	var typ, ref string
	var nb notify.Batch
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `select type, ref from hub.comms where id=$1`, id).Scan(&typ, &ref); err != nil {
			return httpx.ErrNotFound
		}
		for _, u := range t.Unread { // one reminder per person, only to those who have read it in no place at all
			if err := nb.Add(ctx, tx, p.Org, u.Person, "comm", "Reminder: please read "+strings.ToLower(typ)+" "+ref+".", "#FFB86B", ref); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	nb.Send(ctx, s.A)
	return len(t.Unread), nil
}

// Insight caches: the heavy group counts (GroupTotals, Engagement) are kept for a minute, and dropped at once when anyone reads,
// acknowledges or a communication is published in the organization (the org's counts version moves on).
func (s *Service) countsVer(ctx context.Context, org uuid.UUID) string {
	v, err := s.A.Rdb.Get(ctx, "comm:ver:"+org.String()).Result()
	if err != nil {
		return "0"
	}
	return v
}

func (s *Service) bumpCounts(ctx context.Context, org uuid.UUID) {
	_ = s.A.Rdb.Incr(ctx, "comm:ver:"+org.String()).Err()
}

func (s *Service) cacheGet(ctx context.Context, key string, into any) bool {
	b, err := s.A.Rdb.Get(ctx, key).Bytes()
	return err == nil && json.Unmarshal(b, into) == nil
}

func (s *Service) cachePut(ctx context.Context, key string, v any) {
	if b, err := json.Marshal(v); err == nil {
		_ = s.A.Rdb.Set(ctx, key, b, time.Minute).Err()
	}
}
