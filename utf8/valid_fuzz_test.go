//go:build go1.18
// +build go1.18

package utf8

import (
	"testing"
	stdlib "unicode/utf8"
)

func FuzzValid(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		result := Valid(data)
		ref := stdlib.Valid(data)
		if result != ref {
			t.Errorf("Valid(%q) = %v; want %v", data, result, ref)
		}
	})
}
