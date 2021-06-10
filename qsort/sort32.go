package qsort

type uint256 struct {
	a uint64 // hi
	b uint64
	c uint64
	d uint64 // lo
}

func quicksort256(data []uint256, swap func(int, int)) {
	for len(data) > 1 {
		if len(data) < smallCutoff/16 {
			insertionsort256(data, swap)
			return
		}
		medianOfThree256(data, swap)
		p := hoarePartition256(data, swap)
		if p < len(data)-p { // recurse on the smaller side
			quicksort256(data[:p], swap)
			data = data[p+1:]
		} else {
			quicksort256(data[p+1:], swap)
			data = data[:p]
		}
	}
}

func insertionsort256(data []uint256, swap func(int, int)) {
	for i := 1; i < len(data); i++ {
		item := data[i]
		for j := i; j > 0 && less256(item, data[j-1]); j-- {
			swap256(data, j, j-1, swap)
		}
	}
}

func medianOfThree256(data []uint256, swap func(int, int)) {
	end := len(data) - 1
	mid := len(data) / 2
	if less256(data[0], data[mid]) {
		swap256(data, mid, 0, swap)
	}
	if less256(data[end], data[0]) {
		swap256(data, 0, end, swap)
		if less256(data[0], data[mid]) {
			swap256(data, mid, 0, swap)
		}
	}
}

func hoarePartition256(data []uint256, swap func(int, int)) int {
	i, j := 1, len(data)-1
	if len(data) > 0 {
		pivot := data[0]
		for j < len(data) {
			for i < len(data) && less256(data[i], pivot) {
				i++
			}
			for j > 0 && less256(pivot, data[j]) {
				j--
			}
			if i >= j {
				break
			}
			swap256(data, i, j, swap)
			i++
			j--
		}
		swap256(data, 0, j, swap)
	}
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
