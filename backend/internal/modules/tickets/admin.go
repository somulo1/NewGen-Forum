package tickets

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/tenant"
)

func (s *Service) Settings(ctx context.Context, p *tenant.Principal) (Config, error) {
	var c Config
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "tickets", "View"); err != nil {
			return err
		}
		var err error
		c, err = s.config(ctx, tx)
		return err
	})
	return c, err
}

func (s *Service) SetSettings(ctx context.Context, p *tenant.Principal, c Config) error {
	if c.NearPct < 10 || c.NearPct > 99 {
		return httpx.Bad("Choose a warning point between 10% and 99% of the TAT.")
	}
	if hm(c.Start, -1) < 0 || hm(c.End, -1) < 0 || hm(c.End, 0) <= hm(c.Start, 0) {
		return httpx.Bad("Working hours must look like 08:00 to 17:00.")
	}
	if c.loc().String() == "UTC" && c.Timezone != "UTC" {
		return httpx.Bad("Unknown time zone.")
	}
	if len(c.Holidays) > 400 || len(c.Escalate) > 80 {
		return httpx.Bad("Too long.")
	}
	raw, _ := json.Marshal(c)
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "tickets", "Manage TAT settings"); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.ticket_settings (org_id, config) values ($1,$2) on conflict (org_id) do update set config=excluded.config, `+integrity.BumpOn("ticket_settings"), p.Org, raw); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "ticket_settings", "org_id=$1", p.Org); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Changed ticket TAT settings", "Tickets", p.IP, c)
	})
}

// Categories lists the categories a recipient offers: its own, or the desk's when it has none.
func (s *Service) Categories(ctx context.Context, p *tenant.Principal, toType string, toRef *uuid.UUID) ([]Category, error) {
	out := []Category{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "tickets", "View"); err != nil {
			return err
		}
		q := func(t string, r *uuid.UUID) error {
			rs, err := tx.Query(ctx, `select id, unit_type, unit_ref, name, tat_hours, priority, members from hub.ticket_categories where unit_type=$1 and unit_ref is not distinct from $2 order by name`, t, r)
			if err != nil {
				return err
			}
			defer rs.Close()
			for rs.Next() {
				var c Category
				if err := rs.Scan(&c.ID, &c.UnitType, &c.UnitRef, &c.Name, &c.TAT, &c.Priority, &c.Members); err != nil {
					return err
				}
				out = append(out, c)
			}
			return rs.Err()
		}
		if err := q(toType, toRef); err != nil {
			return err
		}
		if len(out) == 0 && toType != "desk" {
			return q("desk", nil)
		}
		return nil
	})
	return out, err
}

func (s *Service) SaveCategory(ctx context.Context, p *tenant.Principal, id *uuid.UUID, c Category) (*Category, error) {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" || len(c.Name) > 60 || c.TAT < 1 || c.TAT > 24*90 {
		return nil, httpx.Bad("Give the category a name and a TAT between 1 hour and 90 days.")
	}
	if c.Priority == "" {
		c.Priority = "Medium"
	}
	if c.UnitType == "" {
		c.UnitType = "desk"
	}
	if c.Members == nil {
		c.Members = []uuid.UUID{}
	}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "tickets", "Set up ticket categories"); err != nil {
			return err
		}
		if id == nil {
			c.ID = uuid.New()
			if _, err := tx.Exec(ctx, `insert into hub.ticket_categories (org_id, id, unit_type, unit_ref, name, tat_hours, priority, members) values ($1,$2,$3,$4,$5,$6,$7,$8)`,
				p.Org, c.ID, c.UnitType, c.UnitRef, c.Name, c.TAT, c.Priority, c.Members); err != nil {
				return err
			}
		} else {
			c.ID = *id
			tag, err := tx.Exec(ctx, `update hub.ticket_categories set name=$2, tat_hours=$3, priority=$4, members=$5, `+integrity.Bump+` where id=$1`, c.ID, c.Name, c.TAT, c.Priority, c.Members)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return httpx.ErrNotFound
			}
		}
		if err := s.A.Sign.Seal(ctx, tx, "ticket_categories", "id=$1", c.ID); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Saved ticket category "+c.Name, "Tickets", p.IP, c)
	})
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) DeleteCategory(ctx context.Context, p *tenant.Principal, id uuid.UUID) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "tickets", "Set up ticket categories"); err != nil {
			return err
		}
		var name string
		err := tx.QueryRow(ctx, `delete from hub.ticket_categories where id=$1 returning name`, id).Scan(&name)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Removed ticket category "+name, "Tickets", p.IP, map[string]any{"id": id})
	})
}
