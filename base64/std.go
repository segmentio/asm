package base64

import (
	"encoding/base64"

	"github.com/segmentio/asm/cpu"
)

var simd struct {
	stdEncodeNative func(dst, src []byte) (int, int)
}

func init() {
	if cpu.X86.Has(cpu.AVX2) {
		simd.stdEncodeNative = stdEncodeAVX2
	}
}

func StdEncode(dst, src []byte) {
	if len(src) >= 28 && simd.stdEncodeNative != nil {
		d, s := simd.stdEncodeNative(dst, src)
		dst = dst[d:]
		src = src[s:]
	}
	base64.StdEncoding.Encode(dst, src)
}
