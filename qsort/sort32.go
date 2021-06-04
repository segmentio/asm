package qsort

type uint256 struct {
	a uint64 // hi
	b uint64
	c uint64
	d uint64 // lo
}

func quicksort256(data []uint256, lo, hi int, swap func(int, int)) {
	for lo < hi {
		if hi-lo < smallCutoff/24 {
			insertionsort256(data, lo, hi, swap)
			return
		}
		mid := lo + (hi-lo)/2
		medianOfThree256(data, mid, lo, hi, swap)
		p := hoarePartition256(data, lo, hi, swap)
		if p-lo < hi-p { // recurse on the smaller side
			quicksort256(data, lo, p-1, swap)
			lo = p + 1
		} else {
			quicksort256(data, p+1, hi, swap)
			hi = p - 1
		}
	}
}

func insertionsort256(data []uint256, lo, hi int, swap func(int, int)) {
	// Extra superfluous checks have been added to prevent the compiler
	// from adding bounds checks in the inner loop.
	i := lo + 1
	if i < 0 || hi >= len(data) {
		return
	}
	for ; i <= hi; i++ {
		item := data[i]
		for j := i; j > 0 && j > lo; j-- {
			if prev := data[j-1]; !less256(item, prev) {
				break
			}
			swap256(data, j, j-1, swap)
		}
	}
}

func medianOfThree256(data []uint256, a, b, c int, swap func(int, int)) int {
	if less256(data[b], data[a]) {
		swap256(data, a, b, swap)
	}
	if less256(data[c], data[b]) {
		swap256(data, b, c, swap)
		if less256(data[b], data[a]) {
			swap256(data, a, b, swap)
		}
	}
	return b
}

func hoarePartition256(data []uint256, lo, hi int, swap func(int, int)) int {
	// Extra superfluous checks have been added to prevent the compiler
	// from adding bounds checks in the inner loops.
	i, j := lo+1, hi
	pivot := data[lo]
	for i >= 0 && hi < len(data) && j < len(data) {
		for ; i <= hi; i++ {
			if item := data[i]; !less256(item, pivot) {
				break
			}
		}
		for ; j > lo; j-- {
			if item := data[j]; !less256(pivot, item) {
				break
			}
		}
		if i >= j {
			break
		}
		swap256(data, i, j, swap)
		i++
		j--
	}
	swap256(data, lo, j, swap)
	return j
}

func swap256(data []uint256, a, b int, swap func(int, int)) {
	data[a], data[b] = data[b], data[a]
	if swap != nil {
		swap(a, b)
	}
}

func less256(a, b uint256) bool {
	return a.a < b.a || (a.a == b.a && a.b < b.b) || (a.a == b.a && a.b == b.b && a.c < b.c) || (a.a == b.a && a.b == b.b && a.c == b.c && a.d <= b.d)
}
