package crypto

import (
	"bytes"
	"testing"
)

func TestSealOpenRoundTripAndBinding(t *testing.T) {
	key := RandomBytes(32)
	ctx := Context{Org: "o1", Table: "people", Column: "phone_enc", RowID: "r1"}
	ct, err := Field(key, ctx, []byte("+254700000000"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(ct, []byte("254700")) {
		t.Fatal("plaintext leaked into ciphertext")
	}
	pt, err := FieldOpen(key, ctx, ct)
	if err != nil || string(pt) != "+254700000000" {
		t.Fatalf("round trip failed: %v %q", err, pt)
	}
	// moving the ciphertext to another row, column or organization must fail
	for _, bad := range []Context{{"o2", "people", "phone_enc", "r1"}, {"o1", "people", "dob_enc", "r1"}, {"o1", "people", "phone_enc", "r2"}, {"o1", "accounts", "phone_enc", "r1"}} {
		if _, err := FieldOpen(key, bad, ct); err == nil {
			t.Fatalf("ciphertext opened in the wrong place: %+v", bad)
		}
	}
	// flipping a bit must fail
	ct[len(ct)-1] ^= 1
	if _, err := FieldOpen(key, ctx, ct); err == nil {
		t.Fatal("tampered ciphertext was accepted")
	}
}

func TestSealIsRandomised(t *testing.T) {
	k := RandomBytes(32)
	a, _ := Seal(k, []byte("same"), nil)
	b, _ := Seal(k, []byte("same"), nil)
	if bytes.Equal(a, b) {
		t.Fatal("identical plaintexts produced identical ciphertexts (nonce reuse)")
	}
}

func TestOrgKeysWrapUnwrap(t *testing.T) {
	ok := NewOrgKeys(RandomBytes(32))
	dek, wrapped, err := ok.NewDEK("org-a", 1)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ok.DEK("org-a", 1, wrapped)
	if err != nil || !bytes.Equal(got, dek) {
		t.Fatal("unwrap failed")
	}
	ok.Forget()
	if _, err := ok.DEK("org-b", 1, wrapped); err == nil {
		t.Fatal("a wrapped key opened for a different organization")
	}
	other := NewOrgKeys(RandomBytes(32))
	if _, err := other.DEK("org-a", 1, wrapped); err == nil {
		t.Fatal("a different KEK unwrapped the key")
	}
}

func TestMACLengthPrefix(t *testing.T) {
	k := RandomBytes(32)
	if Equal(MAC(k, []byte("ab"), []byte("c")), MAC(k, []byte("a"), []byte("bc"))) {
		t.Fatal("MAC is ambiguous across part boundaries")
	}
}

func TestBlindIndexIsStableAndScoped(t *testing.T) {
	k := RandomBytes(32)
	if !bytes.Equal(Blind(k, "email", "a@b.co"), Blind(k, "email", "a@b.co")) {
		t.Fatal("not stable")
	}
	if bytes.Equal(Blind(k, "email", "a@b.co"), Blind(k, "phone", "a@b.co")) {
		t.Fatal("scopes must differ")
	}
}

func TestPasswordHash(t *testing.T) {
	fast := PasswordParams{Time: 1, MemoryKiB: 8 * 1024, Threads: 1}
	h := HashPassword("correct horse", fast)
	if !VerifyPassword("correct horse", h) || VerifyPassword("wrong", h) {
		t.Fatal("password verification wrong")
	}
	if HashPassword("correct horse", fast) == h {
		t.Fatal("salt not random")
	}
	if VerifyPassword("x", "garbage") {
		t.Fatal("garbage hash verified")
	}
}

func TestDigitsAreNumericAndLong(t *testing.T) {
	d := Digits(6)
	if len(d) != 6 {
		t.Fatal("length")
	}
	for _, c := range d {
		if c < '0' || c > '9' {
			t.Fatal("non-digit")
		}
	}
}
