//go:build purego || !amd64
// +build purego !amd64

package utf8

import stdlib "unicode/utf8"

// Valid reports whether p consists entirely of valid UTF-8-encoded runes.
func Valid(p []byte) bool {
	return stdlib.Valid(p)
}
