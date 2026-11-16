// Package people is the directory: who works in the organization, their profiles, onboarding and account status.
package people

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/modules/approvals"
	"hub/internal/modules/auth"
	pp "hub/internal/people"
	"hub/internal/tenant"
)

type Service struct {
	A    *app.App
	Auth *auth.Service
	Ap   *approvals.Engine
}

func New(a *app.App, au *auth.Service, ap *approvals.Engine) *Service {
	s := &Service{A: a, Auth: au, Ap: ap}
	ap.Register("person.update", s.applyUpdate)
	return s
}

type Ref struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Person struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	First        string     `json:"first_name"`
	Last         string     `json:"last_name"`
	Initials     string     `json:"initials"`
	Title        string     `json:"title"`
	Headline     string     `json:"headline"` // what the person wrote about themselves on their profile ("Software developer | DevOps")
	Status       string     `json:"status"`
	Availability string     `json:"availability"`
	Color        string     `json:"color"`
	Sector       *Ref       `json:"sector,omitempty"`
	Branch       *Ref       `json:"branch,omitempty"`
	Unit         *Ref       `json:"unit,omitempty"`
	Department   *Ref       `json:"department,omitempty"`
	Role         string     `json:"role,omitempty"`
	RoleName     string     `json:"role_name,omitempty"`
	Manager      *Ref       `json:"manager,omitempty"`
	ReportsTo    *ReportsTo `json:"reports_to,omitempty"` // single-person reads only
	Joined       string     `json:"joined_on,omitempty"`
	HasAccount   bool       `json:"has_account"`
	HasPhoto     bool       `json:"has_photo"`           // a profile photo is set; the picture itself is at GET /people/{id}/photo
	PhotoVer     int64      `json:"photo_ver,omitempty"` // changes when the profile changes, so a browser knows to fetch the picture again
	// personal data, present only when the viewer may see contact details (or it is their own profile)
	Email  *string `json:"email,omitempty"`
	Phone  *string `json:"phone,omitempty"`
	DOB    *string `json:"date_of_birth,omitempty"`
	Gender *string `json:"gender,omitempty"`
	// profile (full profile right)
	Bio    *string  `json:"bio,omitempty"`
	Skills []string `json:"skills,omitempty"`
	Cover  string   `json:"cover,omitempty"`
}

func initials(first, last string) string {
	s := ""
	for _, n := range []string{first, last} {
		if r := []rune(strings.TrimSpace(n)); len(r) > 0 {
			s += strings.ToUpper(string(r[0]))
		}
	}
	return s
}

const personSelect = `
	select p.id, p.first_name, p.last_name, p.title, p.status, p.availability, p.color, coalesce(p.joined_on::text,''), p.account_id is not null,
	       p.email_enc, p.phone_enc, p.dob_enc, p.gender_enc,
	       sx.id, sx.name, bx.id, bx.name, ux.id, ux.name, dx.id, dx.name,
	       coalesce(r.key,''), coalesce(r.name,''), mg.id, mg.first_name || ' ' || mg.last_name, coalesce(ph.headline,''),
	       pp.photo_enc is not null, coalesce(extract(epoch from pp.updated_at)::bigint, 0)
	from hub.people p
	left join hub.person_headlines ph on ph.person_id = p.id
	left join hub.person_profiles pp on pp.person_id = p.id
	left join hub.org_units sx on sx.id = p.sector_id left join hub.org_units bx on bx.id = p.branch_id
	left join hub.org_units ux on ux.id = p.unit_id   left join hub.org_units dx on dx.id = p.department_id
	left join hub.person_roles pr on pr.person_id = p.id left join hub.roles r on r.id = pr.role_id
	left join hub.person_reports prr on prr.person_id = p.id left join hub.people mg on mg.id = prr.manager_id`

type scanned struct {
	Person
	emailEnc, phoneEnc, dobEnc, genderEnc []byte
}

func scanPerson(r pgx.Row) (*scanned, error) {
	x := &scanned{}
	var sid, bid, uid, did, mid *uuid.UUID
	var sn, bn, un, dn, mn *string
	err := r.Scan(&x.ID, &x.First, &x.Last, &x.Title, &x.Status, &x.Availability, &x.Color, &x.Joined, &x.HasAccount,
		&x.emailEnc, &x.phoneEnc, &x.dobEnc, &x.genderEnc, &sid, &sn, &bid, &bn, &uid, &un, &did, &dn, &x.Role, &x.RoleName, &mid, &mn, &x.Headline, &x.HasPhoto, &x.PhotoVer)
	if err != nil {
		return nil, err
	}
	x.Name = strings.TrimSpace(x.First + " " + x.Last)
	x.Initials = initials(x.First, x.Last)
	mk := func(id *uuid.UUID, n *string) *Ref {
		if id == nil || n == nil {
			return nil
		}
		return &Ref{*id, *n}
	}
	x.Sector, x.Branch, x.Unit, x.Department, x.Manager = mk(sid, sn), mk(bid, bn), mk(uid, un), mk(did, dn), mk(mid, mn)
	return x, nil
}

