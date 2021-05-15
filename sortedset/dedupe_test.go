package sortedset

import (
	"fmt"
	"math/rand"
	"testing"
)

var dedupeSpecializationSizes = []int{16, 32}

var repeatChances = []float64{0, 0.1, 0.5, 1.0}

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
			assertArraysEqual(t, test.expect, actual, test.size)
		})
	}

	// Test the specializations.
	for _, size := range dedupeSpecializationSizes {
		t.Run(fmt.Sprintf("size %d, random", size), func(t *testing.T) {
			const maxCount = 100
			const iterations = 1000

			prng := rand.New(rand.NewSource(0))

			for i := 0; i < iterations; i++ {
				count := prng.Intn(maxCount)
				for _, p := range repeatChances {
					array, uniques := randomSortedArray(prng, size, count, p)
					result := Dedupe(array, size)
					assertArraysEqual(t, uniques, result, size)
				}
			}
		})
	}
}

func BenchmarkDedupe(b *testing.B) {
	for _, size := range dedupeSpecializationSizes {
		for _, p := range repeatChances {
			b.Run(fmt.Sprintf("size %d, with %d%% chance of repeat", size, int(p*100)), func(b *testing.B) {
				const bytes = 64 * 1024

				prng := rand.New(rand.NewSource(0))

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
}
