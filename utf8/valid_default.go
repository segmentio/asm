//go:build purego || !amd64
// +build purego !amd64

package utf8

import (
	"github.com/segmentio/asm/ascii"
	stdlib "unicode/utf8"
)


// Validate is a more precise version of Valid that also indicates whether the
// input was valid ASCII.
func Validate(p []byte) (utf8, ascii bool) {
	valid := ascii.Valid(p)
	if valid {
		return true, true
	}
	return stdlib.Valid(p), false
}