// reveal decrypts what the viewer may see: contact details (email, phone) with the contact right, and date of birth and
// gender only with the full-profile right. Your own record is always yours. All of it leaves the API masked; storage stays clean.
func (s *Service) reveal(dek []byte, org uuid.UUID, x *scanned, contact, sensitive bool) {
	str := func(col string, ct []byte, mask func(string) string) *string {
		v := mask(pp.Dec(dek, org, x.ID, col, ct))
		return &v
	}
	if contact {
		x.Email, x.Phone = str(pp.ColEmail, x.emailEnc, pp.MaskEmail), str(pp.ColPhone, x.phoneEnc, pp.MaskPhone)
	}
	if sensitive {
		x.DOB, x.Gender = str(pp.ColDOB, x.dobEnc, pp.MaskDOB), str(pp.ColGender, x.genderEnc, pp.MaskGender)
	}
}

type ListQuery struct {
	Q, Dept, Branch, Status string
	AvailableOnly           bool
	Cursor                  string
	Limit                   int
}

type Page struct {
	People []*Person `json:"people"`
	Next   string    `json:"next,omitempty"`
	Total  int       `json:"total"`
}

// List pages through the directory with keyset pagination, so page 5,000 costs the same as page 1.
func (s *Service) List(ctx context.Context, p *tenant.Principal, q ListQuery) (*Page, error) {
	if q.Limit <= 0 || q.Limit > 200 {
		q.Limit = 50
	}
	var out Page
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "people", "View directory"); err != nil {
			return err
		}
		contact, err := authz.Can(ctx, tx, p.Person, "people", "See contact details")
		if err != nil {
			return err
		}
		full, err := authz.Can(ctx, tx, p.Person, "people", "View full profiles")
		if err != nil {
			return err
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		where := []string{"p.status <> 'deleted'"}
		var args []any
		add := func(cond string, v any) { args = append(args, v); where = append(where, fmt.Sprintf(cond, len(args))) }
		if q.Status != "" {
			add("p.status = $%d", q.Status)
		}
		if q.AvailableOnly {
			where = append(where, "p.availability = 'Available' and p.status = 'active'")
		}
		if q.Dept != "" {
			add("dx.name = $%d", q.Dept)
		}
		if q.Branch != "" {
			add("bx.name = $%d", q.Branch)
		}
		if t := strings.TrimSpace(q.Q); t != "" {
			add("(lower(p.first_name || ' ' || p.last_name) like $%d or lower(p.title) like $"+fmt.Sprint(len(args)+1)+")", "%"+strings.ToLower(escapeLike(t))+"%")
		}
		// the total is counted once, on the first page; later pages (cursor) do not recount the whole directory
		if q.Cursor == "" {
			if err := tx.QueryRow(ctx, "select count(*) from hub.people p left join hub.org_units dx on dx.id=p.department_id left join hub.org_units bx on bx.id=p.branch_id where "+strings.Join(where, " and "), args...).Scan(&out.Total); err != nil {
				return err
			}
		} else {
			out.Total = -1
		}
		if q.Cursor != "" {
			f, l, id, ok := decodeCursor(q.Cursor)
			if !ok {
				return httpx.Bad("That page marker is not valid.")
			}
			args = append(args, f, l, id)
			n := len(args)
			where = append(where, fmt.Sprintf("(p.first_name, p.last_name, p.id) > ($%d, $%d, $%d)", n-2, n-1, n))
		}
		rs, err := tx.Query(ctx, personSelect+" where "+strings.Join(where, " and ")+fmt.Sprintf(" order by p.first_name, p.last_name, p.id limit %d", q.Limit+1), args...)
		if err != nil {
			return err
		}
		defer rs.Close()
		var rows []*scanned
		for rs.Next() {
			x, err := scanPerson(rs)
			if err != nil {
				return err
			}
			rows = append(rows, x)
		}
		if err := rs.Err(); err != nil {
			return err
		}
		if len(rows) > q.Limit {
			last := rows[q.Limit-1]
			out.Next = encodeCursor(last.First, last.Last, last.ID)
			rows = rows[:q.Limit]
		}
		for _, x := range rows {
			self := x.ID == p.Person
			s.reveal(dek, p.Org, x, contact || self, full || self)
			out.People = append(out.People, &x.Person)
		}
		return nil
	})
	return &out, err
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func encodeCursor(f, l string, id uuid.UUID) string {
	b, _ := json.Marshal([]string{f, l, id.String()})
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(c string) (f, l string, id uuid.UUID, ok bool) {
	b, err := base64.RawURLEncoding.DecodeString(c)
	if err != nil {
		return
	}
	var v []string
	if json.Unmarshal(b, &v) != nil || len(v) != 3 {
		return
	}
	id, err = uuid.Parse(v[2])
	return v[0], v[1], id, err == nil
}

// Get returns one person. Contact details need the right (or it is you); the biography and skills need "View full profiles".
func (s *Service) Get(ctx context.Context, p *tenant.Principal, id uuid.UUID) (*Person, error) {
	var out *Person
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		// anyone may read their own record; others need the right to see the directory
		if id != p.Person {
			if err := authz.Require(ctx, tx, p.Person, "people", "View directory"); err != nil {
				return err
			}
		}
		x, err := scanPerson(tx.QueryRow(ctx, personSelect+" where p.id=$1 and p.status <> 'deleted'", id))
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		self := id == p.Person
		contact, err := authz.Can(ctx, tx, p.Person, "people", "See contact details")
		if err != nil {
			return err
		}
		full, err := authz.Can(ctx, tx, p.Person, "people", "View full profiles")
		if err != nil {
			return err
		}
		s.reveal(dek, p.Org, x, contact || self, full || self)
		if x.ReportsTo, err = s.reportsTo(ctx, tx, x.ID, x.Manager); err != nil {
			return err
		}
		if full || self {
			if err := s.loadProfile(ctx, tx, dek, p.Org, &x.Person); err != nil {
				return err
			}
		}
		out = &x.Person
		return nil
	})
	return out, err
}

