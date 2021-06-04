package qsort

import "unsafe"

// The amount of stack space reserved for the scratch buffer.
const stackSize = 1024

func hybridQuicksort(data []byte, size int) {
	var tmp [stackSize]byte

	switch size {
	case 16:
		hybridQuicksort16(unsafeBytesTo128(data), unsafeBytesTo128(tmp[:]), 0, len(data)/16-1)
	case 32:
		hybridQuicksort32(unsafeBytesTo256(data), unsafeBytesTo256(tmp[:]), 0, len(data)/32-1)
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

func hybridQuicksort16(data, tmp []uint128, lo, hi int) {
	for lo < hi {
		if hi-lo < smallCutoff/16*2 {
			insertionsort16(ptr16(data), lo, hi)
			return
		}
		mid := lo + (hi-lo)/2
		medianOfThree128(data, mid, lo, hi, nil)
		p := hybridPartition16(data, tmp, lo, hi)
		if p-lo < hi-p {
			hybridQuicksort16(data, tmp, lo, p-1)
			lo = p + 1
		} else {
			hybridQuicksort16(data, tmp, p+1, hi)
			hi = p - 1
		}
	}
}

func hybridQuicksort32(data, tmp []uint256, lo, hi int) {
	for lo < hi {
		if hi-lo < smallCutoff/32*2 {
			insertionsort32(ptr32(data), lo, hi)
			return
		}
		mid := lo + (hi-lo)/2
		medianOfThree256(data, mid, lo, hi, nil)
		p := hybridPartition32(data, tmp, lo, hi)
		if p-lo < hi-p {
			hybridQuicksort32(data, tmp, lo, p-1)
			lo = p + 1
		} else {
			hybridQuicksort32(data, tmp, p+1, hi)
			hi = p - 1
		}
	}
}

func hybridPartition16(data, tmp []uint128, lo, hi int) int {
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

func hybridPartition32(data, tmp []uint256, lo, hi int) int {
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
