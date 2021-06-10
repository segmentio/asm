package qsort

import "unsafe"

// The amount of stack space reserved for the scratch buffer.
const stackSize = 1024

func hybridQuicksort(data []byte, size int) {
	var tmp [stackSize]byte

	switch size {
	case 16:
		hybridQuicksort16(unsafeBytesToU128(data), unsafeBytesToU128(tmp[:]))
	case 32:
		hybridQuicksort32(unsafeBytesToU256(data), unsafeBytesToU256(tmp[:]))
	default:
		panic("unreachable")
	}
}

func ptr16(slice []uint128) *byte {
	return (*byte)(unsafe.Pointer(&slice[0]))
}

func ptr32(slice []uint256) *byte {
	return (*byte)(unsafe.Pointer(&slice[0]))
}

func hybridQuicksort16(data, tmp []uint128) {
	for len(data) > 1 {
		if len(data) < smallCutoff/16*2 {
			insertionsort16(unsafeU128ToBytes(data))
			return
		}
		medianOfThree128(data, 0, nil)

		p := hybridPartition16(data, tmp)
		if p < len(data)-p {
			hybridQuicksort16(data[:p], tmp)
			data = data[p+1:]
		} else {
			hybridQuicksort16(data[p+1:], tmp)
			data = data[:p]
		}
	}
}

func hybridQuicksort32(data, tmp []uint256) {
	for len(data) > 1 {
		if len(data) < smallCutoff/32*2 {
			insertionsort32(unsafeU256ToBytes(data))
			return
		}
		medianOfThree256(data, 0, nil)

		p := hybridPartition32(data, tmp)
		if p < len(data)-p {
			hybridQuicksort32(data[:p], tmp)
			data = data[p+1:]
		} else {
			hybridQuicksort32(data[p+1:], tmp)
			data = data[:p]
		}
	}
}

func hybridPartition16(data, tmp []uint128) int {
	lo := 0
	hi := len(data) - 1

	pivot := lo
	lo++
	p := distributeForward16(ptr16(data), ptr16(tmp), len(tmp), lo, hi, pivot)
	if hi-p <= len(tmp) {
		copy(data[p+1:], tmp[len(tmp)-hi+p:])
		data[pivot], data[p] = data[p], data[pivot]
		return p
	}
	lo = p + len(tmp)
	for {
		hi = distributeBackward16(ptr16(data), ptr16(data[lo+1-len(tmp):]), len(tmp), lo, hi, pivot) - len(tmp)
		if hi < lo {
			p = hi
			break
		}
		lo = distributeForward16(ptr16(data), ptr16(data[hi+1:]), len(tmp), lo, hi, pivot) + len(tmp)
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
	p := distributeForward32(ptr32(data), ptr32(tmp), len(tmp), lo, hi, pivot)
	if hi-p <= len(tmp) {
		copy(data[p+1:], tmp[len(tmp)-hi+p:])
		data[pivot], data[p] = data[p], data[pivot]
		return p
	}
	lo = p + len(tmp)
	for {
		hi = distributeBackward32(ptr32(data), ptr32(data[lo+1-len(tmp):]), len(tmp), lo, hi, pivot) - len(tmp)
		if hi < lo {
			p = hi
			break
		}
		lo = distributeForward32(ptr32(data), ptr32(data[hi+1:]), len(tmp), lo, hi, pivot) + len(tmp)
		if hi < lo {
			p = lo - len(tmp)
			break
		}
	}
	copy(data[p+1:], tmp[:])
	data[pivot], data[p] = data[p], data[pivot]
	return p
}
