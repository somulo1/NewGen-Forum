package people

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/db"
	"hub/internal/httpx"
	"hub/internal/integrity"
	hubmail "hub/internal/mail"
	"hub/internal/modules/approvals"
	"hub/internal/modules/auth"
	"hub/internal/modules/chat"
	pp "hub/internal/people"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

// Update asks to change someone's details, role or rights. Flow f11 decides whether someone must approve first.
func (s *Service) Update(ctx context.Context, p *tenant.Principal, id uuid.UUID, c Changes) (*approvals.Request, bool, error) {
	var req *approvals.Request
	var applied bool
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		// moving someone to another role is a roles-and-rights matter; any other change edits their profile
		if !c.roleOnly() {
			if err := authz.Require(ctx, tx, p.Person, "people", "Edit profiles"); err != nil {
				return err
			}
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
			var givesLocked bool
			_ = tx.QueryRow(ctx, `select locked from hub.roles where key=$1`, *c.Role).Scan(&givesLocked)
			if givesLocked {
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
		if err := keepOwner(ctx, tx, id, c.Role); err != nil {
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
	// a masked value sent back must never replace the real one
	for _, v := range []*string{c.Email, c.Phone, c.DOB, c.Gender} {
		if v != nil && pp.IsMasked(*v) {
			return httpx.Bad("Type the full new value, or leave the hidden one as it is.")
		}
	}
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
	if c.Email != nil {
		if _, err := auth.NormEmail(*c.Email); err != nil {
			return err
		}
	}
	if c.Role != nil && *c.Role != "" { // an empty role takes the person's role away
		var ok bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.roles where key=$1)`, *c.Role).Scan(&ok); err != nil {
			return err
		}
		if !ok {
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
	if c.Email != nil {
		if err := s.changeEmail(ctx, tx, a, pl.Person, *c.Email); err != nil {
			return err
		}
	}
	var acct *uuid.UUID
	if err := tx.QueryRow(ctx, `select account_id from hub.people where id=$1`, pl.Person).Scan(&acct); err == nil && acct != nil && (c.Sector != nil || c.Branch != nil || c.Unit != nil || c.Department != nil) {
		if err := chat.SyncAudience(ctx, tx, s.A.Sign, a.Org, *acct); err != nil {
			return err
		}
	}
	if err := keepOwner(ctx, tx, pl.Person, c.Role); err != nil {
		return err
	}
	if c.Role != nil && *c.Role == "" {
		if _, err := tx.Exec(ctx, `delete from hub.person_roles where person_id=$1`, pl.Person); err != nil {
			return err
		}
	} else if c.Role != nil {
		if _, err := tx.Exec(ctx, `insert into hub.person_roles (org_id, person_id, role_id) select $3, $1, id from hub.roles where key=$2
			on conflict (org_id, person_id) do update set role_id=excluded.role_id, `+integrity.BumpOn("person_roles"), pl.Person, *c.Role, a.Org); err != nil {
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

// createLogin gives a newly added person an active account with a first password (they must change it at sign-in) and
// joins the account to the person's record. It fails, leaving the invitation link as the way in, when the email already
// has an account.
func (s *Service) createLogin(ctx context.Context, org, person uuid.UUID, name, email string) (string, error) {
	temp := crypto.RandomToken(9) + "aA1"
	rnd := func() string { return crypto.RandomToken(9) }
	quiz := []auth.QuizIn{{Question: "Set in profile 1", Answer: rnd()}, {Question: "Set in profile 2", Answer: rnd()}, {Question: "Set in profile 3", Answer: rnd()}}
	acct, err := s.Auth.Signup(ctx, name, email, temp, quiz, "")
	if err != nil {
		return "", err
	}
	err = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		if err := db.SetOrg(ctx, tx, org); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update hub.accounts set must_change=true, `+integrity.Bump+` where id=$1`, acct); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "accounts", "id=$1", acct); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update hub.people set account_id=$2, status='active', joined_on=current_date, `+integrity.Bump+` where id=$1 and account_id is null`, person, acct); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "people", "id=$1", person); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.account_orgs (account_id, org_id) values ($1,$2) on conflict do nothing`, acct, org); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "account_orgs", "account_id=$1 and org_id=$2", acct, org); err != nil {
			return err
		}
		return chat.SyncAudience(ctx, tx, s.A.Sign, org, acct)
	})
	if err != nil {
		return "", err
	}
	return temp, nil
}

