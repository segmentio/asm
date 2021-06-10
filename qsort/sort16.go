package qsort

type uint128 struct {
	hi uint64
	lo uint64
}

type smallsort128 func(data []uint128, base int, swap func(int, int))
type partition128 func(data []uint128, base int, swap func(int, int)) int

func quicksort128(data []uint128, base, cutoff int, smallsort smallsort128, partition partition128, swap func(int, int)) {
	for len(data) > 1 {
		if len(data) <= cutoff/16 {
			smallsort(data, base, swap)
			return
		}
		medianOfThree128(data, base, swap)
		p := partition(data, base, swap)
		if p < len(data)-p { // recurse on the smaller side
			quicksort128(data[:p], base, cutoff, smallsort, partition, swap)
			data = data[p+1:]
			base = base + p + 1
		} else {
			quicksort128(data[p+1:], base+p+1, cutoff, smallsort, partition, swap)
			data = data[:p]
		}
	}
}

func insertionsort128(data []uint128, base int, swap func(int, int)) {
	for i := 1; i < len(data); i++ {
		item := data[i]
		for j := i; j > 0 && less128(item, data[j-1]); j-- {
			data[j], data[j-1] = data[j-1], data[j]
			callswap(base, swap, j, j-1)
		}
	}
}

func insertionsort128NoSwap(data []uint128, base int, swap func(int, int)) {
	// FIXME: how can we define the function below as func(data []uint128, base int, swap func(int, int))
	//  to avoid the indirection? It fails because uint128 isn't available at Avo build time
	insertionsort128NoSwapAsm(unsafeU128ToBytes(data))
}

func medianOfThree128(data []uint128, base int, swap func(int, int)) {
	end := len(data) - 1
	mid := len(data) / 2
	if less128(data[0], data[mid]) {
		data[mid], data[0] = data[0], data[mid]
		callswap(base, swap, mid, 0)
	}
	if less128(data[end], data[0]) {
		data[0], data[end] = data[end], data[0]
		callswap(base, swap, 0, end)
		if less128(data[0], data[mid]) {
			data[mid], data[0] = data[0], data[mid]
			callswap(base, swap, mid, 0)
		}
	}
}

func hoarePartition128(data []uint128, base int, swap func(int, int)) int {
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

func hybridPartition128(data, scratch []uint128) int {
	lo := 0
	hi := len(data) - 1

	pivot := lo
	lo++
	p := distributeForward128(unsafeU128Addr(data), unsafeU128Addr(scratch), len(scratch), lo, hi, pivot)
	if hi-p <= len(scratch) {
		copy(data[p+1:], scratch[len(scratch)-hi+p:])
		data[pivot], data[p] = data[p], data[pivot]
		return p
	}
	lo = p + len(scratch)
	for {
		hi = distributeBackward128(unsafeU128Addr(data), unsafeU128Addr(data[lo+1-len(scratch):]), len(scratch), lo, hi, pivot) - len(scratch)
		if hi < lo {
			p = hi
			break
		}
		lo = distributeForward128(unsafeU128Addr(data), unsafeU128Addr(data[hi+1:]), len(scratch), lo, hi, pivot) + len(scratch)
		if hi < lo {
			p = lo - len(scratch)
			break
		}
	}
	copy(data[p+1:], scratch[:])
	data[pivot], data[p] = data[p], data[pivot]
	return p
}

func hybridPartition128Using(scratch []byte) partition128 {
	return func(data []uint128, base int, swap func(int, int)) int {
		return hybridPartition128(data, unsafeBytesToU128(scratch[:]))
	}
}

func less128(a, b uint128) bool {
	return a.hi < b.hi || (a.hi == b.hi && a.lo <= b.lo)
}
