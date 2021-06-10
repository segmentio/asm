package qsort

type uint256 = [4]uint64

type smallsort256 func(data []uint256, base int, swap func(int, int))
type partition256 func(data []uint256, base int, swap func(int, int)) int

func quicksort256(data []uint256, base, cutoff int, smallsort smallsort256, partition partition256, swap func(int, int)) {
	for len(data) > 1 {
		if len(data) <= cutoff/32 {
			smallsort(data, base, swap)
			return
		}
		medianOfThree256(data, base, swap)
		p := partition(data, base, swap)
		if p < len(data)-p { // recurse on the smaller side
			quicksort256(data[:p], base, cutoff, smallsort, partition, swap)
			data = data[p+1:]
			base = base + p + 1
		} else {
			quicksort256(data[p+1:], base+p+1, cutoff, smallsort, partition, swap)
			data = data[:p]
		}
	}
}

func insertionsort256(data []uint256, base int, swap func(int, int)) {
	for i := 1; i < len(data); i++ {
		item := data[i]
		for j := i; j > 0 && less256(item, data[j-1]); j-- {
			data[j], data[j-1] = data[j-1], data[j]
			callswap(base, swap, j, j-1)
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

func hybridPartition256(data, scratch []uint256) int {
	pivot := 0
	lo := 1
	hi := len(data) - 1
	limit := len(scratch)

	p := distributeForward256(data, scratch, limit, lo, hi)
	if hi-p <= limit {
		copy(data[p+1:], scratch[limit-hi+p:])
		data[pivot], data[p] = data[p], data[pivot]
		return p
	}
	lo = p + limit
	for {
		hi = distributeBackward256(data, data[lo+1-limit:], limit, lo, hi) - limit
		if hi < lo {
			p = hi
			break
		}
		lo = distributeForward256(data, data[hi+1:], limit, lo, hi) + limit
		if hi < lo {
			p = lo - limit
			break
		}
	}
	copy(data[p+1:], scratch[:])
	data[pivot], data[p] = data[p], data[pivot]
	return p
}

func hybridPartition256Using(scratch []byte) partition256 {
	return func(data []uint256, base int, swap func(int, int)) int {
		return hybridPartition256(data, unsafeBytesTo256(scratch[:]))
	}
}

func less256(a, b uint256) bool {
	return a[0] < b[0] ||
		(a[0] == b[0] && a[1] < b[1]) ||
		(a[0] == b[0] && a[1] == b[1] && a[2] < b[2]) ||
		(a[0] == b[0] && a[1] == b[1] && a[2] == b[2] && a[3] <= b[3])
}
