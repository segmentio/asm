//go:build arm64 && !purego
// +build arm64,!purego

package base64

import "encoding/base64"

const (
	encLutSize   = 16
	decLutSize   = 48
	minEncodeLen = 16 * 3
	minDecodeLen = 8 * 4
)

func newEncoding(encoder string) *Encoding {
	e := &Encoding{base: base64.NewEncoding(encoder)}
	e.enableEncodeARM64(encoder)
	return e
}

func (e *Encoding) enableEncodeARM64(encoder string) {
	c62, c63 := int8(encoder[62]), int8(encoder[63])
	tab := [encLutSize]int8{
		'a' - 26, '0' - 52, '0' - 52, '0' - 52, '0' - 52, '0' - 52, '0' - 52, '0' - 52,
		'0' - 52, '0' - 52, '0' - 52, c62 - 62, c63 - 63, 'A', 0, 0,
	}

	e.enc = encodeARM64
	e.enclut = tab
}
