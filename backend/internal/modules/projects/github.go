package projects

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/tenant"
)

// GitHub: a project may be linked to one repository. The link keeps the webhook's secret (shown once, to paste into GitHub) and,
// optionally, a read-only access token so Hub can list the repository's branches (needed for private repositories). Both are kept
// encrypted with the organization's key.

type GitHub struct {
	Linked   bool       `json:"linked"`
	Repo     string     `json:"repo,omitempty"`
	Hook     string     `json:"hook,omitempty"` // the webhook path: /v1/hooks/github/{hook}
	Auto     bool       `json:"auto"`           // move tasks without asking (forward moves only)
	HasToken bool       `json:"has_token"`
	Secret   string     `json:"secret,omitempty"` // only right after linking or making a new secret
	Events   []GitEvent `json:"events"`
	CanEdit  bool       `json:"can_edit"`
}

type GitEvent struct {
	ID      uuid.UUID  `json:"id"`
	Task    *uuid.UUID `json:"task,omitempty"`
	Kind    string     `json:"kind"`
	Branch  string     `json:"branch"`
	Title   string     `json:"title"`
	URL     string     `json:"url"`
	Actor   string     `json:"actor"`
	At      time.Time  `json:"at"`
	Suggest string     `json:"suggest,omitempty"` // the status it suggests
	State   string     `json:"state"`
}

var repoRe = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,100}/[A-Za-z0-9_.-]{1,100}$`)

func ghCtx(org, project uuid.UUID, col string) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "project_github", Column: col, RowID: project.String()}
}

// onTeam: the caller is on the project team (and so may see its GitHub activity).
func onTeam(ctx context.Context, tx pgx.Tx, p *tenant.Principal, project uuid.UUID) error {
	var ok bool
	if err := tx.QueryRow(ctx, `select exists (select 1 from hub.conversation_members where conversation_id=$1 and account_id=$2 and left_at is null)`, project, p.Account).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return httpx.ErrNotFound
	}
	return nil
}

// mayLink: the project's owner, or someone who manages project teams.
func mayLink(ctx context.Context, tx pgx.Tx, p *tenant.Principal, project uuid.UUID) (bool, error) {
	var owner uuid.UUID
	if err := tx.QueryRow(ctx, `select owner from hub.projects where conversation_id=$1`, project).Scan(&owner); err != nil {
		return false, httpx.ErrNotFound
	}
	if owner == p.Person {
		return true, nil
	}
	return authz.Can(ctx, tx, p.Person, "projects", "Manage team")
}

func (s *Service) GitHub(ctx context.Context, p *tenant.Principal, project uuid.UUID, task ...uuid.UUID) (*GitHub, error) {
	out := &GitHub{Events: []GitEvent{}}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := onTeam(ctx, tx, p, project); err != nil {
			return err
		}
		var err error
		if out.CanEdit, err = mayLink(ctx, tx, p, project); err != nil {
			return err
		}
		var hook uuid.UUID
		var tok []byte
		err = tx.QueryRow(ctx, `select repo, hook_id, auto, token_enc from hub.project_github where project_id=$1`, project).Scan(&out.Repo, &hook, &out.Auto, &tok)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		out.Linked, out.Hook, out.HasToken = true, "/v1/hooks/github/"+hook.String(), len(tok) > 0
		var only *uuid.UUID
		if len(task) > 0 {
			only = &task[0] // one task's activity
		}
		out.Events, err = gitEvents(ctx, tx, project, only, 50)
		return err
	})
	return out, err
}

func gitEvents(ctx context.Context, tx pgx.Tx, project uuid.UUID, task *uuid.UUID, limit int) ([]GitEvent, error) {
	q := `select id, task_id, kind, branch, title, url, actor, at, coalesce(suggest,''), state from hub.project_git_events where project_id=$1`
	args := []any{project}
	if task != nil {
		q += ` and task_id=$2`
		args = append(args, *task)
	}
	rs, err := tx.Query(ctx, q+fmt.Sprintf(` order by at desc limit %d`, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	out := []GitEvent{}
	for rs.Next() {
		var e GitEvent
		if err := rs.Scan(&e.ID, &e.Task, &e.Kind, &e.Branch, &e.Title, &e.URL, &e.Actor, &e.At, &e.Suggest, &e.State); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rs.Err()
}

type LinkIn struct {
	Repo      string  `json:"repo"`  // owner/name, or the repository's https://github.com/owner/name address
	Token     *string `json:"token"` // nil keeps it; "" removes it
	Auto      bool    `json:"auto"`
	NewSecret bool    `json:"new_secret"`
}

func cleanRepo(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimSuffix(strings.TrimSuffix(v, "/"), ".git")
	for _, pre := range []string{"https://github.com/", "http://github.com/", "github.com/", "git@github.com:"} {
		v = strings.TrimPrefix(v, pre)
	}
	return v
}

// Link links (or relinks) a project to a repository. A new link, or a request for a new secret, returns the secret once.
func (s *Service) Link(ctx context.Context, p *tenant.Principal, project uuid.UUID, in LinkIn) (*GitHub, error) {
	repo := cleanRepo(in.Repo)
	if !repoRe.MatchString(repo) {
		return nil, httpx.Bad("Give the repository as owner/name, for example acme/hub.")
	}
	if in.Token != nil && len(*in.Token) > 300 {
		return nil, httpx.Bad("That access token is too long.")
	}
	secret := ""
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := onTeam(ctx, tx, p, project); err != nil {
			return err
		}
		ok, err := mayLink(ctx, tx, p, project)
		if err != nil {
			return err
		}
		if !ok {
			return httpx.Err(403, "forbidden", "Only the project owner, or someone who manages project teams, links GitHub.")
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		var exists bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.project_github where project_id=$1)`, project).Scan(&exists); err != nil {
			return err
		}
		if !exists || in.NewSecret {
			b := make([]byte, 24)
			if _, err := rand.Read(b); err != nil {
				return err
			}
			secret = hex.EncodeToString(b)
		}
		var secEnc []byte
		if secret != "" {
			if secEnc, err = crypto.Field(dek, ghCtx(p.Org, project, "secret_enc"), []byte(secret)); err != nil {
				return err
			}
		}
		var tokEnc []byte
		if in.Token != nil && strings.TrimSpace(*in.Token) != "" {
			if tokEnc, err = crypto.Field(dek, ghCtx(p.Org, project, "token_enc"), []byte(strings.TrimSpace(*in.Token))); err != nil {
				return err
			}
		}
		if !exists {
			hook := app.NewID()
			if _, err := tx.Exec(ctx, `insert into hub.project_github (org_id, project_id, repo, hook_id, secret_enc, token_enc, auto) values ($1,$2,$3,$4,$5,$6,$7)`,
				p.Org, project, repo, hook, secEnc, tokEnc, in.Auto); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `insert into hub.github_hooks (hook_id, org_id, project_id) values ($1,$2,$3)`, hook, p.Org, project); err != nil {
				return err
			}
		} else if _, err := tx.Exec(ctx, `update hub.project_github set repo=$2, auto=$3, secret_enc=coalesce($4, secret_enc),
				token_enc = case when $5::boolean then $6 else token_enc end, updated_at=now() where project_id=$1`,
			project, repo, in.Auto, secEnc, in.Token != nil, tokEnc); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Linked a project to GitHub "+repo, "Projects", p.IP, map[string]any{"project": project, "repo": repo})
	})
	if err != nil {
		return nil, err
	}
	out, err := s.GitHub(ctx, p, project)
	if out != nil {
		out.Secret = secret
	}
	return out, err
}

