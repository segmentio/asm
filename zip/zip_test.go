package zip

import (
	"math/rand"
	"testing"
)

var size = 1024

func TestSumUint64(t *testing.T) {
	x, y := generateSlices()
	genericXCopy := make([]uint64, len(x))
	copy(genericXCopy, x)
	sumUint64(x, y)
	sumUint64Generic(genericXCopy, y)
	for i:=0; i<len(x);i++ {
		if x[i] != genericXCopy[i] {
			t.Fatalf("mismatch sums at index %d, expected %d : got %d", i, genericXCopy[i], x[i])
		}
	}
}

func TestSumUint64YLarger(t *testing.T) {
	x, y := generateSlices()
	y = append(y, uint64(100))
	genericXCopy := make([]uint64, len(x))
	copy(genericXCopy, x)
	sumUint64(x, y)
	sumUint64Generic(genericXCopy, y)
	for i:=0; i<len(x);i++ {
		if x[i] != genericXCopy[i] {
			t.Fatalf("mismatch sums at index %d, expected %d : got %d", i, genericXCopy[i], x[i])
		}
	}
}

func TestSumUint64XLarger(t *testing.T) {
	x, y := generateSlices()
	y = append(y, uint64(100))
	genericXCopy := make([]uint64, len(x))
	copy(genericXCopy, x)
	sumUint64(x, y)
	sumUint64Generic(genericXCopy, y)
	for i:=0; i<len(x);i++ {
		if x[i] != genericXCopy[i] {
			t.Fatalf("mismatch sums at index %d, expected %d : got %d", i, genericXCopy[i], x[i])
		}
	}
}

func generateSlices() ([]uint64, []uint64) {
	var x []uint64
	var y []uint64
	prng := rand.New(rand.NewSource(0))
	for i := 0; i < size; i++ {
		x = append(x, prng.Uint64())
		y = append(y, prng.Uint64())
	}
	return x, y
}

func BenchmarkSumUnit64(b *testing.B) {
	var x []uint64
	var y []uint64
	prng := rand.New(rand.NewSource(0))
	for j := 0; j < size; j++ {
		x = append(x, prng.Uint64())
		y = append(y, prng.Uint64())
	}
	for i := 0; i < b.N; i++ {
		SumUint64(x, y)
	}
}

func BenchmarkSumUnit64Generic(b *testing.B) {
	var x []uint64
	var y []uint64
	prng := rand.New(rand.NewSource(0))
	for j := 0; j < size; j++ {
		x = append(x, prng.Uint64())
		y = append(y, prng.Uint64())
	}
	for i := 0; i < b.N; i++ {
		sumUint64Generic(x, y)
	}
}

