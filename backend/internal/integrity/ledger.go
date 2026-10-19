package integrity

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/crypto"
)

// Ledger is the per-organization hash chain.
type Ledger struct{ key []byte }

func NewLedger(key []byte) *Ledger { return &Ledger{key: key} }

func genesis(org uuid.UUID) []byte { return crypto.Hash([]byte("hub/ledger/genesis"), org[:]) }

func entryHash(org uuid.UUID, seq int64, ts time.Time, actor *uuid.UUID, action, subject string, payload, prev []byte) []byte {
	var a []byte
	if actor != nil {
		a = actor[:]
	}
	return plainHash([]byte("hub/ledger/v1"), org[:], itob(seq), itob(ts.UnixMicro()), a, []byte(action), []byte(subject), payload, prev)
}

// plainHash is an unkeyed, length-prefixed SHA-256: anyone can recompute the chain; the signature is what needs the secret key.
func plainHash(parts ...[]byte) []byte {
	h := sha256.New()
	for _, p := range parts {
		h.Write(itob(int64(len(p))))
		h.Write(p)
	}
	return h.Sum(nil)
}

// Append adds one entry inside the caller's transaction. It locks the chain head, so entries for one organization are strictly ordered.
func (l *Ledger) Append(ctx context.Context, tx pgx.Tx, org uuid.UUID, actor *uuid.UUID, action, subject string, payload []byte) (int64, error) {
	var seq int64
	var prev []byte
	err := tx.QueryRow(ctx, "select seq, hash from hub.ledger_heads where org_id=$1 for update", org).Scan(&seq, &prev)
	if errors.Is(err, pgx.ErrNoRows) {
		prev = genesis(org)
		if _, err = tx.Exec(ctx, "insert into hub.ledger_heads (org_id, seq, hash) values ($1,0,$2) on conflict (org_id) do nothing", org, prev); err != nil {
			return 0, err
		}
		err = tx.QueryRow(ctx, "select seq, hash from hub.ledger_heads where org_id=$1 for update", org).Scan(&seq, &prev)
	}
	if err != nil {
		return 0, err
	}
	seq++
	ts := time.Now().UTC().Truncate(time.Microsecond)
	ph := sha256.Sum256(payload)
	h := entryHash(org, seq, ts, actor, action, subject, ph[:], prev)
	sig := crypto.MAC(l.key, h)
	if _, err = tx.Exec(ctx, `insert into hub.ledger (org_id, seq, ts, actor, action, subject, payload_hash, prev_hash, hash, sig) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		org, seq, ts, actor, action, subject, ph[:], prev, h, sig); err != nil {
		return 0, err
	}
	if _, err = tx.Exec(ctx, "update hub.ledger_heads set seq=$2, hash=$3 where org_id=$1", org, seq, h); err != nil {
		return 0, err
	}
	return seq, nil
}

// Break describes the first place the chain stops being valid.
type Break struct {
	Seq    int64
	Reason string
}

func (b *Break) Error() string { return fmt.Sprintf("ledger broken at seq %d: %s", b.Seq, b.Reason) }

// Verify walks the whole chain for an organization (cheap per row; run by cmd/integrity on a schedule).
// It returns the number of entries checked, or a *Break at the first problem.
func (l *Ledger) Verify(ctx context.Context, q interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}, org uuid.UUID) (int64, error) {
	rows, err := q.Query(ctx, `select seq, ts, actor, action, subject, payload_hash, prev_hash, hash, sig from hub.ledger where org_id=$1 order by seq`, org)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	prev := genesis(org)
	want := int64(1)
	var n int64
	for rows.Next() {
		var seq int64
		var ts time.Time
		var actor *uuid.UUID
		var action, subject string
		var ph, pv, h, sig []byte
		if err := rows.Scan(&seq, &ts, &actor, &action, &subject, &ph, &pv, &h, &sig); err != nil {
			return n, err
		}
		if seq != want {
			return n, &Break{seq, fmt.Sprintf("expected entry %d: an entry is missing", want)}
		}
		if !crypto.Equal(pv, prev) {
			return n, &Break{seq, "previous hash does not match the entry before it"}
		}
		if !crypto.Equal(h, entryHash(org, seq, ts.UTC().Truncate(time.Microsecond), actor, action, subject, ph, pv)) {
			return n, &Break{seq, "entry contents do not match its hash"}
		}
		if !crypto.Equal(sig, crypto.MAC(l.key, h)) {
			return n, &Break{seq, "signature is not valid"}
		}
		prev, want = h, seq+1
		n++
	}
	if err := rows.Err(); err != nil {
		return n, err
	}
	// The head must point at the last entry, otherwise the tail was cut off.
	var hs int64
	var hh []byte
	err = q.QueryRow(ctx, "select seq, hash from hub.ledger_heads where org_id=$1", org).Scan(&hs, &hh)
	if errors.Is(err, pgx.ErrNoRows) {
		return n, nil
	}
	if err != nil {
		return n, err
	}
	if hs != n || (n > 0 && !crypto.Equal(hh, prev)) {
		return n, &Break{n + 1, "chain head does not match the last entry: entries were removed from the end"}
	}
	return n, nil
}

// MerkleRoot combines entry hashes into one root, so a periodic signed checkpoint can be stored (or anchored elsewhere).
func MerkleRoot(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return crypto.Hash([]byte("empty"))
	}
	level := hashes
	for len(level) > 1 {
		var next [][]byte
		for i := 0; i < len(level); i += 2 {
			if i+1 == len(level) {
				next = append(next, crypto.Hash([]byte{1}, level[i], level[i]))
			} else {
				next = append(next, crypto.Hash([]byte{1}, level[i], level[i+1]))
			}
		}
		level = next
	}
	return level[0]
}
