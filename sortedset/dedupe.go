package sortedset

import (
	"bytes"

	"github.com/segmentio/asm/cpu"
	"github.com/segmentio/asm/internal"
)

// Dedupe scans a slice containing contiguous chunks of a specific size,
// and removes duplicates in place.
func Dedupe(b []byte, size int) []byte {
	if len(b) <= size {
		return b
	}
	if size <= 0 || !internal.MultipleOf(size, len(b)) {
		panic("len(b) % size != 0")
	}

	var pos int
	switch {
	case size == 16 && cpu.X86.Has(cpu.SSE4):
		pos = dedupe16(b)
	case size == 32 && cpu.X86.Has(cpu.AVX2):
		pos = dedupe32(b)
	default:
		pos = dedupeGeneric(b, size)
	}

	return b[:pos]
}

func dedupeGeneric(b []byte, size int) int {
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
