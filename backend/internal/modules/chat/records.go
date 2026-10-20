package chat

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/e2ee"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

// A collection is a kind of encrypted record that lives in a container: a note, a task, a milestone ...
// The server enforces who may write it and that its few plain `meta` fields have the right shape; the content stays sealed.
type collection struct {
	kinds     []string          // container kinds it may live in
	create    [2]string         // module/right needed to add one ("" = any member)
	edit      [2]string         // module/right needed to change someone else's ("" = only the author)
	meta      map[string]string // allowed meta keys -> validator
	authorOwn bool              // authors may edit/delete their own (personal items, or with the own right below)
	own       [2]string         // module/right an author needs to change their own ("" = no right needed)
}

var collections = map[string]collection{
	"note":      {kinds: []string{"note", "sharednote"}, authorOwn: true, meta: map[string]string{"book": "uuid", "pinned": "bool"}},
	"notebook":  {kinds: []string{"note"}, authorOwn: true},
	"draft":     {kinds: []string{"note"}, authorOwn: true}, // the unsaved text in the Home page's "My notes" box
	"task":      {kinds: []string{"project"}, create: [2]string{"projects", "Edit tasks"}, edit: [2]string{"projects", "Move tasks"}, authorOwn: true, own: [2]string{"projects", "Edit own tasks"}, meta: map[string]string{"column": "enum:backlog,todo,doing,review,done", "due": "date", "assignee": "uuid", "sprint": "str", "priority": "enum:low,medium,high"}},
	"milestone": {kinds: []string{"project"}, create: [2]string{"projects", "Manage milestones"}, edit: [2]string{"projects", "Manage milestones"}, meta: map[string]string{"due": "date", "closed": "bool"}},
	"sprint":    {kinds: []string{"project"}, create: [2]string{"projects", "Manage milestones"}, edit: [2]string{"projects", "Manage milestones"}, meta: map[string]string{"start": "date", "end": "date", "milestone": "uuid"}},
	"risk":      {kinds: []string{"project"}, create: [2]string{"projects", "Manage team"}, edit: [2]string{"projects", "Manage team"}, meta: map[string]string{"level": "enum:Low,Medium,High"}},
	"activity":  {kinds: []string{"project"}, authorOwn: true},
	"standup":   {kinds: []string{"project"}, create: [2]string{"projects", "View"}, authorOwn: true, meta: map[string]string{"day": "date"}},
	"rsvp":      {kinds: []string{"event"}, create: [2]string{"calendar", "View"}, authorOwn: true, meta: map[string]string{"going": "bool"}},
	"comment":   {kinds: []string{"project", "event", "folder"}, authorOwn: true},
	"entry":     {kinds: []string{"event"}, create: [2]string{"calendar", "Add events"}, edit: [2]string{"calendar", "Schedule meetings"}, authorOwn: true, own: [2]string{"calendar", "Edit own events"}, meta: map[string]string{"date": "date", "start": "str", "end": "str", "type": "enum:Meeting,Standup,Task,Event,Personal,Reminder", "repeat": "str"}},
}

// mayChange: your own record needs the "own" right where the collection has one (a task, a calendar entry); the general edit
// right covers anyone's. Personal items (notes, your RSVP) are always yours to change.
// sharedNoteEditor: in a shared note (a container made when a note is shared) the owner and the people given edit rights may
// change the note; everyone else in it can only read. It returns nil when the container is not a shared note.
func sharedNoteEditor(ctx context.Context, tx pgx.Tx, p *tenant.Principal, container uuid.UUID) (*bool, error) {
	var kind, role string
	err := tx.QueryRow(ctx, `select c.kind, coalesce(m.role,'') from hub.conversations c left join hub.conversation_members m on m.conversation_id=c.id and m.account_id=$2 and m.left_at is null where c.id=$1`, container, p.Account).Scan(&kind, &role)
	if err != nil || kind != "sharednote" {
		return nil, nil
	}
	ok := role == "owner" || role == "admin"
	return &ok, nil
}

