package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/tenant"
)

func droppedFile(old, next map[string]any) bool {
	now := map[string]bool{}
	list, _ := next["files"].([]any)
	for _, f := range list {
		b, _ := json.Marshal(f)
		now[string(b)] = true
	}
	was, _ := old["files"].([]any)
	for _, f := range was {
		b, _ := json.Marshal(f)
		if !now[string(b)] {
			return true
		}
	}
	return false
}

func contains(l []string, v string) bool {
	for _, x := range l {
		if x == v {
			return true
		}
	}
	return false
}

// Delete removes an item. The history stays. A task others depend on, or with tasks under it, has to be untied first.
func (s *Service) Delete(ctx context.Context, p *tenant.Principal, project, id uuid.UUID) error {
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		pr, err := s.member(ctx, tx, p, project, true)
		if err != nil {
			return err
		}
		var coll string
		var author *uuid.UUID
		if err := tx.QueryRow(ctx, `select collection, created_by from hub.project_items where id=$1 and project_id=$2`, id, project).Scan(&coll, &author); err != nil {
			return httpx.ErrNotFound
		}
		switch coll {
		case "task":
			err = authz.Require(ctx, tx, p.Person, "projects", "Delete tasks")
		case "page":
			err = authz.Require(ctx, tx, p.Person, "projects", "Delete pages")
			if err == nil {
				pages, e := loadPages(ctx, tx, project)
				if e != nil {
					return e
				}
				if !pageAllowed(pages, id.String(), pr, p.Person, "edit") {
					return httpx.Err(403, "page_restricted", "This page's restrictions do not let you change it.")
				}
			}
		case "risk":
			err = authz.Require(ctx, tx, p.Person, "projects", "Delete risks")
		case "activity":
			err = authz.Require(ctx, tx, p.Person, "projects", "Delete activity")
		case "milestone":
			err = authz.Require(ctx, tx, p.Person, "projects", "Delete milestones")
		case "sprint":
			err = authz.Require(ctx, tx, p.Person, "projects", "Delete sprints")
		case "page_comment":
			// your own comment, with the right
			if err = authz.Require(ctx, tx, p.Person, "projects", "Delete comments"); err == nil && (author == nil || *author != p.Person) {
				err = httpx.ErrForbidden
			}
		default:
			err = mayChange(ctx, tx, p, pr, coll, false, author)
		}
		if err != nil {
			return err
		}
		if coll == "task" {
			var used bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from hub.project_items where project_id=$1 and collection='task' and (data->>'parent' = $2 or data->'deps' ? $2))`, project, id.String()).Scan(&used); err != nil {
				return err
			}
			if used {
				return httpx.Err(409, "in_use", "Other tasks sit under this one or depend on it. Change them first.")
			}
		}
		if coll == "milestone" {
			var used bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from hub.project_items where project_id=$1 and collection in ('sprint','task') and data->>'milestone' = $2)`, project, id.String()).Scan(&used); err != nil {
				return err
			}
			if used {
				return httpx.Err(409, "in_use", "Sprints or tasks still sit under this. Move or delete them first.")
			}
		}
		if coll == "sprint" {
			sprints, err := loadItems(ctx, tx, project, "sprint")
			if err != nil {
				return err
			}
			for _, sp := range sprints {
				if closed, _ := sp.data["closed"].(bool); closed && sp.id == id {
					return httpx.Err(409, "closed", "A closed sprint is kept, because velocity reads from it.")
				}
			}
			// its work goes back to the backlog
			tasks, err := loadItems(ctx, tx, project, "task")
			if err != nil {
				return err
			}
			for _, t := range tasks {
				if text(t.data, "sprint") == id.String() {
					t.data["sprint"] = ""
					if err := s.writeBySystem(ctx, tx, p.Org, project, t, "sprint", id.String(), ""); err != nil {
						return err
					}
				}
			}
		}
		if _, err := tx.Exec(ctx, `insert into hub.project_item_history (org_id, item_id, project_id, person_id, who, field) values ($1,$2,$3,$4,$5,'deleted')`, p.Org, id, project, p.Person, p.Name); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `delete from hub.project_items where id=$1`, id)
		return err
	})
	if err == nil {
		s.planChanged(ctx, p, project)
	}
	return err
}

// DeleteProject removes a project and its plan for good. The audit log keeps that it happened; the conversation stays, so its discussion is not lost.
func (s *Service) DeleteProject(ctx context.Context, p *tenant.Principal, project uuid.UUID) error {
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		pr, err := s.member(ctx, tx, p, project, true)
		if err != nil {
			return err
		}
		if err := authz.Require(ctx, tx, p.Person, "projects", "Delete project"); err != nil {
			return err
		}
		for _, q := range []string{`delete from hub.project_item_history where project_id=$1`, `delete from hub.project_items where project_id=$1`, `delete from hub.projects where conversation_id=$1`} {
			if _, err := tx.Exec(ctx, q, project); err != nil {
				return err
			}
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Deleted project "+pr.Code, "Projects", p.IP, map[string]any{"project": project})
	})
	if err == nil {
		s.planChanged(ctx, p, project)
	}
	return err
}

