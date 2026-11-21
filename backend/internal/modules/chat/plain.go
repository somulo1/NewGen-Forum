package chat

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/crypto"
)

// The community wall is not end-to-end encrypted: its posts, comments and moments are read and checked by the server, and kept
// encrypted at rest with the organization's data key, each body bound to its message. They are stored in the same messages table
// (so seen tracking, catch-up, deleting and moments work as before), marked by this prefix; older wall messages stay sealed and
// are still opened on devices.
var plainMark = []byte("HUBP1:")

func plainCtx(org, msg uuid.UUID) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "messages", Column: "body", RowID: msg.String()}
}

// SealPlain turns a JSON body into what is stored for a server-readable message.
func SealPlain(ctx context.Context, tx pgx.Tx, a *app.App, org, msg uuid.UUID, body []byte) ([]byte, error) {
	if !json.Valid(body) {
		return nil, errBadBody
	}
	dek, _, err := a.OrgDEK(ctx, tx, org)
	if err != nil {
		return nil, err
	}
	ct, err := crypto.Field(dek, plainCtx(org, msg), body)
	if err != nil {
		return nil, err
	}
	return append(append([]byte{}, plainMark...), ct...), nil
}

// IsPlain says whether stored bytes are a server-readable body.
func IsPlain(stored []byte) bool { return bytes.HasPrefix(stored, plainMark) }

// OpenPlain gives back the JSON body of a server-readable message.
func OpenPlain(dek []byte, org, msg uuid.UUID, stored []byte) (json.RawMessage, bool) {
	if !IsPlain(stored) {
		return nil, false
	}
	b, err := crypto.FieldOpen(dek, plainCtx(org, msg), stored[len(plainMark):])
	if err != nil {
		return nil, false
	}
	return json.RawMessage(b), true
}

// serverHeld kinds are not end-to-end encrypted: notes (a person's own and the ones they share) are kept by the server, encrypted at
// rest with the organization's key, so they open on any device and no key ever has to be handed from one device to another.
func serverHeld(kind string) bool { return kind == "note" || kind == "sharednote" }

func recordCtx(org, rec uuid.UUID) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "records", Column: "body", RowID: rec.String()}
}

// SealRecord turns a record's JSON body into what is stored for a server-held record.
func SealRecord(ctx context.Context, tx pgx.Tx, a *app.App, org, rec uuid.UUID, body []byte) ([]byte, error) {
	if !json.Valid(body) {
		return nil, errBadBody
	}
	dek, _, err := a.OrgDEK(ctx, tx, org)
	if err != nil {
		return nil, err
	}
	ct, err := crypto.Field(dek, recordCtx(org, rec), body)
	if err != nil {
		return nil, err
	}
	return append(append([]byte{}, plainMark...), ct...), nil
}

// OpenRecord gives back the JSON body of a server-held record.
func OpenRecord(dek []byte, org, rec uuid.UUID, stored []byte) (json.RawMessage, bool) {
	if !IsPlain(stored) {
		return nil, false
	}
	b, err := crypto.FieldOpen(dek, recordCtx(org, rec), stored[len(plainMark):])
	if err != nil {
		return nil, false
	}
	return json.RawMessage(b), true
}

type badBody struct{}

func (badBody) Error() string { return "not a JSON body" }

var errBadBody = badBody{}