func mayChange(ctx context.Context, tx pgx.Tx, p *tenant.Principal, col collection, author, container uuid.UUID) error {
	if ok, err := sharedNoteEditor(ctx, tx, p, container); err != nil || ok != nil {
		if err != nil {
			return err
		}
		if *ok {
			return nil
		}
		return httpx.ErrForbidden
	}
	if author == p.Account && col.authorOwn {
		if col.own[0] == "" {
			return nil
		}
		return authz.RequireOwn(ctx, tx, p.Person, col.edit[0], col.edit[1], col.own[1], true)
	}
	if col.edit[0] == "" {
		return httpx.ErrForbidden
	}
	return authz.Require(ctx, tx, p.Person, col.edit[0], col.edit[1])
}

var dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func validMeta(spec map[string]string, meta map[string]any) error {
	if len(meta) > 12 {
		return httpx.Bad("Too many fields.")
	}
	for k, v := range meta {
		kind, ok := spec[k]
		if !ok {
			return httpx.Bad("Unknown field '" + k + "'.")
		}
		switch {
		case kind == "bool":
			if _, ok := v.(bool); !ok {
				return httpx.Bad("'" + k + "' must be true or false.")
			}
		case kind == "date":
			s, ok := v.(string)
			if _, err := time.Parse("2006-01-02", s); !ok || err != nil || !dateRe.MatchString(s) {
				return httpx.Bad("'" + k + "' must be a date like 2026-10-31.")
			}
		case kind == "uuid":
			s, ok := v.(string)
			if _, err := uuid.Parse(s); !ok || err != nil {
				return httpx.Bad("'" + k + "' must be an id.")
			}
		case kind == "str":
			if s, ok := v.(string); !ok || len(s) > 40 {
				return httpx.Bad("'" + k + "' must be short text.")
			}
		case strings.HasPrefix(kind, "enum:"):
			s, _ := v.(string)
			okv := false
			for _, o := range strings.Split(kind[5:], ",") {
				okv = okv || o == s
			}
			if !okv {
				return httpx.Bad("'" + k + "' has an unknown value.")
			}
		}
	}
	return nil
}