func profileCtx(org, person uuid.UUID, col string) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "person_profiles", Column: col, RowID: person.String()}
}

func (s *Service) loadProfile(ctx context.Context, tx pgx.Tx, dek []byte, org uuid.UUID, x *Person) error {
	var bio, skills []byte
	var cover string
	err := tx.QueryRow(ctx, `select bio_enc, skills_enc, cover from hub.person_profiles where person_id=$1`, x.ID).Scan(&bio, &skills, &cover)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	x.Cover = cover
	if len(bio) > 0 {
		if b, err := crypto.FieldOpen(dek, profileCtx(org, x.ID, "bio_enc"), bio); err == nil {
			v := string(b)
			x.Bio = &v
		}
	}
	if len(skills) > 0 {
		if b, err := crypto.FieldOpen(dek, profileCtx(org, x.ID, "skills_enc"), skills); err == nil {
			_ = json.Unmarshal(b, &x.Skills)
		}
	}
	return nil
}

// ---- onboarding ----

type NewPerson struct {
	First      string     `json:"first_name"`
	Last       string     `json:"last_name"`
	Email      string     `json:"email"`
	Phone      string     `json:"phone"`
	DOB        string     `json:"date_of_birth"`
	Gender     string     `json:"gender"`
	Title      string     `json:"title"`
	Sector     *uuid.UUID `json:"sector_id"`
	Branch     *uuid.UUID `json:"branch_id"`
	Unit       *uuid.UUID `json:"unit_id"`
	Department *uuid.UUID `json:"department_id"`
	Role       string     `json:"role"`
	Color      string     `json:"color"`
	// InviteLink sends the emailed join link instead of a ready account with a first password (the person makes their own account and
	// accepts the link). Without it, the person gets an active account and their first password by email.
	InviteLink bool `json:"invite_link"`
}

type Onboarded struct {
	ID     uuid.UUID `json:"id"`
	Email  string    `json:"email"`
	Status string    `json:"status"`
	Error  string    `json:"error,omitempty"`
}

var genders = map[string]bool{"Female": true, "Male": true, "Other": true, "Prefer not to say": true}

func validate(n *NewPerson) error {
	if n.Color != "" && (len(n.Color) != 7 || n.Color[0] != '#') {
		return httpx.Bad("Choose a colour like #7CC4FF.")
	}
	n.First, n.Last, n.Title, n.Phone = strings.TrimSpace(n.First), strings.TrimSpace(n.Last), strings.TrimSpace(n.Title), strings.TrimSpace(n.Phone)
	if n.First == "" || n.Last == "" || len(n.First) > 60 || len(n.Last) > 60 {
		return httpx.Bad("First name and second name are required.")
	}
	e, err := auth.NormEmail(n.Email)
	if err != nil {
		return err
	}
	n.Email = e
	if _, err := time.Parse("2006-01-02", n.DOB); err != nil {
		return httpx.Bad("Enter the date of birth as YYYY-MM-DD.")
	}
	if !genders[n.Gender] {
		return httpx.Bad("Choose a gender from the list.")
	}
	if len(n.Phone) < 5 || len(n.Phone) > 24 {
		return httpx.Bad("Enter a valid phone number.")
	}
	return nil
}
