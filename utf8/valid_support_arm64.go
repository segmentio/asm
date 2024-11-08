//go:build !purego || arm64
// +build !purego arm64

package utf8

import (
	"github.com/segmentio/asm/cpu"
	"github.com/segmentio/asm/cpu/arm64"
)

var noNEON = !cpu.ARM64.Has(arm64.ASIMD)

// Validate is a more precise version of Valid that also indicates whether the
// input was valid ASCII.
func Validate(p []byte) Validation {
	if noNEON || len(p) < 32 {
		return validate(p)
	}
	r := validateNEON(p)
	return Validation(r)
}
