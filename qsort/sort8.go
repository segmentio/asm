package qsort

import "encoding/binary"

func quicksort8(data []b8, lo, hi int, swap func(int, int)) {
	for lo < hi {
		if hi-lo < smallCutoff/8 {
			insertionsort8(data, lo, hi, swap)
			return
		}
		mid := lo + (hi-lo)/2
		pivot := medianOfThree8(data, lo, mid, hi, swap)
		p := hoarePartition8(data, lo, hi, pivot, swap)
		if p-lo < hi-p { // recurse on the smaller side
			quicksort8(data, lo, p-1, swap)
			lo = p + 1
		} else {
			quicksort8(data, p+1, hi, swap)
			hi = p - 1
		}
	}
}

func insertionsort8(data []b8, lo, hi int, swap func(int, int)) {
	// The following function with manual inlining and some superfluous checks
	// so that the compiler can eliminate bounds checks in the inner loops:
	//
	//     for i := lo + 1; i <= hi; i++ {
	//         for j := i; j > lo && less8(data, j, j-1); j-- {
	// 	           swap8(data, j, j-1, swap)
	//	       }
	//     }
	i := lo + 1
	if i < 0 || hi >= len(data) {
		return
	}
	for ; i <= hi; i++ {
		item := binary.BigEndian.Uint64(data[i][:])
		for j := i; j > 0 && j > lo; j-- {
			if prev := binary.BigEndian.Uint64(data[j-1][:]); item >= prev {
				break
			}
			swap8(data, j, j-1, swap)
		}
	}
}

func medianOfThree8(data []b8, a, b, c int, swap func(int, int)) int {
	if less8(data, b, a) {
		swap8(data, a, b, swap)
	}
	if less8(data, c, b) {
		swap8(data, b, c, swap)
		if less8(data, b, a) {
			swap8(data, a, b, swap)
		}
	}
	return b
}

func hoarePartition8(data []b8, lo, hi, p int, swap func(int, int)) int {
	swap8(data, lo, p, swap)
	i, j := lo+1, hi
	pivot := binary.BigEndian.Uint64(data[lo][:])
	for i >= 0 && hi < len(data) && j < len(data) {
		for ; i <= hi; i++ {
			if item := binary.BigEndian.Uint64(data[i][:]); item >= pivot {
				break
			}
		}
		for ; j >= lo; j-- {
			if item := binary.BigEndian.Uint64(data[j][:]); pivot >= item {
				break
			}
		}
		if i >= j {
			break
		}
		swap8(data, i, j, swap)
		i++
		j--
	}
	swap8(data, lo, j, swap)
	return j
}

func swap8(data []b8, a, b int, swap func(int, int)) {
	data[a], data[b] = data[b], data[a]
	if swap != nil {
		swap(a, b)
	}
}

func less8(data []b8, a, b int) bool {
	return less8cmp(&data[a], &data[b])
}

func less8cmp(a, b *b8) bool {
	x1 := binary.BigEndian.Uint64(a[:8])
	x2 := binary.BigEndian.Uint64(b[:8])
	return x1 < x2
}
