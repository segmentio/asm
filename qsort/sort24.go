package qsort

type uint192 struct {
	hi  uint64
	mid uint64
	lo  uint64
}

func quicksort192(data []uint192, swap func(int, int)) {
	for len(data) > 1 {
		if len(data) < smallCutoff/16 {
			insertionsort192(data, swap)
			return
		}
		medianOfThree192(data, swap)
		p := hoarePartition192(data, swap)
		if p < len(data)-p { // recurse on the smaller side
			quicksort192(data[:p], swap)
			data = data[p+1:]
		} else {
			quicksort192(data[p+1:], swap)
			data = data[:p]
		}
	}
}

func insertionsort192(data []uint192, swap func(int, int)) {
	for i := 1; i < len(data); i++ {
		item := data[i]
		for j := i; j > 0 && less192(item, data[j-1]); j-- {
			swap192(data, j, j-1, swap)
		}
	}
}

func medianOfThree192(data []uint192, swap func(int, int)) {
	end := len(data) - 1
	mid := len(data) / 2
	if less192(data[0], data[mid]) {
		swap192(data, mid, 0, swap)
	}
	if less192(data[end], data[0]) {
		swap192(data, 0, end, swap)
		if less192(data[0], data[mid]) {
			swap192(data, mid, 0, swap)
		}
	}
}

func hoarePartition192(data []uint192, swap func(int, int)) int {
	i, j := 1, len(data)-1
	if len(data) > 0 {
		pivot := data[0]
		for j < len(data) {
			for i < len(data) && less192(data[i], pivot) {
				i++
			}
			for j > 0 && less192(pivot, data[j]) {
				j--
			}
			if i >= j {
				break
			}
			swap192(data, i, j, swap)
			i++
			j--
		}
		swap192(data, 0, j, swap)
	}
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
