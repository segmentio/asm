package dedupe

import (
	"bytes"
)

func Dedupe(b []byte, size int) []byte {
	if size <= 0 || len(b)%size != 0 {
		panic("len(b) % size != 0")
	}
	if len(b) <= size {
		return b
	}

	var pos int
	switch size {
	case 16:
		pos = dedupe16(b)
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
