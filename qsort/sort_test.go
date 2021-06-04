package qsort

import (
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

var prng = rand.New(rand.NewSource(0))

func TestSort8(t *testing.T) {
	testSort(t, 8)
}

func TestSort16(t *testing.T) {
	testSort(t, 16)
}

func TestSort24(t *testing.T) {
	testSort(t, 24)
}

func TestSort32(t *testing.T) {
	testSort(t, 32)
}

func testSort(t *testing.T, size int) {
	const (
		iterations = 10000
		minCount   = 0
		maxCount   = 1000
	)

	buf := make([]byte, maxCount*size)

	for i := 0; i < iterations; i++ {
		count := randint(minCount, maxCount)
		slice := buf[:count*size]
		prng.Read(slice)

		// Test with/without duplicates.
		repeat := randint(0, count)
		for j := repeat; repeat > 0 && j < len(slice) && j + repeat < len(slice); j += repeat {
			copy(slice[j:j+repeat], slice[:repeat])
		}

		expect := make([]byte, len(slice))
		copy(expect, slice)
		sort.Sort(newGeneric(expect, size, nil))

		Sort(slice, size, nil)

		if !reflect.DeepEqual(expect, slice) {
			t.Fatal("buffer was not sorted correctly")
		}
	}
}

func randint(lo, hi int) int {
	if hi == lo {
		return lo
	}
	return prng.Intn(hi-lo) + lo
}

func BenchmarkSort8(b *testing.B) {
	benchSort(b, 8)
}

func BenchmarkSort16(b *testing.B) {
	benchSort(b, 16)
}

func BenchmarkSort24(b *testing.B) {
	benchSort(b, 24)
}

func BenchmarkSort32(b *testing.B) {
	benchSort(b, 32)
}

func benchSort(b *testing.B, size int) {
	for _, count := range []int{1e3, 1e4, 1e5, 1e6} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			buf := make([]byte, count*size)
			unsorted := make([]byte, count*size)
			prng.Read(unsorted)

			b.SetBytes(int64(len(buf)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				copy(buf, unsorted)
				Sort(buf, size, nil)
			}
		})
	}
}
