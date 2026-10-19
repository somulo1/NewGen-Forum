// Package integrity keeps the database honest: row MACs (a row changed outside the system fails verification)
// and a hash-chained ledger (history cannot be rewritten without breaking every later hash).
package integrity

import (
	"context"
	"encoding/binary"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5"

	"hub/internal/crypto"
)

// Bump is the SQL fragment every UPDATE of a protected row must include; the system then seals the row (see Seal).
const Bump = "ver = ver + 1, mac = hub.pending_mac()"

// BumpOn is Bump for upserts (INSERT ... ON CONFLICT DO UPDATE), where the column name is ambiguous without the table.
func BumpOn(table string) string { return "ver = hub." + table + ".ver + 1, mac = hub.pending_mac()" }

// Signer computes and checks row MACs with a key that is never stored in the database.
type Signer struct{ key []byte }

func NewSigner(key []byte) *Signer { return &Signer{key: key} }

// Row returns the MAC of one row. canon is the database's canonical JSON of the row (hub.canon: every column except mac, ver
// and updated_at), so ids, organization and all content are covered; table and version are mixed in so a row cannot be moved
// to another table or rolled back to an older version.
func (s *Signer) Row(table string, ver int, canon string) []byte {
	return crypto.MAC(s.key, []byte("hub/row/v2"), []byte(table), itob(int64(ver)), []byte(canon))
}

func (s *Signer) Verify(mac []byte, table string, ver int, canon string) bool {
	return crypto.Equal(mac, s.Row(table, ver, canon))
}

func itob(n int64) []byte { var b [8]byte; binary.BigEndian.PutUint64(b[:], uint64(n)); return b[:] }

var ident = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// Seal gives the rows matched by where their real MAC. It runs in the same transaction as the write that made them
// "pending", so a pending row is never visible to anyone else. where may be empty only if the caller just wrote one row.
func (s *Signer) Seal(ctx context.Context, tx pgx.Tx, table, where string, args ...any) error {
	if !ident.MatchString(table) {
		return fmt.Errorf("integrity: bad table name %q", table)
	}
	if where == "" {
		return fmt.Errorf("integrity: Seal needs a where clause")
	}
	q := fmt.Sprintf(`select ctid::text, tableoid::int8, ver, hub.canon(t) from hub.%s t where mac = hub.pending_mac() and (%s)`, table, where)
	rs, err := tx.Query(ctx, q, args...)
	if err != nil {
		return err
	}
	type item struct {
		ctid, canon string
		oid         int64
		ver         int
	}
	var items []item
	for rs.Next() {
		var it item
		if err := rs.Scan(&it.ctid, &it.oid, &it.ver, &it.canon); err != nil {
			rs.Close()
			return err
		}
		items = append(items, it)
	}
	rs.Close()
	if err := rs.Err(); err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	up := fmt.Sprintf(`update hub.%s set mac = $1 where ctid = $2::text::tid and tableoid = $3::int8::oid`, table)
	b := &pgx.Batch{}
	for _, it := range items {
		b.Queue(up, s.Row(table, it.ver, it.canon), it.ctid, it.oid)
	}
	br := tx.SendBatch(ctx, b)
	for range items {
		if _, err := br.Exec(); err != nil {
			_ = br.Close()
			return err
		}
	}
	return br.Close()
}
