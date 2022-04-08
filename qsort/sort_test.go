package qsort

import (
	"bytes"
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
		iterations = 1000
		minCount   = 0
		maxCount   = 1000
	)

	buf := make([]byte, maxCount*size)
	// A first test to validate that the swap function is called properly:
	prng.Read(buf)

	values := make([]byte, len(buf))
	copy(values, buf)

	tmp := make([]byte, size)
	Sort(buf, size, func(i, j int) {
		vi := values[i*size : (i+1)*size]
		vj := values[j*size : (j+1)*size]
		copy(tmp, vi)
		copy(vi, vj)
		copy(vj, tmp)
	})

	if !bytes.Equal(buf, values) {
		t.Fatal("values were not sorted correctly by the swap function")
	}

	for i := 0; i < iterations; i++ {
		count := randint(minCount, maxCount)
		slice := buf[:count*size]
		prng.Read(slice)

		// Test with/without duplicates.
		repeat := randint(0, count)
		for j := repeat; repeat > 0 && j < len(slice) && j+repeat < len(slice); j += repeat {
			copy(slice[j:j+repeat], slice[:repeat])
		}

		expect := values[:len(slice)]
		copy(expect, slice)
		sort.Sort(newGeneric(expect, size, nil))

		if !sort.IsSorted(newGeneric(expect, size, nil)) {
			t.Fatal("reference implementation did not produce a sorted output")
		}

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
	for _, count := range []int{1e3, 1e4, 1e5, 1e6} {
		b.Run(strconv.Itoa(count), benchSort(count, 8, random, nil))
	}
}

func BenchmarkSort8Indirect(b *testing.B) {
	swap := func(int, int) {}
	const count = 100000
	b.Run("random", benchSort(count, 8, random, swap))
	b.Run("asc", benchSort(count, 8, asc, swap))
	b.Run("desc", benchSort(count, 8, desc, swap))
}

func BenchmarkSort16(b *testing.B) {
	for _, count := range []int{1e3, 1e4, 1e5, 1e6} {
		b.Run(strconv.Itoa(count), benchSort(count, 16, random, nil))
	}
}

func BenchmarkSort16Indirect(b *testing.B) {
	swap := func(int, int) {}
	const count = 100000
	b.Run("random", benchSort(count, 16, random, swap))
	b.Run("asc", benchSort(count, 16, asc, swap))
	b.Run("desc", benchSort(count, 16, desc, swap))
}

func BenchmarkSort24(b *testing.B) {
	for _, count := range []int{1e3, 1e4, 1e5, 1e6} {
		b.Run(strconv.Itoa(count), benchSort(count, 24, random, nil))
	}
}

func BenchmarkSort24Indirect(b *testing.B) {
	swap := func(int, int) {}
	const count = 100000
	b.Run("random", benchSort(count, 24, random, swap))
	b.Run("asc", benchSort(count, 24, asc, swap))
	b.Run("desc", benchSort(count, 24, desc, swap))
}

func BenchmarkSort32(b *testing.B) {
	for _, count := range []int{1e3, 1e4, 1e5, 1e6} {
		b.Run(strconv.Itoa(count), benchSort(count, 32, random, nil))
	}
}

func BenchmarkSort32Indirect(b *testing.B) {
	swap := func(int, int) {}
	const count = 100000
	b.Run("random", benchSort(count, 32, random, swap))
	b.Run("asc", benchSort(count, 32, asc, swap))
	b.Run("desc", benchSort(count, 32, desc, swap))
}

type order int

const (
	random order = iota
	asc
	desc
)

func benchSort(count, size int, order order, indirect func(int, int)) func(*testing.B) {
	return func(b *testing.B) {
		buf := make([]byte, count*size)
		unsorted := make([]byte, count*size)
		prng.Read(unsorted)

		if order == asc || order == desc {
			sort.Sort(newGeneric(unsorted, size, nil))
		}
		if order == desc {
			g := newGeneric(unsorted, size, nil)
			items := g.Len()
			for i := 0; i < items/2; i++ {
				g.Swap(i, items-1-i)
			}
		}

		b.SetBytes(int64(len(buf)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			copy(buf, unsorted)
			Sort(buf, size, indirect)
		}
	}
}
