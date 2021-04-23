package qsort

import (
	"encoding/binary"
)

func quicksort32(data []b32, lo, hi int, swap func(int, int)) {
	for lo < hi {
		if hi-lo < smallCutoff/32 {
			insertionsort32(data, lo, hi, swap)
			return
		}
		mid := lo + (hi-lo)/2
		pivot := medianOfThree32(data, lo, mid, hi, swap)
		p := hoarePartition32(data, lo, hi, pivot, swap)
		if p-lo < hi-p { // recurse on the smaller side
			quicksort32(data, lo, p-1, swap)
			lo = p + 1
		} else {
			quicksort32(data, p+1, hi, swap)
			hi = p - 1
		}
	}
}

func insertionsort32(data []b32, lo, hi int, swap func(int, int)) {
	// Additional superfluous checks have been added to
	// eliminate bounds checks in the inner loops.
	i := lo + 1
	if i < 0 || hi >= len(data) {
		return
	}
	for ; i <= hi; i++ {
		for j := i; j > 0 && j > lo; j-- {
			if !less32(data, j, j-1) {
				break
			}
			swap32(data, j, j-1, swap)
		}
	}
}

func medianOfThree32(data []b32, a, b, c int, swap func(int, int)) int {
	if less32(data, b, a) {
		swap32(data, a, b, swap)
	}
	if less32(data, c, b) {
		swap32(data, b, c, swap)
		if less32(data, b, a) {
			swap32(data, a, b, swap)
		}
	}
	return b
}

func hoarePartition32(data []b32, lo, hi, p int, swap func(int, int)) int {
	swap32(data, lo, p, swap)
	i, j := lo+1, hi
	for {
		for i <= hi && less32(data, i, lo) {
			i++
		}
		for less32(data, lo, j) {
			j--
		}
		if i >= j {
			break
		}
		swap32(data, i, j, swap)
		i++
		j--
	}
	swap32(data, lo, j, swap)
	return j
}

func swap32(data []b32, a, b int, swap func(int, int)) {
	data[a], data[b] = data[b], data[a]
	if swap != nil {
		swap(a, b)
	}
}

func less32(data []b32, a, b int) bool {
	return less32cmp(&data[a], &data[b])
}

func less32cmp(a, b *b32) bool {
	x1 := binary.BigEndian.Uint64(a[:8])
	x2 := binary.BigEndian.Uint64(b[:8])
	if x1 != x2 {
		return x1 < x2
	}
	x1 = binary.BigEndian.Uint64(a[8:16])
	x2 = binary.BigEndian.Uint64(b[8:16])
	if x1 != x2 {
		return x1 < x2
	}
	x1 = binary.BigEndian.Uint64(a[16:24])
	x2 = binary.BigEndian.Uint64(b[16:24])
	if x1 != x2 {
		return x1 < x2
	}
	return binary.BigEndian.Uint64(a[24:]) < binary.BigEndian.Uint64(b[24:])
}
