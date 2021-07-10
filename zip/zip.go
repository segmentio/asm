package zip

import "github.com/segmentio/asm/cpu"

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