// changeEmail gives a person a new email address: on their record, and (when they have an account) as their sign-in. Their old
// address is told.
func (s *Service) changeEmail(ctx context.Context, tx pgx.Tx, a approvals.Applied, person uuid.UUID, email string) error {
	e, err := auth.NormEmail(email)
	if err != nil {
		return err
	}
	enc, err := pp.Enc(a.DEK, a.Org, person, pp.ColEmail, e)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "savepoint person_email"); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `update hub.people set email_idx=$2, email_enc=$3, updated_at=now(), `+integrity.Bump+` where id=$1 and status <> 'deleted'`, person, pp.EmailIdx(s.A.Cfg.BlindKey, a.Org, e), enc); err != nil {
		if strings.Contains(err.Error(), "people_email_uidx") {
			_, _ = tx.Exec(ctx, "rollback to savepoint person_email")
			return httpx.Err(409, "email_in_use", "Someone with this email is already in the organization.")
		}
		return err
	}
	if err := s.A.Sign.Seal(ctx, tx, "people", "id=$1", person); err != nil {
		return err
	}
	var acct *uuid.UUID
	if err := tx.QueryRow(ctx, `select account_id from hub.people where id=$1`, person).Scan(&acct); err != nil || acct == nil {
		return nil
	}
	old, err := s.Auth.ChangeAccountEmail(ctx, tx, *acct, e)
	if err != nil {
		return err
	}
	if old != "" && old != e {
		go func() {
			_ = hubmail.SendCard(context.Background(), s.A.Mail, old, "The email you sign in to Hub with was changed", hubmail.Card{Accent: hubmail.Amber, Title: "Your sign-in email was changed",
				Preheader: "An administrator changed it", Lead: "An administrator changed the email address of your Hub account to " + e + ". Use the new address to sign in. If you did not expect this, contact your administrator."})
		}()
	}
	return nil
}

// Photo returns a person's profile picture (the bytes and their type). Anyone signed in to the organization may fetch it: it is what
// shows next to a name on a post, a comment or a chat, which they can already see.
func (s *Service) Photo(ctx context.Context, p *tenant.Principal, id uuid.UUID) ([]byte, string, error) {
	var out []byte
	var typ string
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var enc []byte
		var t *string
		err := tx.QueryRow(ctx, `select pp.photo_enc, pp.photo_type from hub.person_profiles pp join hub.people pe on pe.id = pp.person_id
			where pp.person_id=$1 and pe.status <> 'deleted'`, id).Scan(&enc, &t)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && (len(enc) == 0 || t == nil)) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		b, err := crypto.FieldOpen(dek, profileCtx(p.Org, id, "photo_enc"), enc)
		if err != nil {
			return err
		}
		out, typ = b, *t
		return nil
	})
	return out, typ, err
}

// Photos returns many people's pictures at once (up to 40), so a page with many authors makes one request, not one per person.
// Only picture types that cannot run script are returned; people without a picture are left out.
func (s *Service) Photos(ctx context.Context, p *tenant.Principal, ids []uuid.UUID) (map[string]map[string]string, error) {
	out := map[string]map[string]string{}
	if len(ids) == 0 {
		return out, nil
	}
	if len(ids) > 40 {
		ids = ids[:40]
	}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select pp.person_id, pp.photo_enc, pp.photo_type from hub.person_profiles pp join hub.people pe on pe.id = pp.person_id
			where pp.person_id = any($1) and pe.status <> 'deleted' and pp.photo_enc is not null and pp.photo_type is not null`, ids)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var id uuid.UUID
			var enc []byte
			var t string
			if err := rs.Scan(&id, &enc, &t); err != nil {
				return err
			}
			switch t {
			case "image/png", "image/jpeg", "image/webp", "image/gif":
			default:
				continue
			}
			b, err := crypto.FieldOpen(dek, profileCtx(p.Org, id, "photo_enc"), enc)
			if err != nil || len(b) == 0 {
				continue
			}
			out[id.String()] = map[string]string{"type": t, "data": base64.StdEncoding.EncodeToString(b)}
		}
		return rs.Err()
	})
	return out, err
}

// roleOnly says whether the change only moves the person to another role.
func (c Changes) roleOnly() bool { return c.Role != nil && c == Changes{Role: c.Role} }

// keepOwner lets the owner role move from one person to another, but never leaves the organization without an active owner:
// it is given to someone else first, then the last holder may take another role.
func keepOwner(ctx context.Context, tx pgx.Tx, person uuid.UUID, role *string) error {
	if role == nil {
		return nil
	}
	var had, stays bool
	var name string
	var others int
	err := tx.QueryRow(ctx, `select coalesce(r.locked, false), coalesce(r.name, ''), coalesce((select locked from hub.roles where key=$2), false),
		(select count(*) from hub.person_roles x join hub.roles y on y.id = x.role_id join hub.people z on z.id = x.person_id where y.locked and z.status = 'active' and z.id <> $1)
		from hub.people p left join hub.person_roles pr on pr.person_id = p.id left join hub.roles r on r.id = pr.role_id where p.id=$1`, person, *role).Scan(&had, &name, &stays, &others)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil || !had || stays || others > 0 {
		return err
	}
	return httpx.Err(409, "last_owner", "Give the "+name+" role to someone else first: the organization always needs an active "+name+".")
}
