package main

import (
	"testing"
)

func TestStringsIndex(t *testing.T) {
	a1 := []string{}
	if res := stringsIndex(a1, ""); res != -1 {
		t.Errorf("expected -1, got %d", res)
	}
	if res := stringsIndex(a1, "-a"); res != -1 {
		t.Errorf("expected -1, got %d", res)
	}

	a2 := []string{"-a", "bbq"}
	if res := stringsIndex(a2, "-a"); res != 0 {
		t.Errorf("expected 0, got %d", res)
	}
	if res := stringsIndex(a2, "bbq"); res != 1 {
		t.Errorf("expected 1, got %d", res)
	}
	if res := stringsIndex(a2, ""); res != -1 {
		t.Errorf("expected -1, got %d", res)
	}
}
