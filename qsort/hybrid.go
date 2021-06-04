package qsort

// The amount of stack space reserved for the scratch buffer.
const stackSize = 1024

func hybridQuicksort(data []byte, size int) {
	var tmp [stackSize]byte

	switch size {
	case 32:
		hybridQuicksort32(unsafeBytesTo256(data), unsafeBytesTo256(tmp[:]), 0, len(data)/32-1)
	default:
		panic("unreachable")
	}
}

func hybridQuicksort32(data, tmp []uint256, lo, hi int) {
	for lo < hi {
		if hi-lo < smallCutoff/32*2 {
			insertionsort32(data, lo, hi)
			return
		}
		mid := lo + (hi-lo)/2
		medianOfThree32(data, mid, lo, hi)
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

func hybridPartition32(data, tmp []uint256, lo, hi int) int {
	pivot := lo
	lo++
	p := distributeForward32(data, tmp, lo, hi, pivot)
	if hi-p <= len(tmp) {
		copy(data[p+1:], tmp[len(tmp)-hi+p:])
		data[pivot], data[p] = data[p], data[pivot]
		return p
	}
	lo = p + len(tmp)
	for {
		hi = distributeBackward32(data, data[lo-len(tmp)+1:lo+1], lo, hi, pivot) - len(tmp)
		if hi < lo {
			p = hi
			break
		}
		lo = distributeForward32(data, data[hi+1:hi+1+len(tmp)], lo, hi, pivot) + len(tmp)
		if hi < lo {
			p = lo - len(tmp)
			break
		}
	}
	copy(data[p+1:], tmp[:])
	data[pivot], data[p] = data[p], data[pivot]
	return p
}

func insertionsort32(data []uint256, lo, hi int) int

func medianOfThree32(data []uint256, a, b, c int) int

func distributeForward32(data, tmp []uint256, lo, hi, pivot int) int

func distributeBackward32(data, tmp []uint256, lo, hi, pivot int) int
