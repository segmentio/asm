//go:build purego || !amd64
// +build purego !amd64

package utf8

import (
	"unicode/utf8"

	"github.com/segmentio/asm/ascii"
)


// Validate is a more precise version of Valid that also indicates whether the
// input was valid ASCII.
func Validate(p []byte) (validUtf8, validAscii bool) {
	valid := ascii.Valid(p)
	if valid {
		return true, true
	}
	return utf8.Valid(p), false
}
