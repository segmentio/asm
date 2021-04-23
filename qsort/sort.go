package qsort

import "sort"

// The threshold at which log-linear sorting methods switch to
// a quadratic (but cache-friendly) method such as insertionsort.
const smallCutoff = 256 // bytes

func Sort(data []byte, size int, swap func(int, int)) {
	if len(data)/size > 1 {
		switch size {
		case 8:
			quicksort8(unsafeBytesTo8(data), 0, len(data)/8-1, swap)
		case 16:
			quicksort16(unsafeBytesTo16(data), 0, len(data)/16-1, swap)
		case 24:
			quicksort24(unsafeBytesTo24(data), 0, len(data)/24-1, swap)
		case 32:
			quicksort32(unsafeBytesTo32(data), 0, len(data)/32-1, swap)
		default:
			sort.Sort(newGeneric(data, size, swap))
		}
	}
}
