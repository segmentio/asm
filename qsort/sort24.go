package qsort

import (
	"encoding/binary"
)

func quicksort24(data []b24, lo, hi int, swap func(int, int)) {
	for lo < hi {
		if hi-lo < smallCutoff/24 {
			insertionsort24(data, lo, hi, swap)
			return
		}
		mid := lo + (hi-lo)/2
		pivot := medianOfThree24(data, lo, mid, hi, swap)
		p := hoarePartition24(data, lo, hi, pivot, swap)
		if p-lo < hi-p { // recurse on the smaller side
			quicksort24(data, lo, p-1, swap)
			lo = p + 1
		} else {
			quicksort24(data, p+1, hi, swap)
			hi = p - 1
		}
	}
}

func insertionsort24(data []b24, lo, hi int, swap func(int, int)) {
	// Additional superfluous checks have been added to
	// eliminate bounds checks in the inner loops.
	i := lo + 1
	if i < 0 || hi >= len(data) {
		return
	}
	for ; i <= hi; i++ {
		for j := i; j > 0 && j > lo; j-- {
			if !less24(data, j, j-1) {
				break
			}
			swap24(data, j, j-1, swap)
		}
	}
}

func medianOfThree24(data []b24, a, b, c int, swap func(int, int)) int {
	if less24(data, b, a) {
		swap24(data, a, b, swap)
	}
	if less24(data, c, b) {
		swap24(data, b, c, swap)
		if less24(data, b, a) {
			swap24(data, a, b, swap)
		}
	}
	return b
}

func hoarePartition24(data []b24, lo, hi, p int, swap func(int, int)) int {
	swap24(data, lo, p, swap)
	i, j := lo+1, hi
	for {
		for i <= hi && less24(data, i, lo) {
			i++
		}
		for less24(data, lo, j) {
			j--
		}
		if i >= j {
			break
		}
		swap24(data, i, j, swap)
		i++
		j--
	}
	swap24(data, lo, j, swap)
	return j
}

func swap24(data []b24, a, b int, swap func(int, int)) {
	data[a], data[b] = data[b], data[a]
	if swap != nil {
		swap(a, b)
	}
}

func less24(data []b24, a, b int) bool {
	return less24cmp(&data[a], &data[b])
}

func less24cmp(a, b *b24) bool {
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
	return binary.BigEndian.Uint64(a[16:]) < binary.BigEndian.Uint64(b[16:])
}
