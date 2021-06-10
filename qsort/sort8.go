package qsort

// The threshold at which log-linear sorting methods switch to
// a quadratic (but cache-friendly) method such as insertionsort.
const smallCutoff = 256

func quicksort64(data []uint64, base int, swap func(int, int)) {
	for len(data) > 1 {
		if len(data) <= smallCutoff/8 {
			insertionsort64(data, base, swap)
			//smallsort64(data, swap)
			return
		}
		medianOfThree64(data, base, swap)
		p := hoarePartition64(data, base, swap)
		if p < len(data)-p { // recurse on the smaller side
			quicksort64(data[:p], base, swap)
			data = data[p+1:]
			base = base + p + 1
		} else {
			quicksort64(data[p+1:], base+p+1, swap)
			data = data[:p]
		}
	}
}

func smallsort64(data []uint64, base int, swap func(int, int)) {
	if swap != nil {
		insertionsort64(data, base, swap)
	} else {
		bubblesort64NoSwap2(data)
	}
}

func bubblesort64NoSwap1(data []uint64) {
	for i := len(data); i > 1; i-- {
		max := data[0]

		for j := 1; j < i; j++ {
			y := data[j]
			x := uint64(0)

			if max <= y {
				x = max
			} else {
				x = y
			}

			if max <= y {
				max = y
			}

			data[j-1] = x
		}

		data[i-1] = max
	}
}

func bubblesort64NoSwap2(data []uint64) {
	for i := len(data); i > 1; i -= 2 {
		x := data[0]
		y := data[1]

		if y < x {
			x, y = y, x
		}

		for j := 2; j < i; j++ {
			z := data[j]
			w := uint64(0)
			v := uint64(0)

			if y <= z {
				w = y
			} else {
				w = z
			}

			if y <= z {
				y = z
			}

			if x <= z {
				v = x
			} else {
				v = z
			}

			if x <= z {
				x = w
			}

			data[j-2] = v
		}

		data[i-2] = x
		data[i-1] = y
	}
}

func insertionsort64(data []uint64, base int, swap func(int, int)) {
	for i := 1; i < len(data); i++ {
		item := data[i]
		for j := i; j > 0 && item < data[j-1]; j-- {
			swap64(data, j, j-1, base, swap)
		}
	}
}

func insertionsort64NoSwap(data []uint64) {
	for i := 1; i < len(data); i++ {
		item := data[i]
		for j := i; j > 0 && item < data[j-1]; j-- {
			data[j], data[j-1] = data[j-1], data[j]
		}
	}
}

func medianOfThree64(data []uint64, base int, swap func(int, int)) {
	end := len(data) - 1
	mid := len(data) / 2
	if data[0] < data[mid] {
		swap64(data, mid, 0, base, swap)
	}
	if data[end] < data[0] {
		swap64(data, 0, end, base, swap)
		if data[0] < data[mid] {
			swap64(data, mid, 0, base, swap)
		}
	}
}

func hoarePartition64(data []uint64, base int, swap func(int, int)) int {
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
			swap64(data, i, j, base, swap)
			i++
			j--
		}
		swap64(data, 0, j, base, swap)
	}
	return j
}

func swap64(data []uint64, i, j, base int, swap func(int, int)) {
	data[i], data[j] = data[j], data[i]
	if swap != nil {
		swap(base+i, base+j)
	}
}
