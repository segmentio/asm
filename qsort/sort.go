package qsort

import (
	"sort"

	"github.com/segmentio/asm/bswap"
	"github.com/segmentio/asm/cpu"
	"github.com/segmentio/asm/internal"
)

// Sort sorts contiguous big-endian chunks of bytes of a fixed size.
// Sorting specializations are available for sizes of 8, 16, 24 and 32 bytes.
func Sort(data []byte, size int, swap func(int, int)) {
	if len(data) <= size {
		return
	}
	if size <= 0 || !internal.MultipleOf(size, len(data)) {
		panic("input length is not a multiple of element size")
	}

	// No specialization available. Use the slower generic sorting routine.
	if size%8 != 0 || size > 32 {
		sort.Sort(newGeneric(data, size, swap))
		return
	}

	// Byte swap each qword prior to sorting. Doing a single pass here, and
	// again after the sort, is faster than byte swapping during each
	// comparison. The sorting routines have been written to assume that high
	// qwords come before low qwords, and so we're able to use the same
	// Swap64() routine rather than needing separate byte swapping routines
	// for 8, 16, 24, or 32 bytes.
	bswap.Swap64(data)
	defer bswap.Swap64(data)

	// If no indirect swapping is required, try to use the hybrid partitioning scheme from
	// https://blog.reverberate.org/2020/05/29/hoares-rebuttal-bubble-sorts-comeback.html
	if swap == nil && size != 24 && cpu.X86.Has(cpu.AVX2) {
		hybridQuicksort(data, size)
		return
	}

	switch size {
	case 8:
		quicksort64(unsafeBytesTo64(data), 0, smallCutoff, insertionsort64, hoarePartition64, swap)
	case 16:
		quicksort128(unsafeBytesTo128(data), 0, smallCutoff, insertionsort128, hoarePartition128, swap)
	case 24:
		quicksort192(unsafeBytesTo192(data), 0, smallCutoff, insertionsort192, hoarePartition192, swap)
	case 32:
		quicksort256(unsafeBytesTo256(data), 0, smallCutoff, insertionsort256, hoarePartition256, swap)
	}
}

func hybridQuicksort(data []byte, size int) {
	// The hybrid Lomuto/Hoare partition scheme at https://blog.reverberate.org/2020/05/29/hoares-rebuttal-bubble-sorts-comeback.html
	// requires scratch space. We allocate some stack space for the task here,
	// and we do it outside the main Sort() function so that we don't pay the
	// stack cost unless necessary.
	var scratch [1024]byte

	// For reasons unknown, using the smallsort sooner (with larger chunks)
	// yields better results in some cases.
	cutoff := smallCutoff * 2

	switch size {
	case 8:
		quicksort64(unsafeBytesTo64(data), 0, smallCutoff, bubblesort64NoSwap2, hybridPartition64Using(scratch[:]), nil)
	case 16:
		quicksort128(unsafeBytesTo128(data), 0, cutoff, insertionsort128NoSwap, hybridPartition128Using(scratch[:]), nil)
	case 32:
		quicksort256(unsafeBytesTo256(data), 0, cutoff, insertionsort256NoSwap, hybridPartition256Using(scratch[:]), nil)
	}
}

// The threshold at which log-linear sorting methods switch to
// a quadratic (but cache-friendly) method such as insertionsort.
const smallCutoff = 256

func callswap(base int, swap func(int, int), i, j int) {
	if swap != nil {
		swap(base+i, base+j)
	}
}
