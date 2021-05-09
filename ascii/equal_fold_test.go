package ascii

import (
	"math/rand"
	"strings"
	"testing"
	"unicode/utf8"
)

var testValues = [...]string{
	"",
	"a",
	"ab",
	"abc",
	"abcd",
	"hello",
	"Hello World!",
	"Hello\"World!",
	"Hello\\World!",
	"Hello\nWorld!",
	"Hello\rWorld!",
	"Hello\tWorld!",
	"Hello\bWorld!",
	"Hello\fWorld!",
	"H~llo World!",
	"H~llo",
	"你好",
	"~",
	"\x80",
	"\x7F",
	"\xFF",
	"a string of 16B.",
	"an invalid string of 32B. \x00......",
	"some kind of long string with only ascii characters.",
	"some kind of long string with a non-ascii character at the end.\xff",
	strings.Repeat("1234567890", 1000),
}

var testValeusUTF8 []string

func init() {
	for _, test := range testValues {
		if utf8.ValidString(test) {
			testValeusUTF8 = append(testValeusUTF8, test)
		}
	}
}

func TestHasPrefixFoldString(t *testing.T) {
	for _, test := range testValeusUTF8 {
		t.Run(limit(test), func(t *testing.T) {
			prefix := test
			if len(prefix) > 0 {
				prefix = prefix[:len(prefix)/2]
			}
			upper := strings.ToUpper(prefix)
			lower := strings.ToLower(prefix)

			if !HasPrefixFoldString(test, prefix) {
				t.Errorf("%q does not match %q", test, prefix)
			}

			if !HasPrefixFoldString(test, upper) {
				t.Errorf("%q does not match %q", test, upper)
			}

			if !HasPrefixFoldString(test, lower) {
				t.Errorf("%q does not match %q", test, lower)
			}
		})
	}
}

func TestHasSuffixFoldString(t *testing.T) {
	for _, test := range testValeusUTF8 {
		t.Run(limit(test), func(t *testing.T) {
			suffix := test
			if len(suffix) > 0 {
				suffix = suffix[len(suffix)/2:]
			}
			upper := strings.ToUpper(suffix)
			lower := strings.ToLower(suffix)

			if !HasSuffixFoldString(test, suffix) {
				t.Errorf("%q does not match %q", test, suffix)
			}

			if !HasSuffixFoldString(test, upper) {
				t.Errorf("%q does not match %q", test, upper)
			}

			if !HasSuffixFoldString(test, lower) {
				t.Errorf("%q does not match %q", test, lower)
			}
		})
	}
}

func TestEqualFoldString(t *testing.T) {
	// Only test valid UTF-8 otherwise ToUpper/ToLower will convert invalid
	// characters to UTF-8 placeholders, which breaks the case-insensitive
	// equality.
	for _, test := range testValeusUTF8 {
		t.Run(limit(test), func(t *testing.T) {
			upper := strings.ToUpper(test)
			lower := strings.ToLower(test)

			if !EqualFoldString(test, test) {
				t.Errorf("%q does not match %q", test, test)
			}

			if !EqualFoldString(test, upper) {
				t.Errorf("%q does not match %q", test, upper)
			}

			if !EqualFoldString(test, lower) {
				t.Errorf("%q does not match %q", test, lower)
			}

			if len(test) > 1 {
				reverse := make([]byte, len(test))
				for i := range reverse {
					reverse[i] = test[len(test)-(i+1)]
				}

				if EqualFoldString(test, string(reverse)) {
					t.Errorf("%q matches %q", test, reverse)
				}
			}
		})
	}
}

func TestEqualFold(t *testing.T) {
	lower := make([]byte, 260)
	upper := make([]byte, len(lower))
	mixed := make([]byte, len(lower))

	for i := 0; i < len(lower); i++ {
		l := lower[:i+1]
		u := upper[:i+1]
		m := mixed[:i+1]

		l[i] = 'a'
		u[i] = 'z'
		if EqualFold(l, u) {
			t.Errorf("%q matches %q", l, u)
		}

		l[i] = byte(i%26) + 'A'
		u[i] = byte(i%26) + 'a'
		if rand.Int()%2 == 0 {
			m[i] = l[i]
		} else {
			m[i] = u[i]
		}

		if EqualFold(l[:len(l)-1], u) {
			t.Errorf("%q matches %q", l[:len(l)-1], u)
		}

		if !EqualFold(l, u) {
			t.Errorf("%q does not match %q", l, u)
		}

		if !EqualFold(l, m) {
			t.Errorf("%q does not match %q", l, m)
		}

		if !EqualFold(u, m) {
			t.Errorf("%q does not match %q", u, m)
		}
	}
}

func BenchmarkEqualFoldString(b *testing.B) {
	const N = 16

	lower := ""
	upper := ""

	for i := 0; i < N; i++ {
		lower += string(byte(i%26) + 'A')
		upper += string(byte(i%26) + 'a')
	}

	for i := 0; i < b.N; i++ {
		EqualFoldString(lower, upper)
	}
}

func limit(s string) string {
	if len(s) > 17 {
		return s[:17] + "..."
	}
	return s
}
