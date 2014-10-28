package cli

import (
	"testing"
)

func TestHomeDir(t *testing.T) {
	if homeDir() == "" {
		t.Fatal("homeDir should not be blank")
	}
}
