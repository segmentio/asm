package qsort

// The amount of stack space reserved for the scratch buffer.
const stackSize = 1024

func hybridQuicksort(data []byte, size int) {
	var tmp [stackSize]byte

	switch size {
	case 16:
		hybridQuicksort128(unsafeBytesToU128(data), unsafeBytesToU128(tmp[:]))
	case 32:
		hybridQuicksort256(unsafeBytesToU256(data), unsafeBytesToU256(tmp[:]))
	default:
		panic("unreachable")
	}
}

func hybridQuicksort128(data, tmp []uint128) {
	for len(data) > 1 {
		if len(data) < smallCutoff/16*2 {
			insertionsort128NoSwap(unsafeU128ToBytes(data))
			return
		}
		medianOfThree128(data, 0, nil)

		p := hybridPartition16(data, tmp)
		if p < len(data)-p {
			hybridQuicksort128(data[:p], tmp)
			data = data[p+1:]
		} else {
			hybridQuicksort128(data[p+1:], tmp)
			data = data[:p]
		}
	}
}

func hybridQuicksort256(data, tmp []uint256) {
	for len(data) > 1 {
		if len(data) < smallCutoff/32*2 {
			insertionsort256NoSwap(unsafeU256ToBytes(data))
			return
		}
		medianOfThree256(data, 0, nil)

		p := hybridPartition32(data, tmp)
		if p < len(data)-p {
			hybridQuicksort256(data[:p], tmp)
			data = data[p+1:]
		} else {
			hybridQuicksort256(data[p+1:], tmp)
			data = data[:p]
		}
	}
}

func hybridPartition16(data, tmp []uint128) int {
	lo := 0
	hi := len(data) - 1

	pivot := lo
	lo++
	p := distributeForward128(unsafeU128Addr(data), unsafeU128Addr(tmp), len(tmp), lo, hi, pivot)
	if hi-p <= len(tmp) {
		copy(data[p+1:], tmp[len(tmp)-hi+p:])
		data[pivot], data[p] = data[p], data[pivot]
		return p
	}
	lo = p + len(tmp)
	for {
		hi = distributeBackward128(unsafeU128Addr(data), unsafeU128Addr(data[lo+1-len(tmp):]), len(tmp), lo, hi, pivot) - len(tmp)
		if hi < lo {
			p = hi
			break
		}
		lo = distributeForward128(unsafeU128Addr(data), unsafeU128Addr(data[hi+1:]), len(tmp), lo, hi, pivot) + len(tmp)
		if hi < lo {
			p = lo - len(tmp)
			break
		}
	}
	copy(data[p+1:], tmp[:])
	data[pivot], data[p] = data[p], data[pivot]
	return p
}

func hybridPartition32(data, tmp []uint256) int {
	lo := 0
	hi := len(data) - 1

	pivot := lo
	lo++
	p := distributeForward256(unsafeU256Addr(data), unsafeU256Addr(tmp), len(tmp), lo, hi, pivot)
	if hi-p <= len(tmp) {
		copy(data[p+1:], tmp[len(tmp)-hi+p:])
		data[pivot], data[p] = data[p], data[pivot]
		return p
	}
	lo = p + len(tmp)
	for {
		hi = distributeBackward256(unsafeU256Addr(data), unsafeU256Addr(data[lo+1-len(tmp):]), len(tmp), lo, hi, pivot) - len(tmp)
		if hi < lo {
			p = hi
			break
		}
		lo = distributeForward256(unsafeU256Addr(data), unsafeU256Addr(data[hi+1:]), len(tmp), lo, hi, pivot) + len(tmp)
		if hi < lo {
			p = lo - len(tmp)
			break
		}
	}
	copy(data[p+1:], tmp[:])
	data[pivot], data[p] = data[p], data[pivot]
	return p
}
