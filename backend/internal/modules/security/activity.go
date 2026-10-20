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
		return rs.Err()
	})
	return out, err
}
