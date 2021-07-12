package base64

import (
	"encoding/base64"

	"github.com/segmentio/asm/cpu"
)

type Encoding struct {
	enc  [16]int8
	dec  [32]int8
	base *base64.Encoding
}

const (
	minEncodeLen = 28
	minDecodeLen = 45
)

func newEncoding(encoder string) *Encoding {
	// Translate values 0..63 to the Base64 alphabet. There are five sets:
	//
	// From      To         Add    Index  Example
	// [0..25]   [65..90]   +65        0  ABCDEFGHIJKLMNOPQRSTUVWXYZ
	// [26..51]  [97..122]  +71        1  abcdefghijklmnopqrstuvwxyz
	// [52..61]  [48..57]    -4  [2..11]  0123456789
	// [62]      [43]       -19       12  +
	// [63]      [47]       -16       13  /
	enc := [16]int8{int8(encoder[0]), int8(encoder[letterRange]) - letterRange}
	for i, ch := range encoder[2*letterRange:] {
		enc[2+i] = int8(ch) - 2*letterRange - int8(i)
	}

	// Translate values from the Base64 alphabet using five sets. Values outside
	// of these ranges are considered invalid:
	//
	// From       To        Add    Index  Example
	// [47]       [63]      +16        1  /
	// [43]       [62]      +19        2  +
	// [48..57]   [52..61]   +4        3  0123456789
	// [65..90]   [0..25]   -65      4,5  ABCDEFGHIJKLMNOPQRSTUVWXYZ
	// [97..122]  [26..51]  -71      6,7  abcdefghijklmnopqrstuvwxyz
	dec := [32]int8{
		0,
		prefixRange + 1 - int8(encoder[prefixRange+1]),
		prefixRange - int8(encoder[prefixRange]),
		2*letterRange - int8(encoder[2*letterRange]),
		0 - int8(encoder[0]),
		0 - int8(encoder[0]),
		letterRange - int8(encoder[letterRange]),
		letterRange - int8(encoder[letterRange]),
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x15, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11,
		0x11, 0x11, 0x13, 0x1B, 0x1B, 0x1B, 0x1B, 0x1B,
	}
	dec[encoder[62]&15] = 0x1A
	dec[encoder[63]&15] = 0x1A
	dec[(encoder[62]&15)+16] = 0x1A
	dec[(encoder[63]&15)+16] = 0x1A

	return &Encoding{
		enc:  enc,
		dec:  dec,
		base: base64.NewEncoding(encoder),
	}
}

func (enc Encoding) WithPadding(padding rune) *Encoding {
	enc.base = enc.base.WithPadding(padding)
	return &enc
}

func (enc Encoding) Strict() *Encoding {
	enc.base = enc.base.Strict()
	return &enc
}

func (enc *Encoding) Encode(dst, src []byte) {
	if len(src) >= minEncodeLen && cpu.X86.Has(cpu.AVX2) {
		d, s := encodeAVX2(dst, src, enc.enc)
		dst = dst[d:]
		src = src[s:]
	}
	enc.base.Encode(dst, src)
}

func (enc *Encoding) EncodeToString(src []byte) string {
	buf := make([]byte, enc.base.EncodedLen(len(src)))
	enc.Encode(buf, src)
	return string(buf)
}

func (enc *Encoding) EncodedLen(n int) int {
	return enc.base.EncodedLen(n)
}

func (enc *Encoding) Decode(dst, src []byte) (n int, err error) {
	var d, s int
	if len(src) >= minDecodeLen && cpu.X86.Has(cpu.AVX2) {
		d, s = decodeAVX2(dst, src, enc.dec)
		dst = dst[d:]
		src = src[s:]
	}
	n, err = enc.base.Decode(dst, src)
	n += d
	return
}

func (enc *Encoding) DecodeString(s string) ([]byte, error) {
	buf := make([]byte, enc.base.DecodedLen(len(s)))
	n, err := enc.Decode(buf, []byte(s))
	return buf[:n], err
}

func (enc *Encoding) DecodedLen(n int) int {
	return enc.base.DecodedLen(n)
}
