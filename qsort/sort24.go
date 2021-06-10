package qsort

type uint192 struct {
	hi  uint64
	mid uint64
	lo  uint64
}

func quicksort192(data []uint192, base int, swap func(int, int)) {
	for len(data) > 1 {
		if len(data) <= smallCutoff/24 {
			smallsort192(data, base, swap)
			return
		}
		medianOfThree192(data, base, swap)
		p := hoarePartition192(data, base, swap)
		if p < len(data)-p { // recurse on the smaller side
			quicksort192(data[:p], base, swap)
			data = data[p+1:]
			base = base + p + 1
		} else {
			quicksort192(data[p+1:], base+p+1, swap)
			data = data[:p]
		}
	}
}

func smallsort192(data []uint192, base int, swap func(int, int)) {
	if swap != nil {
		insertionsort192(data, base, swap)
	} else {
		bubblesort192NoSwap2(data)
	}
}

func bubblesort192NoSwap1(data []uint192) {
	for i := len(data); i > 1; i-- {
		max := data[0]

		for j := 1; j < i; j++ {
			y := data[j]
			x := uint192{}

			if lessOrEqual192(max, y) {
				x = max
			} else {
				x = y
			}

			if lessOrEqual192(max, y) {
				max = y
			}

			data[j-1] = x
		}

		data[i-1] = max
	}
}

func bubblesort192NoSwap2(data []uint192) {
	for i := len(data); i > 1; i -= 2 {
		x := data[0]
		y := data[1]

		if less192(y, x) {
			x, y = y, x
		}

		for j := 2; j < i; j++ {
			z := data[j]
			w := uint192{}
			v := uint192{}

			if lessOrEqual192(y, z) {
				w = y
			} else {
				w = z
			}

			if lessOrEqual192(y, z) {
				y = z
			}

			if lessOrEqual192(x, z) {
				v = x
			} else {
				v = z
			}

			if lessOrEqual192(x, z) {
				x = w
			}

			data[j-2] = v
		}

		data[i-2] = x
		data[i-1] = y
	}
}

func insertionsort192(data []uint192, base int, swap func(int, int)) {
	for i := 1; i < len(data); i++ {
		item := data[i]
		for j := i; j > 0 && less192(item, data[j-1]); j-- {
			data[j], data[j-1] = data[j-1], data[j]
			swap(base+j, base+j-1)
		}
	}
}

func medianOfThree192(data []uint192, base int, swap func(int, int)) {
	end := len(data) - 1
	mid := len(data) / 2
	if less192(data[0], data[mid]) {
		data[mid], data[0] = data[0], data[mid]
		callswap(base, swap, mid, 0)
	}
	if less192(data[end], data[0]) {
		data[0], data[end] = data[end], data[0]
		callswap(base, swap, 0, end)
		if less192(data[0], data[mid]) {
			data[mid], data[0] = data[0], data[mid]
			callswap(base, swap, mid, 0)
		}
	}
}

func hoarePartition192(data []uint192, base int, swap func(int, int)) int {
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
			data[i], data[j] = data[j], data[i]
			callswap(base, swap, i, j)
			i++
			j--
		}
		data[0], data[j] = data[j], data[0]
		callswap(base, swap, 0, j)
	}
	return j
}

func less192(a, b uint192) bool {
	return a.hi < b.hi || (a.hi == b.hi && a.mid < b.mid) || (a.hi == b.hi && a.mid == b.mid && a.lo <= b.lo)
}

func lessOrEqual192(a, b uint192) bool {
	return !less192(b, a)
}
