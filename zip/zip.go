package zip

import "github.com/segmentio/asm/cpu"

// SumUint64 sums pairs of by index from x and y, similar to python's zip routine.
// If available AVX instructions will be used to operate on many uint64s simultaneously.
//
// Results are returned in the x slice and y is left unaltered. If x and y differ in size
// only len(x) elements will be processed.
func SumUint64(x []uint64, y []uint64) {
	switch {
	case cpu.X86.Has(cpu.AVX):
		sumUint64(x, y)
	default:
		sumUint64Generic(x, y)
	}
}

func sumUint64Generic(x, y []uint64) {
	for i:=0;i<len(x)&&i<len(y); i++ {
		x[i] = x[i] + y[i]
	}
}