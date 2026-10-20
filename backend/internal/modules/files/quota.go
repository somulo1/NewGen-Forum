package files

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/orgsettings"
	"hub/internal/tenant"
)

// Storage limits. Everything a person uploads (Files library, chat and forum attachments, wall photos, ticket evidence)
// counts towards one total. The organization sets a default in security settings; one person can be given a different
// amount. 0 means no limit.

const mb = int64(1) << 20

type Usage struct {
	UsedBytes int64 `json:"used_bytes"`
	QuotaMB   int64 `json:"quota_mb"` // 0 = no limit
	Custom    bool  `json:"custom"`   // this person has their own limit
}

// usage reads a person's limit and what they store now.
func (s *Service) usage(ctx context.Context, tx pgx.Tx, org, person uuid.UUID) (Usage, error) {
	var u Usage
	if err := tx.QueryRow(ctx, `select coalesce(sum(size), 0) from hub.files where uploader=$1 and status='ready'`, person).Scan(&u.UsedBytes); err != nil {
		return u, err
	}
	err := tx.QueryRow(ctx, `select quota_mb from hub.storage_quotas where person_id=$1`, person).Scan(&u.QuotaMB)
	if err == nil {
		u.Custom = true
		return u, nil
	}
	if err != pgx.ErrNoRows {
		return u, err
	}
	sec := orgsettings.DefaultSecurity()
	if err := orgsettings.Get(ctx, s.A, tx, org, "security", &sec); err != nil {
		return u, err
	}
	if sec.UserQuotaMB == "" {
		sec.UserQuotaMB = orgsettings.DefaultSecurity().UserQuotaMB
	}
	u.QuotaMB, _ = strconv.ParseInt(sec.UserQuotaMB, 10, 64)
	return u, nil
}

func sizeText(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	case b >= mb:
		return fmt.Sprintf("%.0f MB", float64(b)/float64(mb))
	default:
		return fmt.Sprintf("%d KB", (b+1023)/1024)
	}
}

// fits refuses an upload of `adding` bytes that would take the person past their limit.
func (s *Service) fits(ctx context.Context, tx pgx.Tx, org, person uuid.UUID, adding int64) error {
	u, err := s.usage(ctx, tx, org, person)
	if err != nil || u.QuotaMB == 0 {
		return err
	}
	limit := u.QuotaMB * mb
	if u.UsedBytes+adding > limit {
		left := limit - u.UsedBytes
		if left < 0 {
			left = 0
		}
		return httpx.Err(413, "quota_exceeded", fmt.Sprintf("This would go over your storage limit of %s (you use %s, %s left). Delete files you no longer need, or ask an administrator for more space.",
			sizeText(limit), sizeText(u.UsedBytes), sizeText(left)))
	}
	return nil
}

// lockPerson serializes one person's uploads, so two at once cannot both squeeze under the limit.
func lockPerson(ctx context.Context, tx pgx.Tx, person uuid.UUID) error {
	_, err := tx.Exec(ctx, `select pg_advisory_xact_lock(hashtextextended('storage:' || $1::text, 0))`, person)
	return err
}

func (s *Service) Usage(ctx context.Context, p *tenant.Principal) (Usage, error) {
	var u Usage
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var err error
		u, err = s.usage(ctx, tx, p.Org, p.Person)
		return err
	})
	return u, err
}

// MyFile is one of a person's stored files, biggest first, so they can see what to delete to free space. The name is
// sealed in the envelope; the person's device opens it.
type MyFile struct {
	ID       uuid.UUID `json:"id"`
	Folder   uuid.UUID `json:"folder"`
	Kind     string    `json:"kind"` // what holds it: folder, forum, dm, wall, ticket, project, event
	Size     int64     `json:"size"`
	Account  uuid.UUID `json:"author_account"`
	Device   uuid.UUID `json:"author_device"`
	Epoch    int       `json:"epoch"`
	Envelope string    `json:"envelope"`
	Created  time.Time `json:"created_at"`
}

