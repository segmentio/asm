//go:build !purego
// +build !purego

package utf8

import (
	stdutf8 "unicode/utf8"
)

func Validate(p []byte) (bool, bool) {
	if len(p) < 32 {
		return stdutf8.Valid(p), false
	}
	return validateAvx(p)
}
