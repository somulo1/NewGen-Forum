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

	"hub/internal/access"
	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/db"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/modules/approvals"
	"hub/internal/modules/auth"
	"hub/internal/modules/chat"
	"hub/internal/orgsettings"
	pp "hub/internal/people"
	"hub/internal/realtime"
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
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	First        string    `json:"first_name"`
	Last         string    `json:"last_name"`
	Initials     string    `json:"initials"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	Availability string    `json:"availability"`
	Color        string    `json:"color"`
	Sector       *Ref      `json:"sector,omitempty"`
	Branch       *Ref      `json:"branch,omitempty"`
	Unit         *Ref      `json:"unit,omitempty"`
	Department   *Ref      `json:"department,omitempty"`
	Role         string    `json:"role,omitempty"`
	RoleName     string    `json:"role_name,omitempty"`
	Manager      *Ref      `json:"manager,omitempty"`
	Joined       string    `json:"joined_on,omitempty"`
	HasAccount   bool      `json:"has_account"`
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
	       coalesce(r.key,''), coalesce(r.name,''), mg.id, mg.first_name || ' ' || mg.last_name
	from hub.people p
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
		&x.emailEnc, &x.phoneEnc, &x.dobEnc, &x.genderEnc, &sid, &sn, &bid, &bn, &uid, &un, &did, &dn, &x.Role, &x.RoleName, &mid, &mn)
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
// gender only with the full-profile right. Your own record is always yours.
func (s *Service) reveal(dek []byte, org uuid.UUID, x *scanned, contact, sensitive bool) {
	str := func(col string, ct []byte) *string { v := pp.Dec(dek, org, x.ID, col, ct); return &v }
	if contact {
		x.Email, x.Phone = str(pp.ColEmail, x.emailEnc), str(pp.ColPhone, x.phoneEnc)
	}
	if sensitive {
		x.DOB, x.Gender = str(pp.ColDOB, x.dobEnc), str(pp.ColGender, x.genderEnc)
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
		if err := tx.QueryRow(ctx, "select count(*) from hub.people p left join hub.org_units dx on dx.id=p.department_id left join hub.org_units bx on bx.id=p.branch_id where "+strings.Join(where, " and "), args...).Scan(&out.Total); err != nil {
			return err
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
		if err := authz.Require(ctx, tx, p.Person, "people", "View directory"); err != nil {
			return err
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

// Onboard adds people (one or a whole CSV import) and emails each an invitation. One bad row never blocks the others.
func (s *Service) Onboard(ctx context.Context, p *tenant.Principal, list []NewPerson) ([]Onboarded, error) {
	if len(list) == 0 || len(list) > 500 {
		return nil, httpx.Bad("Send between 1 and 500 people at a time.")
	}
	out := make([]Onboarded, len(list))
	type mailJob struct{ to, link string }
	var jobs []mailJob
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "people", "Onboard people"); err != nil {
			return err
		}
		// the plan's seats bound how many people can be in the organization
		limit, err := orgsettings.SeatLimit(ctx, s.A, tx, p.Org)
		if err != nil {
			return err
		}
		var used int
		if err := tx.QueryRow(ctx, `select count(*) from hub.people where status <> 'deleted'`).Scan(&used); err != nil {
			return err
		}
		if used+len(list) > limit {
			return httpx.Err(402, "seat_limit", fmt.Sprintf("Your plan has %d seats and %d are in use. Add seats in Settings to add more people.", limit, used))
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		roleByKey := map[string]uuid.UUID{}
		rs, err := tx.Query(ctx, `select key, id from hub.roles`)
		if err != nil {
			return err
		}
		for rs.Next() {
			var k string
			var id uuid.UUID
			_ = rs.Scan(&k, &id)
			roleByKey[k] = id
		}
		rs.Close()
		for i := range list {
			n := list[i]
			out[i] = Onboarded{Email: n.Email}
			if err := validate(&n); err != nil {
				out[i].Status, out[i].Error = "rejected", err.(*httpx.Error).Message
				continue
			}
			out[i].Email = n.Email
			role := n.Role
			if role == "" {
				role = "staff"
			}
			rid, ok := roleByKey[role]
			if !ok || role == "superadmin" {
				out[i].Status, out[i].Error = "rejected", "That role cannot be given when adding a person."
				continue
			}
			id := app.NewID()
			if _, err := tx.Exec(ctx, "savepoint row"); err != nil {
				return err
			}
			link, err := s.insertPerson(ctx, tx, dek, p, id, rid, n)
			if err != nil {
				if _, e2 := tx.Exec(ctx, "rollback to savepoint row"); e2 != nil {
					return e2
				}
				out[i].Status = "rejected"
				if strings.Contains(err.Error(), "people_email_uidx") {
					out[i].Error = "Someone with this email is already in the organization."
				} else {
					var he *httpx.Error
					if errors.As(err, &he) {
						out[i].Error = he.Message
					} else {
						return err
					}
				}
				continue
			}
			out[i].ID, out[i].Status = id, "invited"
			jobs = append(jobs, mailJob{n.Email, link})
		}
		var added int
		for _, o := range out {
			if o.Status == "invited" {
				added++
			}
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Added %d people", added), "People", p.IP, map[string]any{"count": added})
	})
	if err != nil {
		return nil, err
	}
	var orgName string
	_ = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select name from hub.orgs where id=$1`, p.Org).Scan(&orgName)
	})
	for _, j := range jobs {
		_ = s.A.Mail.Send(ctx, j.to, "You have been added to "+orgName+" on Hub", "Open this link to join: "+j.link+"\nIt works once and expires in 7 days.")
	}
	return out, nil
}

