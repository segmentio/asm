package qsort

// The threshold at which log-linear sorting methods switch to
// a quadratic (but cache-friendly) method such as insertionsort.
const smallCutoff = 256

func quicksort64(data []uint64, lo, hi int, swap func(int, int)) {
	for lo < hi {
		if hi-lo < smallCutoff/8 {
			insertionsort64(data, lo, hi, swap)
			return
		}
		mid := lo + (hi-lo)/2
		medianOfThree64(data, mid, lo, hi, swap)
		p := hoarePartition64(data, lo, hi, swap)
		if p-lo < hi-p { // recurse on the smaller side
			quicksort64(data, lo, p-1, swap)
			lo = p + 1
		} else {
			quicksort64(data, p+1, hi, swap)
			hi = p - 1
		}
	}
}

func insertionsort64(data []uint64, lo, hi int, swap func(int, int)) {
	// Extra superfluous checks have been added to prevent the compiler
	// from adding bounds checks in the inner loop.
	i := lo + 1
	if i < 0 || hi >= len(data) {
		return
	}
	for ; i <= hi; i++ {
		item := data[i]
		for j := i; j > 0 && j > lo; j-- {
			if prev := data[j-1]; item >= prev {
				break
			}
			swap64(data, j, j-1, swap)
		}
	}
}

func medianOfThree64(data []uint64, a, b, c int, swap func(int, int)) int {
	if data[b] < data[a] {
		swap64(data, a, b, swap)
	}
	if data[c] < data[b] {
		swap64(data, b, c, swap)
		if data[b] < data[a] {
			swap64(data, a, b, swap)
		}
	}
	return b
}

func hoarePartition64(data []uint64, lo, hi int, swap func(int, int)) int {
	// Extra superfluous checks have been added to prevent the compiler
	// from adding bounds checks in the inner loops.
	i, j := lo+1, hi
	pivot := data[lo]
	for i >= 0 && hi < len(data) && j < len(data) {
		for ; i <= hi; i++ {
			if item := data[i]; item >= pivot {
				break
			}
		}
		for ; j >= lo; j-- {
			if item := data[j]; pivot >= item {
				break
			}
		}
		if i >= j {
			break
		}
		swap64(data, i, j, swap)
		i++
		j--
	}
	swap64(data, lo, j, swap)
	return j
}

func swap64(data []uint64, a, b int, swap func(int, int)) {
	data[a], data[b] = data[b], data[a]
	if swap != nil {
		swap(a, b)
	}
}
