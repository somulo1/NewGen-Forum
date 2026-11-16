package projects

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

// A GitHub delivery: checked against the project's secret, then each branch or pull request it names is matched to tasks, and
// what happened is recorded with the status it suggests:
//   a branch is made or pushed to   -> the first "in progress" status
//   a pull request is opened         -> the first "being checked" status
//   a pull request is merged         -> the first "finished" status
// Only forward moves are suggested. When the project moves tasks automatically, an allowed move is made at once (and shown as
// done by GitHub); otherwise it waits for someone on the team to accept it after checking it is right.

type ghEvent struct {
	kind, branch, title, url, actor string
	cat                             int // the workflow category it suggests (-1: none)
}

// HandleHook takes one webhook delivery. It answers 404 for an unknown hook and 401 for a wrong signature, so nothing is recorded
// unless GitHub (holding the secret) sent it.
func (s *Service) HandleHook(ctx context.Context, hook uuid.UUID, event, delivery, signature, contentType string, body []byte) error {
	var org, project uuid.UUID
	if err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select org_id, project_id from hub.github_hooks where hook_id=$1`, hook).Scan(&org, &project)
	}); err != nil {
		return httpx.ErrNotFound
	}
	// the form-encoded kind of delivery carries the JSON in its "payload" field
	payload := body
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		v, err := url.ParseQuery(string(body))
		if err != nil {
			return httpx.Bad("That delivery could not be read.")
		}
		payload = []byte(v.Get("payload"))
	}
	var changed []uuid.UUID
	err := s.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		var secEnc []byte
		var auto bool
		if err := tx.QueryRow(ctx, `select secret_enc, auto from hub.project_github where project_id=$1`, project).Scan(&secEnc, &auto); err != nil {
			return httpx.ErrNotFound
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, org)
		if err != nil {
			return err
		}
		secret, err := crypto.FieldOpen(dek, ghCtx(org, project, "secret_enc"), secEnc)
		if err != nil {
			return err
		}
		mac := hmac.New(sha256.New, secret)
		mac.Write(body)
		want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(want), []byte(signature)) {
			return httpx.Err(401, "bad_signature", "The delivery is not signed with this project's secret.")
		}
		evs := parseEvent(event, payload)
		if len(evs) == 0 {
			return nil // a ping, or something Hub does not track
		}
		pr, err := scan(tx.QueryRow(ctx, sel+` where p.conversation_id=$1`, project))
		if err != nil {
			return err
		}
		tasks, err := loadItems(ctx, tx, project, "task")
		if err != nil {
			return err
		}
		wf := workflowOf(pr)
		for _, e := range evs {
			matched := matchTasks(pr.Code, tasks, e.branch, e.title)
			if len(matched) == 0 {
				if _, err := tx.Exec(ctx, `insert into hub.project_git_events (org_id, project_id, delivery, kind, branch, title, url, actor) values ($1,$2,$3,$4,$5,$6,$7,$8) on conflict do nothing`,
					org, project, nullable(delivery), e.kind, e.branch, e.title, e.url, e.actor); err != nil {
					return err
				}
				continue
			}
			for _, t := range matched {
				target := suggestFor(wf, t.data, e.cat)
				state := "none"
				if target != nil {
					state = "pending"
					if auto && moveAllowed(wf, text(t.data, "status"), target.Key) {
						old, _ := setStatus(t, *target)
						if err := s.writeBySystemAs(ctx, tx, org, project, t, "status", old, target.Key, "GitHub"); err != nil {
							return err
						}
						if p := text(t.data, "parent"); p != "" {
							if err := s.rollUp(ctx, tx, org, pr, p); err != nil {
								return err
							}
						}
						state = "auto"
						changed = append(changed, t.id)
					}
				}
				var sug *string
				if target != nil {
					sug = &target.Key
				}
				if _, err := tx.Exec(ctx, `insert into hub.project_git_events (org_id, project_id, task_id, delivery, kind, branch, title, url, actor, suggest, state) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) on conflict do nothing`,
					org, project, t.id, nullable(delivery), e.kind, e.branch, e.title, e.url, e.actor, sug, state); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err == nil {
		d, _ := json.Marshal(map[string]any{"project": project, "github": true, "tasks": changed})
		s.A.Publish(ctx, realtime.Event{Type: "project.plan", Channel: realtime.ConvChannel(project), Org: org.String(), Data: d})
	}
	return err
}

func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// parseEvent reads the deliveries Hub tracks: branch made or removed, pushes, and pull requests.
func parseEvent(event string, payload []byte) []ghEvent {
	var p struct {
		Ref     string `json:"ref"`
		RefType string `json:"ref_type"`
		Deleted bool   `json:"deleted"`
		Compare string `json:"compare"`
		Commits []struct {
			Message string `json:"message"`
		} `json:"commits"`
		Pusher struct {
			Name string `json:"name"`
		} `json:"pusher"`
		Sender struct {
			Login string `json:"login"`
		} `json:"sender"`
		Action      string `json:"action"`
		PullRequest struct {
			Title   string `json:"title"`
			HTMLURL string `json:"html_url"`
			Merged  bool   `json:"merged"`
			Draft   bool   `json:"draft"`
			Head    struct {
				Ref string `json:"ref"`
			} `json:"head"`
		} `json:"pull_request"`
	}
	if json.Unmarshal(payload, &p) != nil {
		return nil
	}
	who := p.Sender.Login
	switch event {
	case "create":
		if p.RefType == "branch" {
			return []ghEvent{{kind: "branch_created", branch: p.Ref, title: "Branch created", actor: who, cat: 1}}
		}
	case "delete":
		if p.RefType == "branch" {
			return []ghEvent{{kind: "branch_deleted", branch: p.Ref, title: "Branch deleted", actor: who, cat: -1}}
		}
	case "push":
		if !strings.HasPrefix(p.Ref, "refs/heads/") || p.Deleted {
			return nil
		}
		msg := strconv.Itoa(len(p.Commits)) + " commit(s)"
		if n := len(p.Commits); n > 0 {
			msg += ": " + firstLine(p.Commits[n-1].Message)
		}
		if who == "" {
			who = p.Pusher.Name
		}
		return []ghEvent{{kind: "push", branch: strings.TrimPrefix(p.Ref, "refs/heads/"), title: clip(msg, 200), url: p.Compare, actor: who, cat: 1}}
	case "pull_request":
		e := ghEvent{branch: p.PullRequest.Head.Ref, title: clip(p.PullRequest.Title, 200), url: p.PullRequest.HTMLURL, actor: who, cat: -1}
		switch {
		case p.Action == "closed" && p.PullRequest.Merged:
			e.kind, e.cat = "pr_merged", 3
		case p.Action == "closed":
			e.kind = "pr_closed"
		case (p.Action == "opened" || p.Action == "reopened" || p.Action == "ready_for_review") && !p.PullRequest.Draft:
			e.kind, e.cat = "pr_opened", 2
		case p.Action == "opened" || p.Action == "reopened":
			e.kind, e.cat = "pr_draft", 1
		default:
			return nil
		}
		return []ghEvent{e}
	}
	return nil
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func clip(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// matchTasks finds the tasks a branch belongs to: the task whose stored branch it is, else the task whose key (CODE-12) it names.
func matchTasks(code string, tasks []planRow, branch, title string) []planRow {
	if branch == "" {
		return nil
	}
	var out []planRow
	for _, t := range tasks {
		if strings.EqualFold(text(t.data, "branch"), branch) {
			out = append(out, t)
		}
	}
	if len(out) > 0 {
		return out
	}
	// the task key is the project code (which may itself hold a dash, like PRJ-01) and the task's number
	keyIn := regexp.MustCompile(`(?i)(?:^|[^a-z0-9])` + regexp.QuoteMeta(code) + `-(\d+)(?:[^0-9]|$)`)
	for _, src := range []string{branch, title} {
		for _, m := range keyIn.FindAllStringSubmatch(src, -1) {
			n, _ := strconv.Atoi(m[1])
			for _, t := range tasks {
				if t.ref != nil && *t.ref == n {
					return []planRow{t}
				}
			}
		}
	}
	return nil
}

// suggestFor: the first status of the wanted kind, when it is a step forward for the task.
func suggestFor(wf []Status, data map[string]any, cat int) *Status {
	if cat < 0 {
		return nil
	}
	cur := -1
	for _, st := range wf {
		if st.Key == text(data, "status") {
			cur = st.Cat
		}
	}
	if cur >= cat {
		return nil
	}
	for i := range wf {
		if wf[i].Cat == cat {
			return &wf[i]
		}
	}
	return nil
}

// moveAllowed follows the workflow's allowed moves (an empty list allows any).
func moveAllowed(wf []Status, from, to string) bool {
	for _, st := range wf {
		if st.Key == from {
			return len(st.To) == 0 || contains(st.To, to)
		}
	}
	return true
}

// writeBySystemAs is writeBySystem, naming who made the change (GitHub).
func (s *Service) writeBySystemAs(ctx context.Context, tx pgx.Tx, org, project uuid.UUID, r planRow, field string, old, next any, who string) error {
	if err := s.writeBySystem(ctx, tx, org, project, r, field, old, next); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `update hub.project_item_history set who=$3 where id = (select id from hub.project_item_history where item_id=$1 and project_id=$2 order by at desc, id desc limit 1)`, r.id, project, who)
	return err
}

// Decide records what someone on the team did with a suggestion: "accepted" (they moved the task themselves, through the usual
// rules) or "dismissed".
func (s *Service) Decide(ctx context.Context, p *tenant.Principal, project, event uuid.UUID, decision string) error {
	if decision != "accepted" && decision != "dismissed" {
		return httpx.Bad("Accept or dismiss it.")
	}
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := onTeam(ctx, tx, p, project); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `update hub.project_git_events set state=$3, decided_by=$4, decided_at=now() where id=$1 and project_id=$2 and state='pending'`, event, project, decision, p.Person)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.Err(409, "already_decided", "That suggestion was already dealt with.")
		}
		return nil
	})
}
