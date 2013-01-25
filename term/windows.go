// +build windows

package term

import (
	"os"
)

// IsTerminal returns false on Windows.
func IsTerminal(f *os.File) bool {
	return false
}

// MakeRaw is a no-op on windows. It returns nil.
func MakeRaw(f *os.File) error {
	return nil
}

// Restore is a no-op on windows. It returns nil.
func Restore(f *os.File) error {
	return nil
}

// Cols returns 80 on Windows.
func Cols() (int, error) {
	return 80, nil
}

// Lines returns 24 on Windows.
func Lines() (int, error) {
	return 24, nil
}
