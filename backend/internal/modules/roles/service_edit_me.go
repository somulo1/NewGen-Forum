package roles

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/people"
	"hub/internal/tenant"
)

// EditMe saves the caller's own profile. This never needs approval: it is their own page.
func (s *Service) EditMe(ctx context.Context, p *tenant.Principal, e ProfileEdit) error {
	if e.Availability != nil && !availabilities[*e.Availability] {
		return httpx.Bad("Choose a status from the list.")
	}
	if e.Bio != nil && len(*e.Bio) > 2000 {
		return httpx.Bad("Keep your bio under 2,000 characters.")
	}
	if e.Phone != nil && *e.Phone != "" && (len(*e.Phone) < 5 || len(*e.Phone) > 24) {
		return httpx.Bad("Enter a valid phone number.")
	}
	if e.Cover != nil && (len(*e.Cover) != 7 || (*e.Cover)[0] != '#') {
		return httpx.Bad("Choose a colour like #1F3A2B.")
	}
	var headline string
	if e.Headline != nil {
		headline = strings.Join(strings.Fields(*e.Headline), " ")
		if len([]rune(headline)) > 120 {
			return httpx.Bad("Keep your headline under 120 characters.")
		}
	}
	var skills []string
	if e.Skills != nil {
		for _, k := range *e.Skills {
			if k = strings.TrimSpace(k); k != "" && len(k) <= 40 {
				skills = append(skills, k)
			}
		}
		if len(skills) > 30 {
			return httpx.Bad("Add up to 30 skills.")
		}
	}
	var photoBytes []byte
	var photoType string
	if e.Photo != nil && *e.Photo != "" {
		head, data, ok := strings.Cut(*e.Photo, ",")
		if !ok || !strings.HasPrefix(head, "data:image/") || !strings.HasSuffix(head, ";base64") {
			return httpx.Bad("Choose an image file for your profile photo.")
		}
		photoType = strings.TrimSuffix(strings.TrimPrefix(head, "data:"), ";base64")
		if photoType != "image/png" && photoType != "image/jpeg" && photoType != "image/webp" && photoType != "image/gif" {
			return httpx.Bad("Use a PNG, JPEG, WebP or GIF photo.")
		}
		var err error
		if photoBytes, err = base64.StdEncoding.DecodeString(data); err != nil {
			return httpx.Bad("The photo could not be read.")
		}
		if len(photoBytes) > maxPhoto {
			return httpx.Bad("Profile photos must be 512 KB or smaller.")
		}
	}
	if e.Phone != nil && people.IsMasked(*e.Phone) { // a masked value sent back must never replace the real one
		return httpx.Bad("Type your full new phone number, or leave the hidden one as it is.")
	}
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		if e.Availability != nil || e.Phone != nil {
			var phone any
			if e.Phone != nil {
				b, _ := crypto.Field(dek, crypto.Context{Org: p.Org.String(), Table: "people", Column: "phone_enc", RowID: p.Person.String()}, []byte(*e.Phone))
				if *e.Phone == "" {
					b = nil
				}
				phone = b
			}
			if _, err := tx.Exec(ctx, `update hub.people set availability=coalesce($2, availability), phone_enc = case when $3::boolean then $4 else phone_enc end, updated_at=now(), `+integrity.Bump+` where id=$1`,
				p.Person, e.Availability, e.Phone != nil, phone); err != nil {
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, "people", "id=$1", p.Person); err != nil {
				return err
			}
		}
		if e.Headline != nil {
			if headline == "" {
				if _, err := tx.Exec(ctx, `delete from hub.person_headlines where person_id=$1`, p.Person); err != nil {
					return err
				}
			} else {
				if _, err := tx.Exec(ctx, `insert into hub.person_headlines (org_id, person_id, headline) values ($1,$2,$3) on conflict do nothing`, p.Org, p.Person, headline); err != nil {
					return err
				}
				if err := s.A.Sign.Seal(ctx, tx, "person_headlines", "person_id=$1", p.Person); err != nil {
					return err
				}
				if _, err := tx.Exec(ctx, `update hub.person_headlines set headline=$2, updated_at=now(), `+integrity.Bump+` where person_id=$1`, p.Person, headline); err != nil {
					return err
				}
				if err := s.A.Sign.Seal(ctx, tx, "person_headlines", "person_id=$1", p.Person); err != nil {
					return err
				}
			}
		}
		if e.Cover != nil || e.Photo != nil || e.Bio != nil || e.Skills != nil {
			if _, err := tx.Exec(ctx, `insert into hub.person_profiles (org_id, person_id) values ($1,$2) on conflict do nothing`, p.Org, p.Person); err != nil {
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, "person_profiles", "person_id=$1", p.Person); err != nil {
				return err
			}
			sealIf := func(col string, plain []byte, set bool) (any, error) {
				if !set {
					return nil, nil
				}
				if len(plain) == 0 {
					return []byte{}, nil
				}
				return crypto.Field(dek, pctx(p.Org, p.Person, col), plain)
			}
			bio, err := sealIf("bio_enc", []byte(deref(e.Bio)), e.Bio != nil)
			if err != nil {
				return err
			}
			var sk any
			if e.Skills != nil {
				raw, _ := json.Marshal(skills)
				if sk, err = sealIf("skills_enc", raw, true); err != nil {
					return err
				}
			}
			ph, err := sealIf("photo_enc", photoBytes, e.Photo != nil)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `update hub.person_profiles set
				bio_enc = case when $2::bytea is null then bio_enc when length($2)=0 then null else $2 end,
				skills_enc = case when $3::bytea is null then skills_enc else $3 end,
				photo_enc = case when $4::bytea is null then photo_enc when length($4)=0 then null else $4 end,
				photo_type = case when $4::bytea is null then photo_type when length($4)=0 then null else $5 end,
				cover = coalesce($6, cover), updated_at=now(), `+integrity.Bump+` where person_id=$1`, p.Person, bio, sk, ph, photoType, e.Cover); err != nil {
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, "person_profiles", "person_id=$1", p.Person); err != nil {
				return err
			}
		}
		return nil
	})
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ---- the person's app settings (home widgets, theme ...) ----