func (s *Service) insertPerson(ctx context.Context, tx pgx.Tx, dek []byte, p *tenant.Principal, id, role uuid.UUID, n NewPerson) (string, error) {
	enc := func(col, v string) []byte { b, _ := pp.Enc(dek, p.Org, id, col, v); return b }
	if err := s.checkStructure(ctx, tx, n.Sector, n.Branch, n.Unit, n.Department); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `insert into hub.people (id, org_id, first_name, last_name, title, sector_id, branch_id, unit_id, department_id, email_idx, email_enc, phone_enc, dob_enc, gender_enc, color, joined_on)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,coalesce(nullif($15,''),'#7CC4FF'),current_date)`,
		id, p.Org, n.First, n.Last, n.Title, n.Sector, n.Branch, n.Unit, n.Department, pp.EmailIdx(s.A.Cfg.BlindKey, p.Org, n.Email),
		enc(pp.ColEmail, n.Email), enc(pp.ColPhone, n.Phone), enc(pp.ColDOB, n.DOB), enc(pp.ColGender, n.Gender), n.Color); err != nil {
		return "", err
	}
	if err := s.A.Sign.Seal(ctx, tx, "people", "id=$1", id); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `insert into hub.person_roles (org_id, person_id, role_id) values ($1,$2,$3)`, p.Org, id, role); err != nil {
		return "", err
	}
	if err := s.A.Sign.Seal(ctx, tx, "person_roles", "person_id=$1", id); err != nil {
		return "", err
	}
	secret := crypto.RandomToken(32)
	if _, err := tx.Exec(ctx, `insert into hub.person_invites (org_id, person_id, token_hash, expires_at, created_by) values ($1,$2,$3,$4,$5)`,
		p.Org, id, crypto.Hash([]byte(secret)), app.Now().Add(7*24*time.Hour), p.Person); err != nil {
		return "", err
	}
	if err := s.A.Sign.Seal(ctx, tx, "person_invites", "person_id=$1", id); err != nil {
		return "", err
	}
	return strings.TrimRight(s.A.Cfg.AllowedOrigins[0], "/") + "/join?t=" + p.Org.String() + "." + id.String() + "." + secret, nil
}

// checkStructure makes sure the chosen sector/branch/unit/department exist in this organization and fit together.
func (s *Service) checkStructure(ctx context.Context, tx pgx.Tx, ids ...*uuid.UUID) error {
	kinds := []string{"sector", "branch", "unit", "department"}
	for i, id := range ids {
		if id == nil {
			continue
		}
		var k string
		if err := tx.QueryRow(ctx, `select kind from hub.org_units where id=$1`, *id).Scan(&k); err != nil || k != kinds[i] {
			return httpx.Bad("Choose a valid " + kinds[i] + ".")
		}
	}
	return nil
}

// ---- joining ----

