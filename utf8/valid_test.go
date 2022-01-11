package utf8

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/segmentio/asm/ascii"
	"github.com/segmentio/asm/internal/buffer"
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

	elements := []byte{r.Low, r.High}

	mid := (r.High + r.Low) / 2
	if mid != r.Low && mid != r.High {
		elements = append(elements, mid)
	}

	for _, x := range elements {
		s := current + string(x)
		all = append(all, genExamples(s, ranges[1:])...)
		if x == r.High {
			break
		}
	}
	return all
}

func TestValid(t *testing.T) {
	var examples = []string{
		// Tests copied from the stdlib
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

		// valid at boundary
		strings.Repeat("a", 32+28) + "☺☻☹",
		strings.Repeat("a", 32+29) + "☺☻☹",
		strings.Repeat("a", 32+30) + "☺☻☹",
		strings.Repeat("a", 32+31) + "☺☻☹",
		// invalid at boundary
		strings.Repeat("a", 32+31) + "\xE2a",

		// same inputs as benchmarks
		"0123456789",
		"日本語日本語日本語日",
		"\xF4\x8F\xBF\xBF",

		// bugs found with fuzzing
		"0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\xc60",
		"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\xc300",
		"߀0000000000000000000000000000訨",
		"0000000000000000000000000000000˂00000000000000000000000000000000",
	}

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

	for _, r := range rangesToTest {
		examples = append(examples, genExamples("", r)...)
	}

	for _, i := range []int{300, 316} {
		d := bytes.Repeat(someutf8, i/len(someutf8))
		examples = append(examples, string(d))
	}

	for _, tt := range examples {
		t.Run(tt, func(t *testing.T) {
			check(t, []byte(tt))
		})

		// Generate variations of the input to exercise errors at the
		// boundary, using the vector implementation on 32-sized input,
		// and on non-32-sized inputs.
		//
		// Large examples don't go through those variations because they
		// are likely specific tests.

		if len(tt) >= 32 {
			continue
		}

		t.Run("boundary-"+tt, func(t *testing.T) {
			size := 32 - len(tt)
			prefix := strings.Repeat("a", size)
			b := []byte(prefix + tt)
			check(t, b)
		})
		t.Run("vec-padded-"+tt, func(t *testing.T) {
			prefix := strings.Repeat("a", 32)
			padding := strings.Repeat("b", 32-(len(tt)%32))
			input := prefix + padding + tt
			b := []byte(input)
			if len(b)%32 != 0 {
				panic("test should generate block of 32")
			}
			check(t, b)
		})
		t.Run("vec-"+tt, func(t *testing.T) {
			prefix := strings.Repeat("a", 32)
			input := prefix + tt
			if len(tt)%32 == 0 {
				input += "x"
			}
			b := []byte(input)
			if len(b)%32 == 0 {
				panic("test should not generate block of 32")
			}
			check(t, b)
		})
	}
}

func TestValidPageBoundary(t *testing.T) {
	buf, err := buffer.New(64)
	if err != nil {
		t.Fatal(err)
	}
	defer buf.Release()

	head := buf.ProtectHead()
	tail := buf.ProtectTail()

	data := bytes.Repeat(someutf8, 64/len(someutf8))

	copy(head, data)
	copy(tail, data)

	for i := 0; i <= 32; i++ {
		input := head[:i]
		check(t, input)
	}

	for i := 0; i <= 32; i++ {
		input := tail[i:]
		check(t, input)
	}
}

func check(t *testing.T, b []byte) {
	t.Helper()

	// Check that both Valid and Validate behave properly. Should not be
	// necessary given the definition of Valid, but just in case.

	expected := utf8.Valid(b)
	if Valid(b) != expected {
		err := ioutil.WriteFile("test.out.txt", b, 0600)
		if err != nil {
			panic(err)
		}

		t.Errorf("Valid(%q) = %v; want %v", string(b), !expected, expected)
	}

	v := Validate(b)

	if v.IsUTF8() != expected {
		t.Errorf("Validate(%q) utf8 valid: %v; want %v", string(b), !expected, expected)
	}

	expected = ascii.Valid(b)
	if v.IsASCII() != expected {
		t.Errorf("Validate(%q) ascii valid: %v; want %v", string(b), !expected, expected)
	}
}

var valid1k = bytes.Repeat([]byte("0123456789日本語日本語日本語日abcdefghijklmnopqrstuvwx"), 16)
var valid1M = bytes.Repeat(valid1k, 1024)
var someutf8 = []byte("\xF4\x8F\xBF\xBF")

func BenchmarkValid(b *testing.B) {
	impls := map[string]func([]byte) bool{
		"AVX":    Valid,
		"Stdlib": utf8.Valid,
	}

	type input struct {
		name string
		data []byte
	}
	inputs := []input{
		{"1kValid", valid1k},
		{"1MValid", valid1M},
		{"10ASCII", []byte("0123456789")},
		{"10Japan", []byte("日本語日本語日本語日")},
	}

	const KiB = 1024
	const MiB = 1048576

	for i := 0; i <= 400/len(someutf8); i++ {
		//	for _, i := range []int{1 * KiB, 8 * KiB, 16 * KiB, 64 * KiB, 1 * MiB, 8 * MiB, 32 * MiB, 64 * MiB} {
		d := bytes.Repeat(someutf8, i)
		inputs = append(inputs, input{
			name: fmt.Sprintf("small%d", len(d)),
			data: d,
		})
	}

	for _, i := range []int{300, 316} {
		d := bytes.Repeat(someutf8, i/len(someutf8))
		inputs = append(inputs, input{
			name: fmt.Sprintf("tail%d", len(d)),
			data: d,
		})
	}

	for _, input := range inputs {
		for implName, f := range impls {
			testName := fmt.Sprintf("%s/%s", input.name, implName)

			b.Run(testName, func(b *testing.B) {
				b.SetBytes(int64(len(input.data)))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					f(input.data)
				}
			})
		}
	}
}
