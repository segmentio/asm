package qsort

type uint128 struct {
	hi uint64
	lo uint64
}

func quicksort128(data []uint128, swap func(int, int)) {
	for len(data) > 1 {
		if len(data) < smallCutoff/16 {
			insertionsort128(data, swap)
			return
		}
		medianOfThree128(data, swap)
		p := hoarePartition128(data, swap)
		if p < len(data)-p { // recurse on the smaller side
			quicksort128(data[:p], swap)
			data = data[p+1:]
		} else {
			quicksort128(data[p+1:], swap)
			data = data[:p]
		}
	}
}

func insertionsort128(data []uint128, swap func(int, int)) {
	for i := 1; i < len(data); i++ {
		item := data[i]
		for j := i; j > 0 && less128(item, data[j-1]); j-- {
			swap128(data, j, j-1, swap)
		}
	}
}

func medianOfThree128(data []uint128, swap func(int, int)) {
	end := len(data) - 1
	mid := len(data) / 2
	if less128(data[0], data[mid]) {
		swap128(data, mid, 0, swap)
	}
	if less128(data[end], data[0]) {
		swap128(data, 0, end, swap)
		if less128(data[0], data[mid]) {
			swap128(data, mid, 0, swap)
		}
	}
}

func hoarePartition128(data []uint128, swap func(int, int)) int {
	i, j := 1, len(data)-1
	if len(data) > 0 {
		pivot := data[0]
		for j < len(data) {
			for i < len(data) && less128(data[i], pivot) {
				i++
			}
			for j > 0 && less128(pivot, data[j]) {
				j--
			}
			if i >= j {
				break
			}
			swap128(data, i, j, swap)
			i++
			j--
		}
		swap128(data, 0, j, swap)
	}
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
