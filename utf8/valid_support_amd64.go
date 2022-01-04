//go:build !purego
// +build !purego

package utf8

import (
	"unicode/utf8"

	"github.com/segmentio/asm/ascii"
)

func Validate(p []byte) (validUtf8, validAscii bool) {
	if len(p) < 32 {
		validAscii = ascii.Valid(p)
		if validAscii {
			return true, true
		}
		return utf8.Valid(p), false
	}
	return validateAvx(p)
}
