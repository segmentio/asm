package qsort

type uint192 struct {
	hi  uint64
	mid uint64
	lo  uint64
}

func quicksort192(data []uint192, lo, hi int, swap func(int, int)) {
	for lo < hi {
		if hi-lo < smallCutoff/24 {
			insertionsort192(data, lo, hi, swap)
			return
		}
		mid := lo + (hi-lo)/2
		medianOfThree192(data, mid, lo, hi, swap)
		p := hoarePartition192(data, lo, hi, swap)
		if p-lo < hi-p { // recurse on the smaller side
			quicksort192(data, lo, p-1, swap)
			lo = p + 1
		} else {
			quicksort192(data, p+1, hi, swap)
			hi = p - 1
		}
	}
}

func insertionsort192(data []uint192, lo, hi int, swap func(int, int)) {
	// Extra superfluous checks have been added to prevent the compiler
	// from adding bounds checks in the inner loop.
	i := lo + 1
	if i < 0 || hi >= len(data) {
		return
	}
	for ; i <= hi; i++ {
		item := data[i]
		for j := i; j > 0 && j > lo; j-- {
			if prev := data[j-1]; !less192(item, prev) {
				break
			}
			swap192(data, j, j-1, swap)
		}
	}
}

func medianOfThree192(data []uint192, a, b, c int, swap func(int, int)) int {
	if less192(data[b], data[a]) {
		swap192(data, a, b, swap)
	}
	if less192(data[c], data[b]) {
		swap192(data, b, c, swap)
		if less192(data[b], data[a]) {
			swap192(data, a, b, swap)
		}
	}
	return b
}

func hoarePartition192(data []uint192, lo, hi int, swap func(int, int)) int {
	// Extra superfluous checks have been added to prevent the compiler
	// from adding bounds checks in the inner loops.
	i, j := lo+1, hi
	pivot := data[lo]
	for i >= 0 && hi < len(data) && j < len(data) {
		for ; i <= hi; i++ {
			if item := data[i]; !less192(item, pivot) {
				break
			}
		}
		for ; j > lo; j-- {
			if item := data[j]; !less192(pivot, item) {
				break
			}
		}
		if i >= j {
			break
		}
		swap192(data, i, j, swap)
		i++
		j--
	}
	swap192(data, lo, j, swap)
	return j
}

func swap192(data []uint192, a, b int, swap func(int, int)) {
	data[a], data[b] = data[b], data[a]
	if swap != nil {
		swap(a, b)
	}
}

func less192(a, b uint192) bool {
	return a.hi < b.hi || (a.hi == b.hi && a.mid < b.mid) || (a.hi == b.hi && a.mid == b.mid && a.lo <= b.lo)
}
