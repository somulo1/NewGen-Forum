package integrity

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
	// Cause: "schema" (the record is authentic: a database update added a column after it was sealed), "changed" (its
	// content no longer matches its seal), or "unsealed" (it never got its seal).
	Cause string            `json:"cause,omitempty"`
	Added []string          `json:"added_columns,omitempty"` // for "schema": the column(s) that appeared since it was sealed
	PK    map[string]string `json:"key,omitempty"`           // the record's primary key
	Ident map[string]string `json:"record,omitempty"`        // what identifies it for a person: reference, type, status, dates...
}

// identKeys are the columns that say which record this is, without exposing anything sealed or personal.
var identKeys = []string{"id", "ref", "key", "name", "type", "content_type", "kind", "module", "collection", "status", "role", "flow_key", "created_at", "ts", "conversation_id", "container_id", "conversation", "person_id", "level", "seq"}

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

// findAdded tries leaving out one or two of the row's empty columns: if the seal then matches, those columns were added to the
// table after the record was sealed (a new empty column changes the record's canonical form, not its content).
func (sc *Scanner) findAdded(ctx context.Context, tx pgx.Tx, table, ctid string, nullKeys []string, ver int, mac []byte) []string {
	if len(nullKeys) == 0 || len(nullKeys) > 30 {
		return nil
	}
	rs, err := tx.Query(ctx, fmt.Sprintf(`select array_to_string(c.ks, ','), hub.canon_minus(t, c.ks)
		from hub.%s t, (select array[a] ks from unnest($2::text[]) a union all select array[a, b] from unnest($2::text[]) a, unnest($2::text[]) b where a < b) c
		where t.ctid = $1::tid`, table), ctid, nullKeys)
	if err != nil {
		return nil
	}
	defer rs.Close()
	for rs.Next() {
		var cols, canon string
		if rs.Scan(&cols, &canon) == nil && sc.S.Verify(mac, table, ver, canon) {
			return strings.Split(cols, ",")
		}
	}
	return nil
}

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
		rows, err := tx.Query(ctx, fmt.Sprintf(`select ver, mac, hub.canon(t), ctid::text,
			coalesce((select array_agg(e.key) from jsonb_each(to_jsonb(t)) e where e.value = 'null'::jsonb and e.key not in ('mac','ver','updated_at')), '{}'),
			(to_jsonb(t) - 'mac' - 'ver' - 'updated_at')::text from hub.%s t`, t.name))
		if err != nil {
			return fmt.Errorf("scan %s: %w", t.name, err)
		}
		type suspect struct {
			ver      int
			mac      []byte
			canon    string
			ctid     string
			nullKeys []string
			data     string
			unsealed bool
		}
		var sus []suspect
		for rows.Next() {
			var x suspect
			if err := rows.Scan(&x.ver, &x.mac, &x.canon, &x.ctid, &x.nullKeys, &x.data); err != nil {
				rows.Close()
				return err
			}
			rep.Rows++
			switch {
			case string(x.mac) == string(pendingMAC):
				x.unsealed = true
				sus = append(sus, x)
			case !sc.S.Verify(x.mac, t.name, x.ver, x.canon):
				sus = append(sus, x)
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		if len(sus) == 0 {
			return nil
		}
		// the primary key, to name (and later re-seal) a record
		var pk []string
		pr, err := tx.Query(ctx, `select a.attname from pg_index i join pg_attribute a on a.attrelid = i.indrelid and a.attnum = any(i.indkey) where i.indrelid = $1::regclass and i.indisprimary order by a.attnum`, "hub."+t.name)
		if err != nil {
			return err
		}
		for pr.Next() {
			var c string
			if err := pr.Scan(&c); err != nil {
				pr.Close()
				return err
			}
			pk = append(pk, c)
		}
		pr.Close()
		for _, x := range sus {
			if len(rep.Problems) >= sc.MaxBad {
				break
			}
			row := strings.TrimSpace(x.canon)
			if len(row) > 160 {
				row = row[:160] + "…"
			}
			p := Problem{Table: t.name, Row: row}
			if t.tenant {
				p.Org = org.String()
			}
			var data map[string]any
			_ = json.Unmarshal([]byte(x.data), &data)
			text := func(k string) (string, bool) {
				v, ok := data[k]
				if !ok || v == nil {
					return "", false
				}
				if f, isNum := v.(float64); isNum {
					return strconv.FormatFloat(f, 'f', -1, 64), true
				}
				return fmt.Sprint(v), true
			}
			p.PK, p.Ident = map[string]string{}, map[string]string{}
			for _, k := range pk {
				if v, ok := text(k); ok {
					p.PK[k] = v
				}
			}
			for _, k := range identKeys {
				if v, ok := text(k); ok && len(v) <= 80 {
					p.Ident[k] = v
				}
			}
			switch {
			case x.unsealed:
				p.Cause, p.Reason = "unsealed", "the record never got its seal"
			default:
				p.Cause, p.Reason = "changed", "the record no longer matches its seal: it was changed outside the system, or rolled back"
				// Does it match once the column(s) added since it was sealed are left out? Then the record is authentic.
				if cols := sc.findAdded(ctx, tx, t.name, x.ctid, x.nullKeys, x.ver, x.mac); cols != nil {
					p.Cause, p.Added = "schema", cols
					p.Reason = "a database update added " + strings.Join(cols, ", ") + " after the record was sealed; its content is unchanged"
				}
			}
			rep.Problems = append(rep.Problems, p)
		}
		return nil
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
