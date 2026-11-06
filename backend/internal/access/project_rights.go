package access

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// projectRights are the finer rights of the Projects widget. An organization made before them gets them once, and each role keeps what it
// could already do: a new right starts as its older right was. Rights with no older one start switched off (the project manager still has them).
var projectRights = []struct{ Name, From string }{
	{"View documents", "View"},
	{"View reports", "View"},
	{"View timeline", "View"},
	{"View change history", "View"},
	{"Delete tasks", "Edit tasks"},
	{"Rank backlog", "Edit tasks"},
	{"Plan sprints", "Manage milestones"},
	{"Close sprints", "Manage milestones"},
	{"Reschedule tasks", "Manage milestones"},
	{"Save baseline", "Manage milestones"},
	{"Request gate sign-off", "Manage milestones"},
	{"Manage workflow", "Manage team"},
	{"Change project type", "Manage team"},
	{"Finish tasks", ""},
	{"Edit any task", ""},
	{"Create pages", "Edit tasks"},
	{"Edit pages", "Edit tasks"},
	{"Delete pages", "Edit tasks"},
	{"Lock pages", "Manage team"},
	{"Comment on pages", "View"},
	{"Attach to pages", "Edit tasks"},
	{"Export pages", "Export"},
	{"Import pages", "Create pages"},
	{"Restrict pages", "Lock pages"},
	{"Delete project", ""},
	{"Delete milestones", "Manage milestones"},
	{"Delete sprints", "Plan sprints"},
	{"Delete comments", "Comment on pages"},
	{"Remove attachments", "Attach to pages"},
	{"Remove workflow statuses", "Manage workflow"},
	{"Delete risks", "Manage team"},
	{"Delete activity", ""},
}

// EnsureProjectRights is safe to run on every request: it does nothing once the rights exist.
func EnsureProjectRights(ctx context.Context, tx pgx.Tx, org uuid.UUID, seal func(table, where string, args ...any) error) error {
	var done, any bool
	if err := tx.QueryRow(ctx, `select exists (select 1 from hub.rights where module='projects' and name=$1), exists (select 1 from hub.rights where module='projects')`, projectRights[len(projectRights)-1].Name).Scan(&done, &any); err != nil || done || !any {
		return err
	}
	for i, r := range projectRights {
		tag, err := tx.Exec(ctx, `insert into hub.rights (org_id, module, name, label, kind, sort, builtin) values ($1,'projects',$2,$2,$3,$4,true) on conflict do nothing`, org, r.Name, KindOf(r.Name), 20+i)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 || r.From == "" {
			continue
		}
		for _, table := range []struct{ name, key string }{{"role_rights", "role_id"}, {"person_rights", "person_id"}, {"role_group_rights", "group_id"}} {
			if _, err := tx.Exec(ctx, `insert into hub.`+table.name+` (org_id, `+table.key+`, module, right_name, allowed) select org_id, `+table.key+`, module, $1, allowed from hub.`+table.name+` where module='projects' and right_name=$2 on conflict do nothing`, r.Name, r.From); err != nil {
				return err
			}
		}
	}
	if err := seal("rights", "org_id=$1 and module='projects'", org); err != nil {
		return err
	}
	if err := seal("role_rights", "org_id=$1 and module='projects'", org); err != nil {
		return err
	}
	if err := seal("person_rights", "org_id=$1 and module='projects'", org); err != nil {
		return err
	}
	if err := seal("role_group_rights", "org_id=$1 and module='projects'", org); err != nil {
		return err
	}
	return LinkLockedRoles(ctx, tx, org, seal)
}
