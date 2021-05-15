package sortedset

import "bytes"

func Union(dst, a, b []byte, size int) []byte {
	if size <= 0 || len(a)%size != 0 || len(b)%size != 0 {
		panic("input lengths must be a multiple of size")
	}
	if cap(dst) < len(a)+len(b) {
		panic("cap(dst) < len(a)+len(b)")
	}

	// Fast paths for non-overlapping sets.
	switch {
	case len(a) == 0:
		return dst[:copy(dst[:cap(dst)], b)]
	case len(b) == 0:
		return dst[:copy(dst[:cap(dst)], a)]
	case bytes.Compare(a[len(a)-size:], b[:size]) < 0:
		k := copy(dst[:len(a)], a)
		k += copy(dst[k:k+len(b)], b)
		return dst[:k]
	case bytes.Compare(b[len(b)-size:], a[:size]) < 0:
		k := copy(dst[:len(b)], b)
		k += copy(dst[k:k+len(a)], a)
		return dst[:k]
	}

	var pos int
	switch size {
	default:
		pos = unionGeneric(dst, a, b, size)
	}

	return dst[:pos]
}

func unionGeneric(dst, a, b []byte, size int) int {
	i, j, k := 0, 0, 0
	for i < len(a) && j < len(b) {
		itemA := a[i : i+size]
		itemB := b[j : j+size]
		switch bytes.Compare(itemA, itemB) {
		case 0:
			copy(dst[k:k+size], itemA)
			i += size
			j += size
		case -1:
			copy(dst[k:k+size], itemA)
			i += size
		case 1:
			copy(dst[k:k+size], itemB)
			j += size
		}
		k += size
	}
	if i < len(a) {
		k += copy(dst[k:k+len(a)-i], a[i:])
	} else if j < len(b) {
		k += copy(dst[k:k+len(b)-j], b[j:])
	}
	return k
}
