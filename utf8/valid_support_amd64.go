//go:build !purego
// +build !purego

package utf8

import (
	"github.com/segmentio/asm/cpu"
	"github.com/segmentio/asm/cpu/x86"
)

var noAVX2 = !cpu.X86.Has(x86.AVX2)

// Validate is a more precise version of Valid that also indicates whether the
// input was valid ASCII.
func Validate(p []byte) Validation {
	if noAVX2 || len(p) < 32 {
		return validate(p)
	}
	r := validateAvx(p)
	return Validation(r)
}
