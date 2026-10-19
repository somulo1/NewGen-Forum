package e2ee

import (
	"bytes"
	"testing"
)

func TestEndToEnd(t *testing.T) {
	alice, _ := NewDevice()
	bob, _ := NewDevice()
	eve, _ := NewDevice()
	ck, _ := NewConversationKey()
	ctx := Context{Org: "o", Conversation: "c", Sender: "alice", MessageID: "m1", Epoch: 1}
	wctx := []byte("o|c|1")

	wb, err := Wrap(ck, bob.Public().Identity, wctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateWrapped(wb); err != nil {
		t.Fatal(err)
	}
	got, err := bob.Unwrap(wb, wctx)
	if err != nil || !bytes.Equal(got, ck) {
		t.Fatal("bob could not unwrap")
	}
	if _, err := eve.Unwrap(wb, wctx); err == nil {
		t.Fatal("someone else unwrapped bob's key")
	}
	if _, err := bob.Unwrap(wb, []byte("o|c|2")); err == nil {
		t.Fatal("wrapped key opened for the wrong epoch/context")
	}

	env, err := alice.Seal(ck, []byte("meet at 9"), ctx)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(env, []byte("meet")) {
		t.Fatal("plaintext in the envelope")
	}
	pt, err := Open(got, env, alice.Public().Signing, ctx)
	if err != nil || string(pt) != "meet at 9" {
		t.Fatalf("bob could not read: %v", err)
	}
	// wrong sender key, replay in another conversation / as another message, tampering
	if _, err := Open(got, env, eve.Public().Signing, ctx); err != ErrSig {
		t.Fatal("forged sender accepted")
	}
	for _, bad := range []Context{{"o", "c2", "alice", "m1", 1}, {"o", "c", "alice", "m2", 1}, {"o", "c", "alice", "m1", 2}, {"o2", "c", "alice", "m1", 1}, {"o", "c", "bob", "m1", 1}} {
		if _, err := Open(got, env, alice.Public().Signing, bad); err == nil {
			t.Fatalf("envelope accepted in the wrong context %+v", bad)
		}
	}
	env[len(env)-70] ^= 1
	if _, err := Open(got, env, alice.Public().Signing, ctx); err == nil {
		t.Fatal("tampered envelope accepted")
	}
	otherKey, _ := NewConversationKey()
	env2, _ := alice.Seal(ck, []byte("x"), ctx)
	if _, err := Open(otherKey, env2, alice.Public().Signing, ctx); err == nil {
		t.Fatal("opened with the wrong conversation key")
	}
}

func TestDeterministicDemoDevices(t *testing.T) {
	a, _ := DeviceFromSeed([]byte("seed-1"))
	b, _ := DeviceFromSeed([]byte("seed-1"))
	c, _ := DeviceFromSeed([]byte("seed-2"))
	if !bytes.Equal(a.Public().Identity, b.Public().Identity) || bytes.Equal(a.Public().Identity, c.Public().Identity) {
		t.Fatal("seed derivation wrong")
	}
}