// fillViewers marks which projects the person is on, and gives those they are not on what their conversation would have told them.
func (s *Service) fillViewers(ctx context.Context, tx pgx.Tx, p *tenant.Principal, list []*Project) error {
	if len(list) == 0 {
		return nil
	}
	dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
	if err != nil {
		return err
	}
	// three queries for the whole page, whatever its size (not three per project)
	ids := make([]uuid.UUID, len(list))
	by := make(map[uuid.UUID]*Project, len(list))
	for i, x := range list {
		ids[i] = x.ID
		by[x.ID] = x
		x.Member = false
	}
	rs, err := tx.Query(ctx, `select conversation_id from hub.conversation_members where conversation_id = any($1) and account_id=$2 and left_at is null`, ids, p.Account)
	if err != nil {
		return err
	}
	for rs.Next() {
		var id uuid.UUID
		if err := rs.Scan(&id); err != nil {
			rs.Close()
			return err
		}
		by[id].Member = true
	}
	rs.Close()
	if err := rs.Err(); err != nil {
		return err
	}
	var look []uuid.UUID
	for _, x := range list {
		if !x.Member {
			look = append(look, x.ID)
		}
	}
	if len(look) == 0 {
		return nil
	}
	rs, err = tx.Query(ctx, `select id, name_enc, about_enc from hub.conversations where id = any($1)`, look)
	if err != nil {
		return err
	}
	for rs.Next() {
		var id uuid.UUID
		var name, about []byte
		if err := rs.Scan(&id, &name, &about); err != nil {
			rs.Close()
			return err
		}
		open := func(col string, ct []byte) string {
			if len(ct) == 0 {
				return ""
			}
			b, err := crypto.FieldOpen(dek, crypto.Context{Org: p.Org.String(), Table: "conversations", Column: col, RowID: id.String()}, ct)
			if err != nil {
				return ""
			}
			return string(b)
		}
		by[id].Name, by[id].About = open("name_enc", name), open("about_enc", about)
	}
	rs.Close()
	if err := rs.Err(); err != nil {
		return err
	}
	rs, err = tx.Query(ctx, `select m.conversation_id, pe.first_name || ' ' || pe.last_name from hub.conversation_members m join hub.people pe on pe.account_id = m.account_id
		where m.conversation_id = any($1) and m.left_at is null order by 1, 2`, look)
	if err != nil {
		return err
	}
	defer rs.Close()
	for rs.Next() {
		var id uuid.UUID
		var n string
		if err := rs.Scan(&id, &n); err != nil {
			return err
		}
		by[id].TeamNames = append(by[id].TeamNames, n)
	}
	return rs.Err()
}

// eligibleSQL: people who can be on a project: their role, or a right given to them personally, lets them view projects.
const eligibleSQL = `select p.id from hub.people p
	left join hub.person_roles pr on pr.person_id = p.id
	left join hub.person_rights x on x.person_id = p.id and x.module = 'projects' and x.right_name = 'View'
	where p.status = 'active' and coalesce(x.allowed, case when pr.role_id is null then null else hub.role_right(pr.role_id, 'projects', 'View') end, false)`

// Eligible lists who may be chosen for a project team.
func (s *Service) Eligible(ctx context.Context, p *tenant.Principal) ([]uuid.UUID, error) {
	out := []uuid.UUID{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := s.ensureRights(ctx, tx, p.Org); err != nil {
			return err
		}
		can, err := authz.Can(ctx, tx, p.Person, "projects", "Create project")
		if err != nil {
			return err
		}
		if !can {
			if err := authz.Require(ctx, tx, p.Person, "projects", "Manage team"); err != nil {
				return err
			}
		}
		rs, err := tx.Query(ctx, eligibleSQL)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var id uuid.UUID
			if err := rs.Scan(&id); err != nil {
				return err
			}
			out = append(out, id)
		}
		return rs.Err()
	})
	return out, err
}

// checkTeam refuses a project whose conversation includes someone without access to projects.
func checkTeam(ctx context.Context, tx pgx.Tx, conv uuid.UUID) error {
	var name string
	err := tx.QueryRow(ctx, `select p.first_name || ' ' || p.last_name from hub.conversation_members m join hub.people p on p.account_id = m.account_id
		where m.conversation_id = $1 and m.left_at is null and p.id not in (`+eligibleSQL+`) limit 1`, conv).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return httpx.Err(400, "not_eligible", name+" has no role or permission that allows projects, so cannot join this team.")
}

// ImportItem is one record of an older encrypted plan, opened by a member's device.
type ImportItem struct {
	ID         uuid.UUID      `json:"id"`
	Collection string         `json:"collection"`
	Data       map[string]any `json:"data"`
	CreatedAt  *time.Time     `json:"created_at"`
}

// Import moves an older encrypted plan over, once. Any team member's device may do it: they can all read it already.
func (s *Service) Import(ctx context.Context, p *tenant.Principal, project uuid.UUID, items []ImportItem) error {
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		pr, err := s.member(ctx, tx, p, project, true)
		if err != nil {
			return err
		}
		if pr.Imported {
			return nil // another member's device got there first
		}
		ref := 0
		for _, it := range items {
			if !collection[it.Collection] || it.Data == nil {
				return httpx.Bad("Unknown kind of item.")
			}
			var r *int
			if it.Collection == "task" {
				ref++
				n := ref
				r = &n
			}
			at := time.Now()
			if it.CreatedAt != nil {
				at = *it.CreatedAt
			}
			b, _ := json.Marshal(it.Data)
			if _, err := tx.Exec(ctx, `insert into hub.project_items (org_id, id, project_id, collection, ref, data, created_at) values ($1,$2,$3,$4,$5,$6,$7) on conflict do nothing`, p.Org, it.ID, project, it.Collection, r, b, at); err != nil {
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, "project_items", "id=$1", it.ID); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `update hub.projects set plan_imported_at=now(), updated_at=now(), `+integrity.Bump+` where conversation_id=$1`, project); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "projects", "conversation_id=$1", project); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Moved the plan of %s out of end-to-end encryption", pr.Code), "Projects", p.IP, map[string]any{"project": project, "items": len(items)})
	})
	if err == nil {
		s.planChanged(ctx, p, project)
	}
	return err
}
