package people

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/db"
	"hub/internal/httpx"
	"hub/internal/integrity"
	hubmail "hub/internal/mail"
	"hub/internal/modules/chat"
	"hub/internal/orgsettings"
	pp "hub/internal/people"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

// Onboard adds people (one or a whole CSV import) and emails each an invitation. One bad row never blocks the others.
func (s *Service) Onboard(ctx context.Context, p *tenant.Principal, list []NewPerson) ([]Onboarded, error) {
	if len(list) == 0 || len(list) > 500 {
		return nil, httpx.Bad("Send between 1 and 500 people at a time.")
	}
	out := make([]Onboarded, len(list))
	type mailJob struct {
		to, link, name string
		id             uuid.UUID
		inviteLink     bool
	}
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
		type roleRef struct {
			id     uuid.UUID
			locked bool
		}
		roleByKey := map[string]roleRef{}
		rs, err := tx.Query(ctx, `select key, id, locked from hub.roles`)
		if err != nil {
			return err
		}
		for rs.Next() {
			var k string
			var r roleRef
			_ = rs.Scan(&k, &r.id, &r.locked)
			roleByKey[k] = r
		}
		rs.Close()
		// the organization's default role (if it set one) is what a person gets when no role is chosen
		var defaultKey string
		_ = tx.QueryRow(ctx, `select r.key from hub.default_role d join hub.roles r on r.id = d.role_id`).Scan(&defaultKey)
		// every level the organization has places in is mandatory: everyone is placed in a sector, branch, unit and department
		// (those it uses), so each person has the dashboards their widgets decide
		haveKind := map[string]bool{}
		if ks, err := tx.Query(ctx, `select distinct kind from hub.org_units`); err == nil {
			for ks.Next() {
				var k string
				_ = ks.Scan(&k)
				haveKind[k] = true
			}
			ks.Close()
		}
		missingLevel := func(n NewPerson) string {
			for _, x := range []struct {
				kind string
				v    *uuid.UUID
			}{{"sector", n.Sector}, {"branch", n.Branch}, {"unit", n.Unit}, {"department", n.Department}} {
				if haveKind[x.kind] && x.v == nil {
					return "Choose a " + x.kind + ": every person must be placed in one."
				}
			}
			return ""
		}
		for i := range list {
			n := list[i]
			out[i] = Onboarded{Email: n.Email}
			if err := validate(&n); err != nil {
				out[i].Status, out[i].Error = "rejected", err.(*httpx.Error).Message
				continue
			}
			if m := missingLevel(n); m != "" {
				out[i].Status, out[i].Error = "rejected", m
				continue
			}
			out[i].Email = n.Email
			role := n.Role
			if role == "" {
				role = defaultKey
			}
			// Everyone in an organization has a role, and the role is what gives them rights: there are no people without one.
			if role == "" {
				out[i].Status, out[i].Error = "rejected", "Choose a role for this person (or set a default role on the Roles page)."
				continue
			}
			var rid uuid.UUID
			if role != "" {
				ref, ok := roleByKey[role]
				if !ok || ref.locked {
					out[i].Status, out[i].Error = "rejected", "That role cannot be given when adding a person."
					continue
				}
				rid = ref.id
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
			jobs = append(jobs, mailJob{to: n.Email, link: link, name: n.First + " " + n.Last, id: id, inviteLink: n.InviteLink})
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
		// The person gets a working account at once: active, with a first password in the email that they must change when they sign in.
		var temp string
		var err error = errors.New("invitation link requested")
		if !j.inviteLink {
			temp, err = s.createLogin(ctx, p.Org, j.id, j.name, j.to)
		}
		if err == nil {
			login := strings.TrimRight(s.A.Cfg.AllowedOrigins[0], "/") + "/login"
			_ = hubmail.SendCard(ctx, s.A.Mail, j.to, "Your Hub account at "+orgName+" is ready", hubmail.Card{Accent: hubmail.Green, Org: orgName, Title: "Your account is ready", Preheader: "Sign in with the password in this email",
				Lead: "Your administrator added you to " + orgName + " on Hub. Sign in with the email and first password below. You will choose your own password the first time.",
				Rows: [][2]string{{"Email", j.to}, {"First password", temp}}, Action: [2]string{"Sign in", login}, Note: "Keep this email private. The first password works until you change it."})
			continue
		}
		if !j.inviteLink {
			slog.Warn("could not create the account for a new person; sending the invitation link instead", "err", err)
		}
		_ = hubmail.SendCard(ctx, s.A.Mail, j.to, "You have been added to "+orgName+" on Hub", hubmail.Card{Accent: hubmail.Green, Org: orgName, Title: "You have been added to " + orgName, Preheader: "Join your team on Hub",
			Lead: "Your administrator added you to " + orgName + " on Hub. Open the link to set up your account and join.", Action: [2]string{"Join " + orgName, j.link}, Note: "The link works once and expires in 7 days."})
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
	if role != uuid.Nil {
		if _, err := tx.Exec(ctx, `insert into hub.person_roles (org_id, person_id, role_id) values ($1,$2,$3)`, p.Org, id, role); err != nil {
			return "", err
		}
		if err := s.A.Sign.Seal(ctx, tx, "person_roles", "person_id=$1", id); err != nil {
			return "", err
		}
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
		if err := s.A.Sign.Seal(ctx, tx, "account_orgs", "account_id=$1 and org_id=$2", account, org); err != nil {
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
		s.Auth.RevokeAccountInOrg(ctx, *account, p.Org, target) // they are signed out of this organization right away, the phone too
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
	Email      *string    `json:"email,omitempty"`
	Role       *string    `json:"role,omitempty"`
	Manager    *uuid.UUID `json:"manager_id,omitempty"`
}

type updatePayload struct {
	Person  uuid.UUID `json:"person"`
	Changes Changes   `json:"changes"`
}