// Accept links the signed-in account to the person the invitation was made for.
func (s *Service) Accept(ctx context.Context, account uuid.UUID, token string) (uuid.UUID, error) {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return uuid.Nil, httpx.Bad("That invitation link is not valid.")
	}
	org, e1 := uuid.Parse(parts[0])
	person, e2 := uuid.Parse(parts[1])
	if e1 != nil || e2 != nil {
		return uuid.Nil, httpx.Bad("That invitation link is not valid.")
	}
	bad := httpx.Err(400, "invite_invalid", "That invitation is not valid, has expired or was already used.")
	err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		if err := db.SetOrg(ctx, tx, org); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `update hub.person_invites set used_at=now(), `+integrity.Bump+` where org_id=$1 and person_id=$2 and token_hash=$3 and used_at is null and expires_at > now()`, org, person, crypto.Hash([]byte(parts[2])))
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return bad
		}
		if err := s.A.Sign.Seal(ctx, tx, "person_invites", "person_id=$1", person); err != nil {
			return err
		}
		email, err := s.Auth.AccountEmail(ctx, tx, account)
		if err != nil {
			return err
		}
		tag, err = tx.Exec(ctx, `update hub.people set account_id=$3, updated_at=now(), `+integrity.Bump+` where id=$1 and org_id=$2 and account_id is null and status='active' and email_idx=$4`,
			person, org, account, pp.EmailIdx(s.A.Cfg.BlindKey, org, email))
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return bad // the whole transaction rolls back, so the invitation is not used up by the wrong account
		}
		if err := s.A.Sign.Seal(ctx, tx, "people", "id=$1", person); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.account_orgs (account_id, org_id) values ($1,$2) on conflict do nothing`, account, org); err != nil {
			return err
		}
		if err := chat.SyncAudience(ctx, tx, s.A.Sign, org, account); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, org, &person, "New member", "Joined the organization", "People", "", map[string]any{"person": person})
	})
	if err != nil {
		return uuid.Nil, err
	}
	return org, nil
}

// ---- status actions (the Action card) ----

var statusFor = map[string]string{"activate": "active", "unblock": "active", "deactivate": "inactive", "block": "blocked", "delete": "deleted"}

func (s *Service) SetStatus(ctx context.Context, p *tenant.Principal, id uuid.UUID, action string) error {
	target, ok := statusFor[action]
	if !ok {
		return httpx.Bad("Unknown action.")
	}
	if id == p.Person && action != "activate" {
		return httpx.Err(409, "self_action", "You cannot do that to your own account.")
	}
	var account *uuid.UUID
	var name string
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "people", "Edit profiles"); err != nil {
			return err
		}
		var cur string
		var locked bool
		err := tx.QueryRow(ctx, `select p.status, p.account_id, p.first_name || ' ' || p.last_name, coalesce(r.locked,false) from hub.people p left join hub.person_roles pr on pr.person_id=p.id left join hub.roles r on r.id=pr.role_id where p.id=$1 for update of p`, id).Scan(&cur, &account, &name, &locked)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		if cur == "deleted" {
			return httpx.Err(409, "deleted", "This person was deleted.")
		}
		if action == "unblock" && cur != "blocked" || action == "activate" && cur == "active" || action == "deactivate" && cur != "active" || action == "block" && cur == "blocked" {
			return httpx.Err(409, "no_change", "That person is already "+cur+".")
		}
		if locked {
			mine, err := authz.Can(ctx, tx, p.Person, "roles", "Manage roles & rights")
			if err != nil {
				return err
			}
			var n int
			if err := tx.QueryRow(ctx, `select count(*) from hub.person_roles pr join hub.roles r on r.id=pr.role_id join hub.people x on x.id=pr.person_id where r.locked and x.status='active'`).Scan(&n); err != nil {
				return err
			}
			if !mine || n <= 1 && target != "active" {
				return httpx.Err(403, "protected", "The last super admin and other protected accounts cannot be changed this way.")
			}
		}
		if _, err := tx.Exec(ctx, `update hub.people set status=$2, deleted_at = case when $2='deleted' then now() end, updated_at=now(), `+integrity.Bump+` where id=$1`, id, target); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "people", "id=$1", id); err != nil {
			return err
		}
		if account != nil {
			if target == "active" {
				err = chat.SyncAudience(ctx, tx, s.A.Sign, p.Org, *account)
			} else {
				err = chat.Evict(ctx, tx, s.A.Sign, p.Org, *account)
			}
			if err != nil {
				return err
			}
		}
		verb := map[string]string{"activate": "Activated", "unblock": "Unblocked", "deactivate": "Deactivated", "block": "Blocked", "delete": "Deleted"}[action]
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, verb+" "+name, "People", p.IP, map[string]any{"person": id, "from": cur, "to": target})
	})
	if err != nil {
		return err
	}
	if account != nil && target != "active" {
		s.Auth.RevokeAccountInOrg(ctx, *account, p.Org) // they are signed out of this organization right away
	}
	data, _ := json.Marshal(map[string]any{"person": id, "status": target})
	s.A.Publish(ctx, realtime.Event{Type: "person.status", Channel: realtime.OrgChannel(p.Org), Org: p.Org.String(), Data: data})
	if account != nil {
		s.A.Publish(ctx, realtime.Event{Type: "session.ended", Channel: realtime.AcctChannel(*account), Org: p.Org.String()})
	}
	return nil
}

// ---- admin password reset ----

// ResetPassword gives the person a temporary password (shown once to the administrator) that must be changed at next sign-in.
func (s *Service) ResetPassword(ctx context.Context, p *tenant.Principal, id uuid.UUID) (string, error) {
	var account *uuid.UUID
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "people", "Edit profiles"); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `select account_id from hub.people where id=$1 and status <> 'deleted'`, id).Scan(&account); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		if account == nil {
			return httpx.Err(409, "no_account", "This person has not joined yet. Send the invitation again instead.")
		}
		if id == p.Person {
			return httpx.Err(409, "self_action", "Use Change password in your profile.")
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Reset a password", "People", p.IP, map[string]any{"person": id})
	})
	if err != nil {
		return "", err
	}
	return s.Auth.AdminSetTemporaryPassword(ctx, *account)
}

// ---- edits (go through the approval flow) ----

type Changes struct {
	First      *string    `json:"first_name,omitempty"`
	Last       *string    `json:"last_name,omitempty"`
	Title      *string    `json:"title,omitempty"`
	Phone      *string    `json:"phone,omitempty"`
	DOB        *string    `json:"date_of_birth,omitempty"`
	Gender     *string    `json:"gender,omitempty"`
	Sector     *uuid.UUID `json:"sector_id,omitempty"`
	Branch     *uuid.UUID `json:"branch_id,omitempty"`
	Unit       *uuid.UUID `json:"unit_id,omitempty"`
	Department *uuid.UUID `json:"department_id,omitempty"`
	Role       *string    `json:"role,omitempty"`
	Manager    *uuid.UUID `json:"manager_id,omitempty"`
}

type updatePayload struct {
	Person  uuid.UUID `json:"person"`
	Changes Changes   `json:"changes"`
}

// Update asks to change someone's details, role or rights. Flow f11 decides whether someone must approve first.
func (s *Service) Update(ctx context.Context, p *tenant.Principal, id uuid.UUID, c Changes) (*approvals.Request, bool, error) {
	var req *approvals.Request
	var applied bool
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "people", "Edit profiles"); err != nil {
			return err
		}
		var name string
		if err := tx.QueryRow(ctx, `select first_name || ' ' || last_name from hub.people where id=$1 and status <> 'deleted'`, id).Scan(&name); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		if c.Role != nil {
			if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
				return err
			}
			if *c.Role == "superadmin" {
				ok, _ := authz.Can(ctx, tx, p.Person, "roles", "Manage roles & rights")
				var mine bool
				_ = tx.QueryRow(ctx, `select coalesce(r.locked,false) from hub.person_roles pr join hub.roles r on r.id=pr.role_id where pr.person_id=$1`, p.Person).Scan(&mine)
				if !ok || !mine {
					return httpx.ErrForbidden
				}
			}
		}
		if err := s.validateChanges(ctx, tx, c); err != nil {
			return err
		}
		// protected (locked-role) people can only be changed by someone holding a protected role themselves
		var targetLocked, mineLocked bool
		_ = tx.QueryRow(ctx, `select coalesce(r.locked,false) from hub.person_roles pr join hub.roles r on r.id=pr.role_id where pr.person_id=$1`, id).Scan(&targetLocked)
		_ = tx.QueryRow(ctx, `select coalesce(r.locked,false) from hub.person_roles pr join hub.roles r on r.id=pr.role_id where pr.person_id=$1`, p.Person).Scan(&mineLocked)
		if targetLocked && !mineLocked {
			return httpx.Err(403, "protected", "Only a super admin can change another super admin.")
		}
		var err error
		req, applied, err = s.Ap.Submit(ctx, tx, p, "f11", "person.update", "Change details for "+name, "", updatePayload{id, c})
		return err
	})
	return req, applied, err
}

func (s *Service) validateChanges(ctx context.Context, tx pgx.Tx, c Changes) error {
	for _, v := range []*string{c.First, c.Last} {
		if v != nil && (strings.TrimSpace(*v) == "" || len(*v) > 60) {
			return httpx.Bad("Names cannot be empty.")
		}
	}
	if c.DOB != nil {
		if _, err := time.Parse("2006-01-02", *c.DOB); err != nil {
			return httpx.Bad("Enter the date of birth as YYYY-MM-DD.")
		}
	}
	if c.Gender != nil && !genders[*c.Gender] {
		return httpx.Bad("Choose a gender from the list.")
	}
	if c.Phone != nil && (len(*c.Phone) < 5 || len(*c.Phone) > 24) {
		return httpx.Bad("Enter a valid phone number.")
	}
	if c.Role != nil {
		if _, ok := access.Default().Defaults[*c.Role]; !ok {
			return httpx.Bad("Unknown role.")
		}
	}
	return s.checkStructure(ctx, tx, c.Sector, c.Branch, c.Unit, c.Department)
}

// applyUpdate runs when f11 is satisfied: it re-validates (the world may have changed while it waited) and writes.
func (s *Service) applyUpdate(ctx context.Context, tx pgx.Tx, a approvals.Applied) error {
	var pl updatePayload
	if err := json.Unmarshal(a.Payload, &pl); err != nil {
		return err
	}
	c := pl.Changes
	if err := s.validateChanges(ctx, tx, c); err != nil {
		return err
	}
	enc := func(col string, v *string) any {
		if v == nil {
			return nil
		}
		b, _ := pp.Enc(a.DEK, a.Org, pl.Person, col, *v)
		return b
	}
	tag, err := tx.Exec(ctx, `update hub.people set
		first_name = coalesce($2, first_name), last_name = coalesce($3, last_name), title = coalesce($4, title),
		phone_enc = coalesce($5, phone_enc), dob_enc = coalesce($6, dob_enc), gender_enc = coalesce($7, gender_enc),
		sector_id = coalesce($8, sector_id), branch_id = coalesce($9, branch_id), unit_id = coalesce($10, unit_id), department_id = coalesce($11, department_id),
		updated_at = now(), `+integrity.Bump+` where id=$1 and status <> 'deleted'`,
		pl.Person, c.First, c.Last, c.Title, enc(pp.ColPhone, c.Phone), enc(pp.ColDOB, c.DOB), enc(pp.ColGender, c.Gender), c.Sector, c.Branch, c.Unit, c.Department)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httpx.ErrNotFound
	}
	if err := s.A.Sign.Seal(ctx, tx, "people", "id=$1", pl.Person); err != nil {
		return err
	}
	var acct *uuid.UUID
	if err := tx.QueryRow(ctx, `select account_id from hub.people where id=$1`, pl.Person).Scan(&acct); err == nil && acct != nil && (c.Sector != nil || c.Branch != nil || c.Unit != nil || c.Department != nil) {
		if err := chat.SyncAudience(ctx, tx, s.A.Sign, a.Org, *acct); err != nil {
			return err
		}
	}
	if c.Role != nil {
		if _, err := tx.Exec(ctx, `update hub.person_roles set role_id=(select id from hub.roles where key=$2), `+integrity.Bump+` where person_id=$1`, pl.Person, *c.Role); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "person_roles", "person_id=$1", pl.Person); err != nil {
			return err
		}
	}
	if c.Manager != nil {
		if *c.Manager == pl.Person {
			return httpx.Bad("Nobody reports to themselves.")
		}
		if _, err := tx.Exec(ctx, `insert into hub.person_reports (org_id, person_id, manager_id) values ($1,$2,$3) on conflict (org_id, person_id) do update set manager_id=excluded.manager_id, `+integrity.BumpOn("person_reports"), a.Org, pl.Person, *c.Manager); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "person_reports", "person_id=$1", pl.Person); err != nil {
			return err
		}
	}
	data, _ := json.Marshal(map[string]any{"person": pl.Person})
	go s.A.Publish(context.Background(), realtime.Event{Type: "person.updated", Channel: realtime.OrgChannel(a.Org), Org: a.Org.String(), Data: data})
	return nil
}
