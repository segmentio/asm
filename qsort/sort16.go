package qsort

import "encoding/binary"

func quicksort16(data []b16, lo, hi int, swap func(int, int)) {
	for lo < hi {
		if hi-lo < smallCutoff/16 {
			insertionsort16(data, lo, hi, swap)
			return
		}
		mid := lo + (hi-lo)/2
		pivot := medianOfThree16(data, lo, mid, hi, swap)
		p := hoarePartition16(data, lo, hi, pivot, swap)
		if p-lo < hi-p { // recurse on the smaller side
			quicksort16(data, lo, p-1, swap)
			lo = p + 1
		} else {
			quicksort16(data, p+1, hi, swap)
			hi = p - 1
		}
	}
}

func insertionsort16(data []b16, lo, hi int, swap func(int, int)) {
	// The following function with manual inlining and some superfluous checks
	// so that the compiler can eliminate bounds checks in the inner loops:
	//
	//     for i := lo + 1; i <= hi; i++ {
	//         for j := i; j > lo && less16(data, j, j-1); j-- {
	// 	           swap16(data, j, j-1, swap)
	//	       }
	//     }
	i := lo + 1
	if i < 0 || hi >= len(data) {
		return
	}
	for ; i <= hi; i++ {
		item_lo := binary.BigEndian.Uint64(data[i][:8])
		item_hi := binary.BigEndian.Uint64(data[i][8:])
		for j := i; j > 0 && j > lo; j-- {
			if prev_lo := binary.BigEndian.Uint64(data[j-1][:8]); item_lo != prev_lo {
				if item_lo >= prev_lo {
					break
				}
			} else if prev_hi := binary.BigEndian.Uint64(data[j-1][8:]); item_hi >= prev_hi {
				break
			}
			swap16(data, j, j-1, swap)
		}
	}
}

func medianOfThree16(data []b16, a, b, c int, swap func(int, int)) int {
	if less16(data, b, a) {
		swap16(data, a, b, swap)
	}
	if less16(data, c, b) {
		swap16(data, b, c, swap)
		if less16(data, b, a) {
			swap16(data, a, b, swap)
		}
	}
	return b
}

func hoarePartition16(data []b16, lo, hi, p int, swap func(int, int)) int {
	swap16(data, lo, p, swap)
	i, j := lo+1, hi
	pivot_lo := binary.BigEndian.Uint64(data[lo][:8])
	pivot_hi := binary.BigEndian.Uint64(data[lo][8:])
	for i >= 0 && hi < len(data) && j < len(data) {
		for ; i <= hi; i++ {
			if item_lo := binary.BigEndian.Uint64(data[i][:8]); pivot_lo != item_lo {
				if item_lo >= pivot_lo {
					break
				}
			} else if item_hi := binary.BigEndian.Uint64(data[i][8:]); item_hi >= pivot_hi {
				break
			}
		}
		for ; j >= lo; j-- {
			if item_lo := binary.BigEndian.Uint64(data[j][:8]); pivot_lo != item_lo {
				if pivot_lo >= item_lo {
					break
				}
			} else if item_hi := binary.BigEndian.Uint64(data[j][8:]); pivot_hi >= item_hi {
				break
			}
		}
		if i >= j {
			break
		}
		swap16(data, i, j, swap)
		i++
		j--
	}
	swap16(data, lo, j, swap)
	return j
}

func swap16(data []b16, a, b int, swap func(int, int)) {
	data[a], data[b] = data[b], data[a]
	if swap != nil {
		swap(a, b)
	}
}

func less16(data []b16, a, b int) bool {
	return less16cmp(&data[a], &data[b])
}

func less16cmp(a, b *b16) bool {
	x1 := binary.BigEndian.Uint64(a[:8])
	x2 := binary.BigEndian.Uint64(b[:8])
	if x1 != x2 {
		return x1 < x2
	}
	return binary.BigEndian.Uint64(a[8:]) < binary.BigEndian.Uint64(b[8:])
}