type Record struct {
	ID           uuid.UUID      `json:"id"`
	Collection   string         `json:"collection"`
	Container    uuid.UUID      `json:"container"`
	Parent       *uuid.UUID     `json:"parent,omitempty"`
	Author       uuid.UUID      `json:"author_account"`
	AuthorName   string         `json:"author_name,omitempty"`
	AuthorPerson *uuid.UUID     `json:"author_person,omitempty"`
	Device       *uuid.UUID     `json:"author_device,omitempty"`
	Epoch        int            `json:"epoch"`
	Envelope     string         `json:"envelope,omitempty"`
	Meta         map[string]any `json:"meta"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Deleted      bool           `json:"deleted,omitempty"`
}

type RecordIn struct {
	ID         uuid.UUID      `json:"id"`
	Collection string         `json:"collection"`
	Parent     *uuid.UUID     `json:"parent"`
	Epoch      int            `json:"epoch"`
	Device     uuid.UUID      `json:"device"`
	Envelope   string         `json:"envelope"`
	Meta       map[string]any `json:"meta"`
}

func (s *Service) checkRecord(ctx context.Context, tx pgx.Tx, p *tenant.Principal, container uuid.UUID, in RecordIn) (collection, int, error) {
	col, ok := collections[in.Collection]
	if !ok {
		return col, 0, httpx.Bad("Unknown kind of record.")
	}
	kind, _, _, epoch, err := s.lockConvNoLock(ctx, tx, p, container)
	if err != nil {
		return col, 0, err
	}
	okKind := false
	for _, k := range col.kinds {
		okKind = okKind || k == kind
	}
	if !okKind {
		return col, 0, httpx.Bad("That cannot be stored here.")
	}
	env, err := base64.StdEncoding.DecodeString(in.Envelope)
	if err != nil || e2ee.ValidateEnvelope(env) != nil {
		return col, 0, httpx.Bad("That is not a valid encrypted record.")
	}
	if in.Epoch != epoch {
		return col, 0, httpx.Err(409, "epoch_changed", "The key changed. Fetch the new key and save again.")
	}
	if err := s.myDevice(ctx, tx, p.Account, in.Device); err != nil {
		return col, 0, err
	}
	return col, epoch, validMeta(col.meta, in.Meta)
}

// SaveRecord creates (POST) or replaces (PUT) one record. Replacing keeps its id and bumps its version.
func (s *Service) SaveRecord(ctx context.Context, p *tenant.Principal, container uuid.UUID, in RecordIn, update bool) (*Record, error) {
	if in.ID == uuid.Nil || in.ID.Version() != 7 {
		return nil, httpx.Bad("Records need a time-ordered id (UUID v7) made by the sender.")
	}
	if in.Meta == nil {
		in.Meta = map[string]any{}
	}
	var out *Record
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		col, epoch, err := s.checkRecord(ctx, tx, p, container, in)
		if err != nil {
			return err
		}
		env, _ := base64.StdEncoding.DecodeString(in.Envelope)
		meta, _ := json.Marshal(in.Meta)
		if !update {
			if ok, err := sharedNoteEditor(ctx, tx, p, container); err != nil {
				return err
			} else if ok != nil && !*ok {
				return httpx.ErrForbidden
			}
			if col.create[0] != "" {
				if err := authz.Require(ctx, tx, p.Person, col.create[0], col.create[1]); err != nil {
					return err
				}
			}
			if in.Parent != nil {
				var ok bool
				if err := tx.QueryRow(ctx, `select exists (select 1 from hub.records where org_id=$1 and id=$2 and container_id=$3)`, p.Org, *in.Parent, container).Scan(&ok); err != nil {
					return err
				}
				if !ok {
					return httpx.Bad("The record you are answering is not here.")
				}
			}
			if _, err := tx.Exec(ctx, `insert into hub.records (org_id, id, collection, container_id, parent_id, author_account, author_person, author_device, epoch, ciphertext, meta)
				values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, p.Org, in.ID, in.Collection, container, in.Parent, p.Account, p.Person, in.Device, epoch, env, meta); err != nil {
				return err
			}
		} else {
			var author uuid.UUID
			var coll string
			err := tx.QueryRow(ctx, `select author_account, collection from hub.records where org_id=$1 and id=$2 and container_id=$3 and deleted_at is null for update`, p.Org, in.ID, container).Scan(&author, &coll)
			if errors.Is(err, pgx.ErrNoRows) || err == nil && coll != in.Collection {
				return httpx.ErrNotFound
			}
			if err != nil {
				return err
			}
			if err := mayChange(ctx, tx, p, col, author, container); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `update hub.records set ciphertext=$3, meta=$4, epoch=$5, author_device=$6, updated_at=now(), `+integrity.Bump+` where org_id=$1 and id=$2`, p.Org, in.ID, env, meta, epoch, in.Device); err != nil {
				return err
			}
		}
		if err := s.A.Sign.Seal(ctx, tx, "records", "id=$1 and org_id=$2", in.ID, p.Org); err != nil {
			return err
		}
		pid, dev := p.Person, in.Device
		out = &Record{ID: in.ID, Collection: in.Collection, Container: container, Parent: in.Parent, Author: p.Account, AuthorName: p.Name, Device: &dev, Epoch: epoch, Envelope: in.Envelope, Meta: in.Meta, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
		_ = pid
		return nil
	})
	if err != nil {
		if isDup(err) {
			return nil, httpx.Err(409, "duplicate", "That record already exists.")
		}
		return nil, err
	}
	d, _ := json.Marshal(out)
	s.A.Publish(ctx, realtime.Event{Type: "record.saved", Channel: realtime.ConvChannel(container), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
	return out, nil
}

func (s *Service) DeleteRecord(ctx context.Context, p *tenant.Principal, container, id uuid.UUID) error {
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if _, _, _, _, err := s.lockConvNoLock(ctx, tx, p, container); err != nil {
			return err
		}
		var author uuid.UUID
		var coll string
		err := tx.QueryRow(ctx, `select author_account, collection from hub.records where org_id=$1 and id=$2 and container_id=$3 and deleted_at is null for update`, p.Org, id, container).Scan(&author, &coll)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		col := collections[coll]
		if err := mayChange(ctx, tx, p, col, author, container); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update hub.records set ciphertext='\x00'::bytea, deleted_at=now(), updated_at=now(), `+integrity.Bump+` where org_id=$1 and id=$2`, p.Org, id); err != nil {
			return err
		}
		return s.A.Sign.Seal(ctx, tx, "records", "id=$1 and org_id=$2", id, p.Org)
	})
	if err == nil {
		d, _ := json.Marshal(map[string]any{"id": id, "container": container})
		s.A.Publish(ctx, realtime.Event{Type: "record.deleted", Channel: realtime.ConvChannel(container), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
	}
	return err
}

// Records lists a container's records of one collection, oldest changes last, with `since` for catching up after being offline.
func (s *Service) Records(ctx context.Context, p *tenant.Principal, container uuid.UUID, coll string, since *time.Time, limit int) ([]Record, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	out := []Record{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if _, _, _, _, err := s.lockConvNoLock(ctx, tx, p, container); err != nil {
			return err
		}
		var first int
		if err := tx.QueryRow(ctx, `select first_epoch from hub.conversation_members where conversation_id=$1 and account_id=$2`, container, p.Account).Scan(&first); err != nil {
			return err
		}
		q := `select r.id, r.collection, r.parent_id, r.author_account, coalesce(pe.first_name || ' ' || pe.last_name,''), r.author_person, r.author_device, r.epoch, r.ciphertext, r.meta, r.created_at, r.updated_at, r.deleted_at is not null
			from hub.records r left join hub.people pe on pe.id = r.author_person
			where r.org_id=$1 and r.container_id=$2 and r.epoch >= $3`
		args := []any{p.Org, container, first}
		if coll != "" {
			args = append(args, coll)
			q += ` and r.collection = $` + strconv.Itoa(len(args))
		}
		if since != nil {
			args = append(args, *since)
			q += ` and r.updated_at > $` + strconv.Itoa(len(args))
		}
		rs, err := tx.Query(ctx, q+` order by r.updated_at, r.id limit `+strconv.Itoa(limit), args...)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			r := Record{Container: container}
			var ct, meta []byte
			if err := rs.Scan(&r.ID, &r.Collection, &r.Parent, &r.Author, &r.AuthorName, &r.AuthorPerson, &r.Device, &r.Epoch, &ct, &meta, &r.CreatedAt, &r.UpdatedAt, &r.Deleted); err != nil {
				return err
			}
			_ = json.Unmarshal(meta, &r.Meta)
			if !r.Deleted {
				r.Envelope = base64.StdEncoding.EncodeToString(ct)
			}
			out = append(out, r)
		}
		return rs.Err()
	})
	return out, err
}

// Calendar returns the calendar entries I can see between two dates (inclusive), across every event I am part of. Only the
// date and time are plain; titles and notes stay sealed.
func (s *Service) Calendar(ctx context.Context, p *tenant.Principal, from, to string) ([]Record, error) {
	if !dateRe.MatchString(from) || !dateRe.MatchString(to) || to < from {
		return nil, httpx.Bad("Give a from and a to date like 2026-10-01.")
	}
	if d1, _ := time.Parse("2006-01-02", from); d1.AddDate(0, 0, 100).Before(mustDate(to)) {
		return nil, httpx.Bad("Ask for 100 days or fewer at a time.")
	}
	out := []Record{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "calendar", "View"); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select r.id, r.container_id, r.author_account, coalesce(pe.first_name || ' ' || pe.last_name,''), r.author_person, r.author_device, r.epoch, r.ciphertext, r.meta, r.created_at, r.updated_at
			from hub.records r
			join hub.conversation_members m on m.conversation_id = r.container_id and m.account_id = $2 and m.left_at is null and m.first_epoch <= r.epoch
			left join hub.people pe on pe.id = r.author_person
			where r.org_id = $1 and r.collection = 'entry' and r.deleted_at is null and (r.meta->>'date' between $3 and $4 or (coalesce(r.meta->>'repeat','None') <> 'None' and r.meta->>'date' <= $4))
			order by r.meta->>'date', r.meta->>'start', r.id limit 2000`, p.Org, p.Account, from, to)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			r := Record{Collection: "entry"}
			var ct, meta []byte
			if err := rs.Scan(&r.ID, &r.Container, &r.Author, &r.AuthorName, &r.AuthorPerson, &r.Device, &r.Epoch, &ct, &meta, &r.CreatedAt, &r.UpdatedAt); err != nil {
				return err
			}
			_ = json.Unmarshal(meta, &r.Meta)
			r.Envelope = base64.StdEncoding.EncodeToString(ct)
			out = append(out, r)
		}
		return rs.Err()
	})
	return out, err
}

func mustDate(s string) time.Time { t, _ := time.Parse("2006-01-02", s); return t }
