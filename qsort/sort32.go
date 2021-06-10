package qsort

type uint256 struct {
	a uint64 // hi
	b uint64
	c uint64
	d uint64 // lo
}

func quicksort256(data []uint256, base int, swap func(int, int)) {
	for len(data) > 1 {
		if len(data) <= smallCutoff/32 {
			smallsort256(data, base, swap)
			return
		}
		medianOfThree256(data, base, swap)
		p := hoarePartition256(data, base, swap)
		if p < len(data)-p { // recurse on the smaller side
			quicksort256(data[:p], base, swap)
			data = data[p+1:]
			base = base + p + 1
		} else {
			quicksort256(data[p+1:], base+p+1, swap)
			data = data[:p]
		}
	}
}

func smallsort256(data []uint256, base int, swap func(int, int)) {
	if swap != nil {
		insertionsort256(data, base, swap)
	} else {
		bubblesort256NoSwap(data)
	}
}

func bubblesort256NoSwap(data []uint256) {
	for i := len(data); i > 1; i -= 2 {
		x := data[0]
		y := data[1]

		if less256(y, x) {
			x, y = y, x
		}

		for j := 2; j < i; j++ {
			z := data[j]
			w := uint256{}
			v := uint256{}

			if lessOrEqual256(y, z) {
				w = y
			} else {
				w = z
			}

			if lessOrEqual256(y, z) {
				y = z
			}

			if lessOrEqual256(x, z) {
				v = x
			} else {
				v = z
			}

			if lessOrEqual256(x, z) {
				x = w
			}

			data[j-2] = v
		}

		data[i-2] = x
		data[i-1] = y
	}
}

func insertionsort256(data []uint256, base int, swap func(int, int)) {
	for i := 1; i < len(data); i++ {
		item := data[i]
		for j := i; j > 0 && less256(item, data[j-1]); j-- {
			data[j], data[j-1] = data[j-1], data[j]
			swap(base+j, base+j-1)
		}
	}
}

func medianOfThree256(data []uint256, base int, swap func(int, int)) {
	end := len(data) - 1
	mid := len(data) / 2
	if less256(data[0], data[mid]) {
		data[mid], data[0] = data[0], data[mid]
		callswap(base, swap, mid, 0)
	}
	if less256(data[end], data[0]) {
		data[0], data[end] = data[end], data[0]
		callswap(base, swap, 0, end)
		if less256(data[0], data[mid]) {
			data[mid], data[0] = data[0], data[mid]
			callswap(base, swap, mid, 0)
		}
	}
}

func hoarePartition256(data []uint256, base int, swap func(int, int)) int {
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

func less256(a, b uint256) bool {
	return a.a < b.a || (a.a == b.a && a.b < b.b) || (a.a == b.a && a.b == b.b && a.c < b.c) || (a.a == b.a && a.b == b.b && a.c == b.c && a.d <= b.d)
}

func lessOrEqual256(a, b uint256) bool {
	return !less256(b, a)
}
