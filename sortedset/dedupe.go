package sortedset

import (
	"bytes"
)

// Dedupe scans a slice containing contiguous chunks of size n,
// and removes duplicates in place.
func Dedupe(b []byte, n int) []byte {
	if len(b)%n != 0 {
		panic("input length is not a multiple of the item size")
	}
	return b[:dedupe(b, n)]
}

func dedupe(b []byte, n int) int {
	switch n {
	case 1:
		return dedupe1(b)
	case 2:
		return dedupe2(b)
	case 4:
		return dedupe4(b)
	case 8:
		return dedupe8(b)
	case 16:
		return dedupe16(b)
	case 32:
		return dedupe32(b)
	default:
		return dedupeGeneric(b, n)
	}
}

func dedupeGeneric(b []byte, size int) int {
	if len(b) <= size {
		return len(b)
	}
	pos := size
	prev := b[:size]
	for i := size; i < len(b); i += size {
		item := b[i : i+size]
		if !bytes.Equal(prev, item) {
			copy(b[pos:], item)
			pos += size
			prev = item
		}
	}
	return pos
}
