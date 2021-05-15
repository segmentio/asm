package sortedset

import "bytes"

func Intersect(dst, a, b []byte, size int) []byte {
	if size <= 0 || len(a)%size != 0 || len(b)%size != 0 {
		panic("input lengths must be a multiple of size")
	}
	if cap(dst) < len(a) && cap(dst) < len(b) {
		panic("cap(dst) < min(len(a),len(b))")
	}

	// Fast paths for non-overlapping sets.
	if len(a) == 0 || len(b) == 0 || bytes.Compare(a[len(a)-size:], b[:size]) < 0 || bytes.Compare(b[len(b)-size:], a[:size]) < 0 {
		return dst[:0]
	}

	var pos int
	switch size {
	default:
		pos = intersectGeneric(dst, a, b, size)
	}

	return dst[:pos]
}

func intersectGeneric(dst, a, b []byte, size int) int {
	i, j, k := 0, 0, 0
	for i < len(a) && j < len(b) {
		itemA := a[i : i+size]
		itemB := b[j : j+size]
		switch bytes.Compare(itemA, itemB) {
		case 0:
			copy(dst[k:k+size], itemA)
			i += size
			j += size
			k += size
		case -1:
			i += size
		case 1:
			j += size
		}
	}
	return k
}