func (s *Service) Unlink(ctx context.Context, p *tenant.Principal, project uuid.UUID) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := onTeam(ctx, tx, p, project); err != nil {
			return err
		}
		if ok, err := mayLink(ctx, tx, p, project); err != nil || !ok {
			if err == nil {
				err = httpx.ErrForbidden
			}
			return err
		}
		var hook uuid.UUID
		if err := tx.QueryRow(ctx, `delete from hub.project_github where project_id=$1 returning hook_id`, project).Scan(&hook); err != nil {
			return httpx.ErrNotFound
		}
		_, err := tx.Exec(ctx, `delete from hub.github_hooks where hook_id=$1`, hook)
		return err
	})
}

// Branches lists the repository's branches from GitHub (up to 300), with the access token when one is kept.
func (s *Service) Branches(ctx context.Context, p *tenant.Principal, project uuid.UUID) ([]string, error) {
	var repo string
	var token string
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := onTeam(ctx, tx, p, project); err != nil {
			return err
		}
		var tok []byte
		if err := tx.QueryRow(ctx, `select repo, token_enc from hub.project_github where project_id=$1`, project).Scan(&repo, &tok); err != nil {
			return httpx.Err(409, "not_linked", "Link this project to a GitHub repository first.")
		}
		if len(tok) > 0 {
			dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
			if err != nil {
				return err
			}
			if b, err := crypto.FieldOpen(dek, ghCtx(p.Org, project, "token_enc"), tok); err == nil {
				token = string(b)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := []string{}
	cl := &http.Client{Timeout: 10 * time.Second}
	for page := 1; page <= 3; page++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/repos/%s/branches?per_page=100&page=%d", githubAPI(), repo, page), nil)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("User-Agent", "hub-projects")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res, err := cl.Do(req)
		if err != nil {
			return nil, httpx.Err(502, "github_unreachable", "GitHub could not be reached. Try again in a moment.")
		}
		var list []struct {
			Name string `json:"name"`
		}
		ok := res.StatusCode == 200 && json.NewDecoder(http.MaxBytesReader(nil, res.Body, 2<<20)).Decode(&list) == nil
		res.Body.Close()
		if !ok {
			if res.StatusCode == 404 || res.StatusCode == 401 || res.StatusCode == 403 {
				return nil, httpx.Err(409, "github_denied", "GitHub did not show this repository's branches. For a private repository, add an access token that can read it.")
			}
			return nil, httpx.Err(502, "github_error", "GitHub did not answer as expected. Try again in a moment.")
		}
		for _, b := range list {
			out = append(out, b.Name)
		}
		if len(list) < 100 {
			break
		}
	}
	return out, nil
}

// githubAPI is GitHub's API address; HUB_GITHUB_API points it elsewhere (GitHub Enterprise, or tests).
func githubAPI() string {
	if v := strings.TrimRight(os.Getenv("HUB_GITHUB_API"), "/"); v != "" {
		return v
	}
	return "https://api.github.com"
}
