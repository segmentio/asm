package qsort

// The threshold at which log-linear sorting methods switch to
// a quadratic (but cache-friendly) method such as insertionsort.
const smallCutoff = 256

func quicksort64(data []uint64, swap func(int, int)) {
	for len(data) > 1 {
		if len(data) < smallCutoff/8 {
			insertionsort64(data, swap)
			return
		}
		medianOfThree64(data, swap)
		p := hoarePartition64(data, swap)
		if p < len(data)-p { // recurse on the smaller side
			quicksort64(data[:p], swap)
			data = data[p+1:]
		} else {
			quicksort64(data[p+1:], swap)
			data = data[:p]
		}
	}
}

func insertionsort64(data []uint64, swap func(int, int)) {
	for i := 1; i < len(data); i++ {
		item := data[i]
		for j := i; j > 0 && item < data[j-1]; j-- {
			swap64(data, j, j-1, swap)
		}
	}
}

func medianOfThree64(data []uint64, swap func(int, int)) {
	if len(data) > 0 {
		end := len(data) - 1
		mid := len(data) / 2
		if data[0] < data[mid] {
			swap64(data, mid, 0, swap)
		}
		if data[end] < data[0] {
			swap64(data, 0, end, swap)
			if data[0] < data[mid] {
				swap64(data, mid, 0, swap)
			}
		}
	}
}

func hoarePartition64(data []uint64, swap func(int, int)) int {
	i, j := 1, len(data)-1
	if len(data) > 0 {
		pivot := data[0]
		for j < len(data) {
			for i < len(data) && data[i] < pivot {
				i++
			}
			for j > 0 && pivot < data[j] {
				j--
			}
			if i >= j {
				break
			}
			swap64(data, i, j, swap)
			i++
			j--
		}
		swap64(data, 0, j, swap)
	}
	return j
}

func swap64(data []uint64, a, b int, swap func(int, int)) {
	data[a], data[b] = data[b], data[a]
	if swap != nil {
		swap(a, b)
	}
}
