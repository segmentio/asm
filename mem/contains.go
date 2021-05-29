package mem

import (
	"bytes"

	"github.com/klauspost/cpuid/v2"
)

var containsByte = containsByteStd

func init() {
	if cpuid.CPU.Supports(cpuid.AVX2) {
		containsByte = containsByteAVX2
	}
}

func ContainsByte(haystack []byte, needle byte) bool {
	return len(haystack) > 0 && containsByte(haystack, needle)
}

func containsByteStd(haystack []byte, needle byte) bool {
	return bytes.IndexByte(haystack, needle) != -1
}
