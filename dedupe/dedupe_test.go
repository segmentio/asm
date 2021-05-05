package dedupe

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestDedupe(t *testing.T) {
	for _, test := range []struct {
		name   string
		b      []byte
		size   int
		expect []byte
	}{
		{
			name: "empty",
			size: 1,
		},
		{
			name:   "size 1, all dupes",
			b:      []byte{1, 1, 1, 1, 1, 1, 1, 1},
			size:   1,
			expect: []byte{1},
		},
		{
			name:   "size 1, no dupes",
			b:      []byte{1, 2, 3, 4, 5, 6, 7, 8},
			size:   1,
			expect: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			name:   "size 1, some dupes",
			b:      []byte{0, 0, 0, 1, 1, 2, 3, 3, 4, 4, 4},
			size:   1,
			expect: []byte{0, 1, 2, 3, 4},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			actual := Dedupe(test.b, test.size)
			if !reflect.DeepEqual(actual, test.expect) {
				t.Fatalf("not equal: %v vs expected %v", actual, test.expect)
			}
		})
	}
}

var repeatChances = []float64{0, 0.1, 0.5, 1.0}

func TestDedupe16(t *testing.T) {
	testDedupeSize(t, 16)
}

func TestDedupe32(t *testing.T) {
	testDedupeSize(t, 32)
}

func testDedupeSize(t *testing.T, size int) {
	const maxCount = 100
	const iterations = 1000

	prng := rand.New(rand.NewSource(0))

	for i := 0; i < iterations; i++ {
		count := prng.Intn(maxCount)
		for _, p := range repeatChances {
			array, uniques := randomSortedArray(prng, size, count, p)
			result := Dedupe(array, size)
			if !reflect.DeepEqual(result, uniques) {
				t.Fatal("unexpected result")
			}
		}
	}
}

func BenchmarkDedupe16(b *testing.B) {
	benchmarkDedupeSize(b, 16)
}

func BenchmarkDedupe32(b *testing.B) {
	benchmarkDedupeSize(b, 32)
}

func benchmarkDedupeSize(b *testing.B, size int) {
	const bytes = 64 * 1024

	prng := rand.New(rand.NewSource(0))

	for _, p := range repeatChances {
		b.Run(fmt.Sprintf("%.2f", p), func(b *testing.B) {
			array, _ := randomSortedArray(prng, size, bytes/size, p)
			buf := make([]byte, len(array))

			b.SetBytes(bytes)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				copy(buf, array)
				Dedupe(buf, size)
			}
		})
	}
}

func randomSortedArray(prng *rand.Rand, size int, count int, repeatChance float64) (array []byte, uniques []byte) {
	if count == 0 {
		return nil, nil
	}

	pool := make([]byte, size*count)
	prng.Read(pool)
	sort.Sort(&chunks{b: pool, size: size})

	// Sanity checks:
	for i := size; i < len(pool); i += size {
		switch bytes.Compare(pool[i-size:i], pool[i:i+size]) {
		case 0:
			panic("duplicate item in pool")
		case 1:
			panic("not sorted correctly")
		}
	}

	array = make([]byte, 0, size*count)

	uniq := size
	for i := 0; i < count; i++ {
		array = append(array, pool[uniq-size:uniq]...)
		if prng.Float64() < repeatChance && i != count-1 {
			uniq += size
		}
	}

	uniques = pool[:uniq]
	return
}

type chunks struct {
	b    []byte
	size int
	tmp  []byte
}

func (s *chunks) Len() int {
	return len(s.b) / s.size
}

func (s *chunks) Less(i, j int) bool {
	return bytes.Compare(s.slice(i), s.slice(j)) < 0
}

func (s *chunks) Swap(i, j int) {
	tmp := make([]byte, s.size)
	copy(tmp, s.slice(j))
	copy(s.slice(j), s.slice(i))
	copy(s.slice(i), tmp)
}

func (s *chunks) slice(i int) []byte {
	return s.b[i*s.size : (i+1)*s.size]
}
