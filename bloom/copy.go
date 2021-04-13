package bloom

import (
	"github.com/klauspost/cpuid/v2"
	"github.com/segmentio/asm/internal/unsafebytes"
)

var simd struct {
	copy64 func(dst, src *byte, n int)
}

func init() {
	if cpuid.CPU.Supports(cpuid.AVX, cpuid.AVX2) {
		simd.copy64 = copyAVX2
	}
}

// Copy performs a OR of src and dst into dst, returning the number of bytes
// written to dst.
func Copy(dst, src []byte) int {
	switch {
	case len(dst) < len(src):
		src = src[:len(dst)]
	case len(dst) > len(src):
		dst = dst[:len(src)]
	}

	n := 0

	if len(dst) >= 64 && simd.copy64 != nil {
		simd.copy64(unsafebytes.Pointer(dst), unsafebytes.Pointer(src), len(dst)/64)
		n = (len(dst) / 64) * 64
	}

	if n >= 0 && n < len(dst) && n < len(src) {
		dst = dst[n:]
		src = src[n:]

		for i := range dst {
			dst[i] |= src[i]
		}

		n += len(dst)
	}

	return n
}
