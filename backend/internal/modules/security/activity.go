package security

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/tenant"
)

// Who uses the system: how many different people used it today (each person once, however often they signed in), and
// who has not used it for a week or for two weeks.

type IdlePerson struct {
	Person   uuid.UUID  `json:"person"`
	Name     string     `json:"name"`
	Title    string     `json:"title"`
	LastSeen *time.Time `json:"last_seen,omitempty"` // nil: never used it
}

type Activity struct {
	Day      string       `json:"day"`
	Today    int          `json:"today"`       // people who used the system today
	Active   int          `json:"active"`      // active people in the organization
	Week     []IdlePerson `json:"inactive_1w"` // not seen for 7 to 13 days
	TwoWeeks []IdlePerson `json:"inactive_2w"` // not seen for 14 days or more, or never
	// adoption: different people who used Hub in each month, quarter and year (each person once, however often they signed in),
	// out of the people who had an account by the end of that period
	Monthly   []PeriodUse `json:"monthly"`
	Quarterly []PeriodUse `json:"quarterly"`
	Yearly    []PeriodUse `json:"yearly"`
}

type PeriodUse struct {
	Label  string `json:"label"`
	Start  string `json:"start"`
	Users  int    `json:"users"`  // different people who used it in the period
	People int    `json:"people"` // people with an account by the period's end (not removed)
	Rate   int    `json:"rate"`   // users out of people, in percent
}

var nairobi = func() *time.Location {
	l, err := time.LoadLocation("Africa/Nairobi")
	if err != nil {
		return time.UTC
	}
	return l
}()

func (m *Module) Activity(ctx context.Context, p *tenant.Principal) (*Activity, error) {
	day := time.Now().In(nairobi).Format("2006-01-02")
	out := &Activity{Day: day, Week: []IdlePerson{}, TwoWeeks: []IdlePerson{}}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if ok, _ := authz.Can(ctx, tx, p.Person, "security", "View logs"); !ok {
			if err := authz.Require(ctx, tx, p.Person, "security", "View alerts"); err != nil {
				return err
			}
		}
		if err := tx.QueryRow(ctx, `select (select count(distinct person_id) from hub.person_days where day=$1::date),
			(select count(*) from hub.people where status='active' and account_id is not null)`, day).Scan(&out.Today, &out.Active); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select pe.id, pe.first_name || ' ' || pe.last_name, pe.title, max(d.day)
			from hub.people pe left join hub.person_days d on d.person_id = pe.id
			where pe.status='active' and pe.account_id is not null
			group by pe.id having coalesce(max(d.day), '1970-01-01'::date) <= $1::date - 7
			order by max(d.day) nulls first, 2 limit 1000`, day)
		if err != nil {
			return err
		}
		defer rs.Close()
		today, _ := time.ParseInLocation("2006-01-02", day, nairobi)
		for rs.Next() {
			var x IdlePerson
			var last *time.Time
			if err := rs.Scan(&x.Person, &x.Name, &x.Title, &last); err != nil {
				return err
			}
			x.LastSeen = last
			if last != nil && today.Sub(*last) < 14*24*time.Hour {
				out.Week = append(out.Week, x)
			} else {
				out.TwoWeeks = append(out.TwoWeeks, x)
			}
		}
		if err := rs.Err(); err != nil {
			return err
		}
		rs.Close()
		var err2 error
		if out.Monthly, err2 = periods(ctx, tx, "month", 12, "Mon 2006"); err2 != nil {
			return err2
		}
		if out.Quarterly, err2 = periods(ctx, tx, "quarter", 8, ""); err2 != nil {
			return err2
		}
		out.Yearly, err2 = periods(ctx, tx, "year", 3, "2006")
		return err2
	})
	return out, err
}

// periods counts, for the last n months, quarters or years (newest last), the different people who used Hub and the people who
// could have: one query.
func periods(ctx context.Context, tx pgx.Tx, unit string, n int, layout string) ([]PeriodUse, error) {
	step := map[string]string{"month": "1 month", "quarter": "3 months", "year": "1 year"}[unit]
	rs, err := tx.Query(ctx, `with p as (
			select generate_series(date_trunc($1, (now() at time zone 'Africa/Nairobi')::date) - ($2 - 1) * $3::interval,
				date_trunc($1, (now() at time zone 'Africa/Nairobi')::date), $3::interval)::date as s)
		select p.s, (p.s + $3::interval)::date,
			(select count(distinct d.person_id) from hub.person_days d where d.day >= p.s and d.day < (p.s + $3::interval)::date),
			(select count(*) from hub.people pe where pe.account_id is not null and pe.status <> 'deleted' and pe.created_at < p.s + $3::interval)
		from p order by p.s`, unit, n, step)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	out := []PeriodUse{}
	for rs.Next() {
		var s, e time.Time
		var x PeriodUse
		if err := rs.Scan(&s, &e, &x.Users, &x.People); err != nil {
			return nil, err
		}
		x.Start = s.Format("2006-01-02")
		if layout != "" {
			x.Label = s.Format(layout)
		} else {
			x.Label = "Q" + string(rune('0'+(int(s.Month())-1)/3+1)) + " " + s.Format("2006")
		}
		if x.People > 0 {
			x.Rate = (x.Users*100 + x.People/2) / x.People
		}
		out = append(out, x)
	}
	return out, rs.Err()
}