func (s *Service) MyFiles(ctx context.Context, p *tenant.Principal, limit int) ([]MyFile, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	out := []MyFile{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select f.id, f.folder_id, c.kind, f.size, f.author_account, f.author_device, f.epoch, f.envelope, f.created_at
			from hub.files f join hub.conversations c on c.id = f.folder_id
			where f.uploader=$1 and f.status='ready' order by f.size desc, f.id limit `+strconv.Itoa(limit), p.Person)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var f MyFile
			var env []byte
			if err := rs.Scan(&f.ID, &f.Folder, &f.Kind, &f.Size, &f.Account, &f.Device, &f.Epoch, &env, &f.Created); err != nil {
				return err
			}
			f.Envelope = base64.StdEncoding.EncodeToString(env)
			out = append(out, f)
		}
		return rs.Err()
	})
	return out, err
}

type PersonUsage struct {
	Person uuid.UUID `json:"person"`
	Name   string    `json:"name"`
	Usage
}

// TopUsage lists the people who store the most, and everyone with their own limit, for the administrator.
func (s *Service) TopUsage(ctx context.Context, p *tenant.Principal, limit int) ([]PersonUsage, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	out := []PersonUsage{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Edit organization profile"); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `with used as (select uploader, sum(size) b from hub.files where status='ready' group by uploader)
			select pe.id, pe.first_name || ' ' || pe.last_name, coalesce(u.b, 0), q.quota_mb
			  from hub.people pe left join used u on u.uploader = pe.id left join hub.storage_quotas q on q.person_id = pe.id
			 where u.b is not null or q.person_id is not null
			 order by coalesce(u.b, 0) desc, pe.id limit `+strconv.Itoa(limit))
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var r PersonUsage
			var q *int64
			if err := rs.Scan(&r.Person, &r.Name, &r.UsedBytes, &q); err != nil {
				return err
			}
			if q != nil {
				r.QuotaMB, r.Custom = *q, true
			}
			out = append(out, r)
		}
		return rs.Err()
	})
	if err != nil {
		return nil, err
	}
	// people on the default limit carry the organization's default
	var def Usage
	_ = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var err error
		def, err = s.usage(ctx, tx, p.Org, uuid.Nil)
		return err
	})
	for i := range out {
		if !out[i].Custom {
			out[i].QuotaMB = def.QuotaMB
		}
	}
	return out, nil
}

// SetPersonQuota gives one person their own limit (MB, 0 = no limit), or puts them back on the default (nil).
func (s *Service) SetPersonQuota(ctx context.Context, p *tenant.Principal, person uuid.UUID, quotaMB *int64) error {
	if quotaMB != nil && (*quotaMB < 0 || *quotaMB > 10<<20) {
		return httpx.Bad("Enter a limit between 0 (no limit) and 10 TB.")
	}
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Edit organization profile"); err != nil {
			return err
		}
		var name string
		if err := tx.QueryRow(ctx, `select first_name || ' ' || last_name from hub.people where id=$1`, person).Scan(&name); err != nil {
			return httpx.ErrNotFound
		}
		if quotaMB == nil {
			if _, err := tx.Exec(ctx, `delete from hub.storage_quotas where person_id=$1`, person); err != nil {
				return err
			}
			return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Put "+name+" back on the default storage limit", "Settings", p.IP, map[string]any{"person": person})
		}
		if _, err := tx.Exec(ctx, `insert into hub.storage_quotas (org_id, person_id, quota_mb) values ($1,$2,$3)
			on conflict (org_id, person_id) do update set quota_mb=excluded.quota_mb, updated_at=now(), `+integrity.BumpOn("storage_quotas"), p.Org, person, *quotaMB); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "storage_quotas", "person_id=$1", person); err != nil {
			return err
		}
		what := "no storage limit"
		if *quotaMB > 0 {
			what = "a storage limit of " + sizeText(*quotaMB*mb)
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Gave "+name+" "+what, "Settings", p.IP, map[string]any{"person": person, "quota_mb": *quotaMB})
	})
}

func (s *Service) quotaRoutes(r chi.Router, org func(func(http.ResponseWriter, *http.Request, *tenant.Principal) error) http.HandlerFunc) {
	r.Get("/storage/usage", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		u, err := s.Usage(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, u)
		return nil
	}))
	r.Get("/storage/files", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		n, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		l, err := s.MyFiles(r.Context(), p, n)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"files": l})
		return nil
	}))
	r.Get("/storage/people", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		n, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		l, err := s.TopUsage(r.Context(), p, n)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"people": l})
		return nil
	}))
	// {"quota_mb": 2048} for an own limit, {"quota_mb": null} to go back to the organization's default
	r.Put("/storage/people/{id}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		var in struct {
			QuotaMB *int64 `json:"quota_mb"`
		}
		if err := httpx.Decode(r, &in, 1<<10); err != nil {
			return err
		}
		if err := s.SetPersonQuota(r.Context(), p, id, in.QuotaMB); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
}
