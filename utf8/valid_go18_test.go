//go:build go1.18
// +build go1.18

package utf8

import (
	"testing"
	stdlib "unicode/utf8"

	"github.com/segmentio/asm/ascii"
)

func FuzzValid(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		vu, va := Validate(data)
		ru := stdlib.Valid(data)
		if ru != vu {
			t.Errorf("Validate(%q) UTF8 = %v; want %v", data, vu, ru)
		}
		ra := ascii.Valid(data)
		if ra != va {
			t.Errorf("Validate(%q) ASCII = %v; want %v", data, va, ra)
		}
	})
}
