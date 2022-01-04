//go:build purego || !amd64
// +build purego !amd64

package utf8

import (
	segascii "github.com/segmentio/asm/ascii"
	stdutf8 "unicode/utf8"
)


// Validate is a more precise version of Valid that also indicates whether the
// input was valid ASCII.
func Validate(p []byte) (utf8, ascii bool) {
	valid := segascii.Valid(p)
	if valid {
		return true, true
	}
	return stdutf8.Valid(p), false
}
