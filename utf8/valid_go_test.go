package utf8

import (
	"bytes"
	"fmt"
	"testing"
	stdutf8 "unicode/utf8"
)

type byteRange struct {
	Low  byte
	High byte
}

func one(b byte) byteRange {
	return byteRange{b, b}
}

func genExamples(current string, ranges []byteRange) []string {
	if len(ranges) == 0 {
		return []string{string(current)}
	}
	r := ranges[0]
	var all []string

	for x := r.Low; x <= r.High; x++ {
		s := current + string(x)
		all = append(all, genExamples(s, ranges[1:])...)
		if x == r.High {
			break
		}
	}
	return all
}

func TestValid(t *testing.T) {
	// Tests copied from the stdlib
	var examples = []string{
		"",
		"a",
		"abc",
		"Ж",
		"ЖЖ",
		"брэд-ЛГТМ",
		"☺☻☹",

		// overlong
		"\xE0\x80",
		// unfinished continuation
		"aa\xE2",

		string([]byte{66, 250}),

		string([]byte{66, 250, 67}),

		"a\uFFFDb",

		"\xF4\x8F\xBF\xBF", // U+10FFFF

		"\xF4\x90\x80\x80", // U+10FFFF+1; out of range
		"\xF7\xBF\xBF\xBF", // 0x1FFFFF; out of range

		"\xFB\xBF\xBF\xBF\xBF", // 0x3FFFFFF; out of range

		"\xc0\x80",     // U+0000 encoded in two bytes: incorrect
		"\xed\xa0\x80", // U+D800 high surrogate (sic)
		"\xed\xbf\xbf", // U+DFFF low surrogate (sic)
	}

	for _, tt := range examples {
		t.Run(tt, func(t *testing.T) {
			b := []byte(tt)
			expected := stdutf8.Valid(b)
			if Valid(b) != expected {
				t.Errorf("Valid(%q) = %v; want %v", tt, !expected, expected)
			}
		})
	}
}

// Takes about 10s to run on my machine.
func TestValidExhaustive(t *testing.T) {
	any := byteRange{0, 0xFF}
	ascii := byteRange{0, 0x7F}
	cont := byteRange{0x80, 0xBF}

	rangesToTest := [][]byteRange{
		{one(0x20), ascii, ascii, ascii},

		// 2-byte sequences
		{one(0xC2)},
		{one(0xC2), ascii},
		{one(0xC2), cont},
		{one(0xC2), {0xC0, 0xFF}},
		{one(0xC2), cont, cont},
		{one(0xC2), cont, cont, cont},

		// 3-byte sequences
		{one(0xE1)},
		{one(0xE1), cont},
		{one(0xE1), cont, cont},
		{one(0xE1), cont, cont, ascii},
		{one(0xE1), cont, ascii},
		{one(0xE1), cont, cont, cont},

		// 4-byte sequences
		{one(0xF1)},
		{one(0xF1), cont},
		{one(0xF1), cont, cont},
		{one(0xF1), cont, cont, cont},
		{one(0xF1), cont, cont, ascii},
		{one(0xF1), cont, cont, cont, ascii},

		// overlong
		{{0xC0, 0xC1}, any},
		{{0xC0, 0xC1}, any, any},
		{{0xC0, 0xC1}, any, any, any},
		{one(0xE0), {0x0, 0x9F}, cont},
		{one(0xE0), {0xA0, 0xBF}, cont},
	}

	t.Parallel()

	for _, r := range rangesToTest {
		r := r
		name := fmt.Sprintf("%v", r)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			examples := genExamples("", r)

			for _, e := range examples {
				b := []byte(e)
				expected := stdutf8.Valid(b)
				if Valid(b) != expected {
					t.Errorf("Valid(%q) = %v; want %v", e, !expected, expected)
				}
			}

		})
	}
}

var valid1k = bytes.Repeat([]byte("0123456789日本語日本語日本語日abcdefghijklmnopqrstuvwx"), 16)
var valid1M = bytes.Repeat(valid1k, 1024)

func BenchmarkValid(b *testing.B) {
	impls := map[string]func([]byte) bool{
		"AVX":    Valid,
		"Stdlib": stdutf8.Valid,
	}

	inputs := map[string][]byte{
		"1kValid": valid1k,
		"1MValid": valid1M,
		"10ASCII": []byte("0123456789"),
		"10Japan": []byte("日本語日本語日本語日"),
	}

	for inputName, input := range inputs {
		for implName, f := range impls {
			testName := fmt.Sprintf("%s/%s", inputName, implName)
			b.Run(testName, func(b *testing.B) {
				b.SetBytes(int64(len(input)))
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					f(input)
				}
			})
		}
	}
}