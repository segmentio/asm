//go:build !purego
// +build !purego

package utf8

import (
	stdutf8 "unicode/utf8"

	segascii "github.com/segmentio/asm/ascii"
)

func Validate(p []byte) (bool, bool) {
	asciivalid := segascii.Valid(p)
	if asciivalid {
		return true, true
	}
	if len(p) < 32 {
		return stdutf8.Valid(p), false
	}
	return validateAvx(p)
}
