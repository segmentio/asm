package qsort

type uint128 struct {
	hi uint64
	lo uint64
}

func quicksort128(data []uint128, lo, hi int, swap func(int, int)) {
	for lo < hi {
		if hi-lo < smallCutoff/16 {
			insertionsort128(data, lo, hi, swap)
			return
		}
		mid := lo + (hi-lo)/2
		pivot := medianOfThree128(data, mid, lo, hi, swap)
		p := hoarePartition128(data, lo, hi, pivot, swap)
		if p-lo < hi-p { // recurse on the smaller side
			quicksort128(data, lo, p-1, swap)
			lo = p + 1
		} else {
			quicksort128(data, p+1, hi, swap)
			hi = p - 1
		}
	}
}

func insertionsort128(data []uint128, lo, hi int, swap func(int, int)) {
	// Extra superfluous checks have been added to prevent the compiler
	// from adding bounds checks in the inner loop.
	i := lo + 1
	if i < 0 || hi >= len(data) {
		return
	}
	for ; i <= hi; i++ {
		item := data[i]
		for j := i; j > 0 && j > lo; j-- {
			if prev := data[j-1]; !less128(item, prev) {
				break
			}
			swap128(data, j, j-1, swap)
		}
	}
}

func medianOfThree128(data []uint128, a, b, c int, swap func(int, int)) int {
	if less128(data[b], data[a]) {
		swap128(data, a, b, swap)
	}
	if less128(data[c], data[b]) {
		swap128(data, b, c, swap)
		if less128(data[b], data[a]) {
			swap128(data, a, b, swap)
		}
	}
	return b
}

func hoarePartition128(data []uint128, lo, hi, p int, swap func(int, int)) int {
	// Extra superfluous checks have been added to prevent the compiler
	// from adding bounds checks in the inner loops.
	i, j := lo+1, hi
	pivot := data[lo]
	for i >= 0 && hi < len(data) && j < len(data) {
		for ; i <= hi; i++ {
			if item := data[i]; !less128(item, pivot) {
				break
			}
		}
		for ; j >= lo; j-- {
			if item := data[j]; !less128(pivot, item) {
				break
			}
		}
		if i >= j {
			break
		}
		swap128(data, i, j, swap)
		i++
		j--
	}
	swap128(data, lo, j, swap)
	return j
}

func swap128(data []uint128, a, b int, swap func(int, int)) {
	data[a], data[b] = data[b], data[a]
	if swap != nil {
		swap(a, b)
	}
}

func less128(a, b uint128) bool {
	return a.hi < b.hi || (a.hi == b.hi && a.lo <= b.lo)
}