func prefsCtx(org, person uuid.UUID) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "person_prefs", Column: "data_enc", RowID: person.String()}
}

func (s *Service) Prefs(ctx context.Context, p *tenant.Principal) (json.RawMessage, error) {
	out := json.RawMessage(`{}`)
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var enc []byte
		err := tx.QueryRow(ctx, `select data_enc from hub.person_prefs where person_id=$1`, p.Person).Scan(&enc)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		raw, err := crypto.FieldOpen(dek, prefsCtx(p.Org, p.Person), enc)
		if err != nil {
			return err
		}
		out = raw
		return nil
	})
	return out, err
}

func (s *Service) SetPrefs(ctx context.Context, p *tenant.Principal, raw json.RawMessage) error {
	var probe map[string]any
	if len(raw) > 16<<10 || json.Unmarshal(raw, &probe) != nil || probe == nil {
		return httpx.Bad("Settings must be a JSON object under 16 KB.")
	}
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		enc, err := crypto.Field(dek, prefsCtx(p.Org, p.Person), raw)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.person_prefs (org_id, person_id, data_enc) values ($1,$2,$3) on conflict (org_id, person_id) do update set data_enc=excluded.data_enc, updated_at=now(), `+integrity.BumpOn("person_prefs"), p.Org, p.Person, enc); err != nil {
			return err
		}
		return s.A.Sign.Seal(ctx, tx, "person_prefs", "person_id=$1", p.Person)
	})
}

// PersonRights lists the personal overrides one person has on top of their role.
func (s *Service) PersonRights(ctx context.Context, p *tenant.Principal, person uuid.UUID) ([]map[string]any, error) {
	out := []map[string]any{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select module, right_name, allowed from hub.person_rights where person_id=$1 order by module, right_name`, person)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var m, r string
			var a bool
			if err := rs.Scan(&m, &r, &a); err != nil {
				return err
			}
			out = append(out, map[string]any{"module": m, "right": r, "allowed": a})
		}
		return rs.Err()
	})
	return out, err
}
