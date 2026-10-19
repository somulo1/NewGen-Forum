// Package people holds what every module needs to know about person records: how personal fields are protected.
package people

import (
	"github.com/google/uuid"

	"hub/internal/crypto"
)

// Fields that are personal data and stored encrypted with the organization's data key.
const (
	ColEmail  = "email_enc"
	ColPhone  = "phone_enc"
	ColDOB    = "dob_enc"
	ColGender = "gender_enc"
)

func ctx(org, id uuid.UUID, col string) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "people", Column: col, RowID: id.String()}
}

// Enc encrypts an empty string as nil so "not set" is visible without decrypting.
func Enc(dek []byte, org, id uuid.UUID, col, v string) ([]byte, error) {
	if v == "" {
		return nil, nil
	}
	return crypto.Field(dek, ctx(org, id, col), []byte(v))
}

func Dec(dek []byte, org, id uuid.UUID, col string, ct []byte) string {
	if len(ct) == 0 {
		return ""
	}
	b, err := crypto.FieldOpen(dek, ctx(org, id, col), ct)
	if err != nil {
		return ""
	}
	return string(b)
}

// EmailIdx is the blind index used to look a person up by email inside one organization.
func EmailIdx(blindKey []byte, org uuid.UUID, email string) []byte {
	return crypto.Blind(blindKey, "person-email:"+org.String(), email)
}
