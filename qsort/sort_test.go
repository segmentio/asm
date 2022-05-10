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

// Note, "8", "16", "32" etc are all byte measurements, not bits. So a 32 byte
// integer, for example, which you might see in e.g. a SHA256 hash.

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

func TestPivot8(t *testing.T) {
	lo := uint64(1)
	mid := uint64(2)
	hi := uint64(3)

	for i := 0; i < 1000; i++ {
		input := []uint64{lo, mid, hi}
		rand.Shuffle(3, func(i, j int) {
			input[i], input[j] = input[j], input[i]
		})
		medianOfThree64(input, 3, nil)
		if input[0] != mid {
			t.Fatal("medianOfThree128 did not put pivot in first position")
		}
	}
}

func TestPivot16(t *testing.T) {
	lo := uint128{lo: 1}
	mid := uint128{lo: 2}
	hi := uint128{lo: 3}

	for i := 0; i < 1000; i++ {
		input := []uint128{lo, mid, hi}
		rand.Shuffle(3, func(i, j int) {
			input[i], input[j] = input[j], input[i]
		})
		medianOfThree128(input, 3, nil)
		if input[0] != mid {
			t.Fatal("medianOfThree128 did not put pivot in first position")
		}
	}
}

func TestPivot24(t *testing.T) {
	lo := uint192{lo: 1}
	mid := uint192{lo: 2}
	hi := uint192{lo: 3}

	for i := 0; i < 1000; i++ {
		input := []uint192{lo, mid, hi}
		rand.Shuffle(3, func(i, j int) {
			input[i], input[j] = input[j], input[i]
		})
		medianOfThree192(input, 3, nil)
		if input[0] != mid {
			t.Fatal("medianOfThree192 did not put pivot in first position")
		}
	}
}

func TestPivot32(t *testing.T) {
	lo := uint256{d: 1}
	mid := uint256{d: 2}
	hi := uint256{d: 3}

	for i := 0; i < 1000; i++ {
		input := []uint256{lo, mid, hi}
		rand.Shuffle(3, func(i, j int) {
			input[i], input[j] = input[j], input[i]
		})
		medianOfThree256(input, 3, nil)
		if input[0] != mid {
			t.Fatal("medianOfThree256 did not put pivot in first position")
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
		b.Run("random-"+strconv.Itoa(count), benchSort(count, 8, 0, random, nil))
		if count > 1e4 {
			b.Run("partially-ordered(10)-"+strconv.Itoa(count), benchSort(count, 8, 10, random, nil))
			b.Run("partially-ordered(100)-"+strconv.Itoa(count), benchSort(count, 8, 100, random, nil))
			b.Run("partially-ordered(1000)-"+strconv.Itoa(count), benchSort(count, 8, 1000, random, nil))
		}
	}
}

func stdlibSort8(b *testing.B, size int) {
	// 8 bytes per int64
	b.SetBytes(8 * int64(size))
	data := make([]int64, size)
	unsorted := make([]int64, size)
	for j := 0; j < len(unsorted); j++ {
		unsorted[j] = int64(rand.Intn(size / 10))
	}
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		copy(data, unsorted)
		b.StartTimer()
		sort.Slice(data, func(i, j int) bool { return data[i] < data[j] })
		b.StopTimer()
	}
}

func stdlibSort8PartiallySorted(b *testing.B, size int, partitions int) {
	// 8 bytes per int64
	b.SetBytes(8 * int64(size))
	data := make([]int64, size)
	// panic if not a whole number
	partitionSize := int(size / partitions)
	partitionOrder := rand.Perm(partitions)
	groupedPartitions := make([][]int64, partitions)

	for i := 0; i < len(groupedPartitions); i++ {
		partition := make([]int64, partitionSize)
		for j := 0; j < len(partition); j++ {
			partition[j] = int64(rand.Intn(size / 10))
		}
		sort.Slice(partition, func(i, j int) bool { return partition[i] < partition[j] })
		groupedPartitions[partitionOrder[i]] = partition
	}

	partiallyOrdered := make([]int64, size)
	for _, partition := range groupedPartitions {
		partiallyOrdered = append(partiallyOrdered, partition...)
	}

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		copy(data, partiallyOrdered)
		b.StartTimer()
		sort.Slice(data, func(i, j int) bool { return data[i] < data[j] })
		b.StopTimer()
	}
}

