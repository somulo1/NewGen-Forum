package integrity

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Problem is one thing the scanner found wrong.
type Problem struct {
	Table  string `json:"table"`
	Org    string `json:"org,omitempty"`
	Row    string `json:"row,omitempty"`
	Reason string `json:"reason"`
}

type Report struct {
	Tables   int       `json:"tables"`
	Rows     int64     `json:"rows"`
	Orgs     int       `json:"orgs"`
	Ledgers  int       `json:"ledgers"`
	Problems []Problem `json:"problems"`
}

func (r *Report) OK() bool { return len(r.Problems) == 0 }

// Scanner re-checks the database from the outside. It reads with a role that cannot write, holds the MAC key (which the
// database never has), and so can tell any row that was changed without the system apart from one the system wrote.
type Scanner struct {
	S      *Signer
	L      *Ledger
	MaxBad int // stop reporting after this many problems (default 200)
	// OnlyOrgs limits the per-organization checks (tests and targeted audits); nil scans everyone.
	OnlyOrgs []uuid.UUID
}

var pendingMAC = func() []byte {
	b := make([]byte, 32)
	for i := range b {
		b[i] = 0xff
	}
	return b
}()

// Run scans every sealed table for every organization, and every organization's ledger chain.
func (sc *Scanner) Run(ctx context.Context, conn *pgx.Conn, only string) (*Report, error) {
	if sc.MaxBad == 0 {
		sc.MaxBad = 200
	}
	rep := &Report{}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly, IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// which tables are sealed, and which of those are per organization
	rs, err := tx.Query(ctx, `select c.relname, exists (select 1 from pg_attribute a where a.attrelid = c.oid and a.attname = 'org_id' and not a.attisdropped)
		from pg_class c join pg_namespace n on n.oid = c.relnamespace
		where n.nspname = 'hub' and c.relkind in ('r','p') and not c.relispartition
		  and exists (select 1 from pg_trigger t where t.tgrelid = c.oid and t.tgname like '%\_mac' and not t.tgisinternal)
		order by c.relname`)
	if err != nil {
		return nil, err
	}
	type tbl struct {
		name   string
		tenant bool
	}
	var tables []tbl
	for rs.Next() {
		var t tbl
		if err := rs.Scan(&t.name, &t.tenant); err != nil {
			rs.Close()
			return nil, err
		}
		if only == "" || only == t.name {
			tables = append(tables, t)
		}
	}
	rs.Close()
	rep.Tables = len(tables)

	var orgs []uuid.UUID
	ors, err := tx.Query(ctx, `select id from hub.orgs order by id`)
	if err != nil {
		return nil, err
	}
	for ors.Next() {
		var id uuid.UUID
		if err := ors.Scan(&id); err != nil {
			ors.Close()
			return nil, err
		}
		orgs = append(orgs, id)
	}
	ors.Close()
	if len(sc.OnlyOrgs) > 0 {
		keep := map[uuid.UUID]bool{}
		for _, o := range sc.OnlyOrgs {
			keep[o] = true
		}
		filtered := orgs[:0]
		for _, o := range orgs {
			if keep[o] {
				filtered = append(filtered, o)
			}
		}
		orgs = filtered
	}
	rep.Orgs = len(orgs)

	scanTable := func(t tbl, org uuid.UUID) error {
		rows, err := tx.Query(ctx, fmt.Sprintf(`select ver, mac, hub.canon(t) from hub.%s t`, t.name))
		if err != nil {
			return fmt.Errorf("scan %s: %w", t.name, err)
		}
		defer rows.Close()
		for rows.Next() {
			var ver int
			var mac []byte
			var canon string
			if err := rows.Scan(&ver, &mac, &canon); err != nil {
				return err
			}
			rep.Rows++
			bad := ""
			switch {
			case string(mac) == string(pendingMAC):
				bad = "row was never sealed (its MAC is the pending marker)"
			case !sc.S.Verify(mac, t.name, ver, canon):
				bad = "row does not match its MAC (changed outside the system, or rolled back)"
			}
			if bad != "" && len(rep.Problems) < sc.MaxBad {
				row := strings.TrimSpace(canon)
				if len(row) > 160 {
					row = row[:160] + "…"
				}
				p := Problem{Table: t.name, Row: row, Reason: bad}
				if t.tenant {
					p.Org = org.String()
				}
				rep.Problems = append(rep.Problems, p)
			}
		}
		return rows.Err()
	}
	for _, t := range tables {
		if !t.tenant {
			if err := scanTable(t, uuid.Nil); err != nil {
				return nil, err
			}
			continue
		}
		for _, o := range orgs {
			if _, err := tx.Exec(ctx, `select set_config('hub.org_id', $1, true)`, o.String()); err != nil {
				return nil, err
			}
			if err := scanTable(t, o); err != nil {
				return nil, err
			}
		}
	}
	if only == "" || only == "ledger" {
		for _, o := range orgs {
			if _, err := tx.Exec(ctx, `select set_config('hub.org_id', $1, true)`, o.String()); err != nil {
				return nil, err
			}
			rep.Ledgers++
			if _, err := sc.L.Verify(ctx, tx, o); err != nil {
				rep.Problems = append(rep.Problems, Problem{Table: "ledger", Org: o.String(), Reason: err.Error()})
			}
		}
	}
	return rep, nil
}
