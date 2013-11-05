package main

import (
	"testing"
)

var expectedIdentities = []string{"Blake Gentry <blakesgentry@gmail.com>"}

func TestPubKeyRing(t *testing.T) {
	k, err := PubKeyRing()
	if err != nil {
		t.Fatal(err)
	}
	if len(k) != 1 {
		t.Errorf("expected 1 key in public ring, got %d", len(k))
	}
	if len(k[0].Identities) != 1 {
		t.Errorf("expected 1 identity, got %d", len(k[0].Identities))
	}
	for _, name := range expectedIdentities {
		_, ok := k[0].Identities[name]
		if !ok {
			t.Errorf("identity not found for %q", name)
		}
	}
}