func BenchmarkStdlibSort8(b *testing.B) {
	for _, size := range []int{1e5, 1e6} {
		b.Run("random-"+strconv.Itoa(size), func(b *testing.B) {
			stdlibSort8(b, size)
		})
		b.Run("partially-sorted(10)-"+strconv.Itoa(size), func(b *testing.B) {
			stdlibSort8PartiallySorted(b, size, 10)
		})
		b.Run("partially-sorted(100)-"+strconv.Itoa(size), func(b *testing.B) {
			stdlibSort8PartiallySorted(b, size, 100)
		})
		b.Run("partially-sorted(1000)-"+strconv.Itoa(size), func(b *testing.B) {
			stdlibSort8PartiallySorted(b, size, 1000)
		})
	}
}

func BenchmarkSort8Indirect(b *testing.B) {
	swap := func(int, int) {}
	const count = 100000
	b.Run("random", benchSort(count, 8, 0, random, swap))
	b.Run("asc", benchSort(count, 8, 0, asc, swap))
	b.Run("desc", benchSort(count, 8, 0, desc, swap))
}

func BenchmarkSort16(b *testing.B) {
	for _, count := range []int{1e3, 1e4, 1e5, 1e6} {
		b.Run(strconv.Itoa(count), benchSort(count, 16, 0, random, nil))
	}
}

func BenchmarkSort16Indirect(b *testing.B) {
	swap := func(int, int) {}
	const count = 100000
	b.Run("random", benchSort(count, 16, 0, random, swap))
	b.Run("asc", benchSort(count, 16, 0, asc, swap))
	b.Run("desc", benchSort(count, 16, 0, desc, swap))
}

func BenchmarkSort24(b *testing.B) {
	for _, count := range []int{1e3, 1e4, 1e5, 1e6} {
		b.Run(strconv.Itoa(count), benchSort(count, 24, 0, random, nil))
	}
}

func BenchmarkSort24Indirect(b *testing.B) {
	swap := func(int, int) {}
	const count = 100000
	b.Run("random", benchSort(count, 24, 0, random, swap))
	b.Run("asc", benchSort(count, 24, 0, asc, swap))
	b.Run("desc", benchSort(count, 24, 0, desc, swap))
}

func BenchmarkSort32(b *testing.B) {
	for _, count := range []int{1e3, 1e4, 1e5, 1e6} {
		b.Run(strconv.Itoa(count), benchSort(count, 0, 32, random, nil))
	}
}

func BenchmarkSort32Indirect(b *testing.B) {
	swap := func(int, int) {}
	const count = 100000
	b.Run("random", benchSort(count, 32, 0, random, swap))
	b.Run("asc", benchSort(count, 32, 0, asc, swap))
	b.Run("desc", benchSort(count, 32, 0, desc, swap))
}

type order int

const (
	random order = iota
	asc
	desc
	partiallyOrdered
)

func benchSort(count, size, partitions int, order order, indirect func(int, int)) func(*testing.B) {
	return func(b *testing.B) {
		b.StopTimer()
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

		if order == partiallyOrdered {
			// panic if not a whole number
			partitionSize := int((count * size) / partitions)
			partitionOrder := rand.Perm(partitions)
			groupedPartitions := make([][]byte, partitions)

			for i := 0; i < len(groupedPartitions); i++ {
				partition := make([]byte, partitionSize)
				sort.Sort(newGeneric(partition, partitionSize, nil))
				groupedPartitions[partitionOrder[i]] = partition
			}

			for _, partition := range groupedPartitions {
				unsorted = append(unsorted, partition...)
			}
		}

		b.SetBytes(int64(len(buf)))

		for i := 0; i < b.N; i++ {
			copy(buf, unsorted)
			b.StartTimer()
			Sort(buf, size, indirect)
			b.StopTimer()
		}
	}
}
