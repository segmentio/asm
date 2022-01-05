package utf8

import (
	"unicode/utf8"

	"github.com/segmentio/asm/ascii"
)

type Validation byte

const (
	Invalid = 0
	UTF8    = 0b01
	ASCII   = 0b10 | UTF8
)

func (v Validation) IsASCII() bool { return (v & ASCII) == ASCII }

func (v Validation) IsUTF8() bool { return (v & UTF8) == UTF8 }

func (v Validation) IsInvalid() bool { return v == Invalid }

func validate(p []byte) Validation {
	if ascii.Valid(p) {
		return ASCII
	}
	if utf8.Valid(p) {
		return UTF8
	}
	return Invalid
}
