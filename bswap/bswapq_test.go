package bswap

import (
	"encoding/binary"
	"io"
	"math/rand"
	"testing"
)

func TestBSwapQ(t *testing.T) {
	input := make([]byte, 64*1024)
	prng := rand.New(rand.NewSource(0))
	io.ReadFull(prng, input)

	output := make([]byte, 1024)
	copy(output, input)
	BSwapQ(output)

	for i := 0; i < len(output); i += 8 {
		u1 := binary.BigEndian.Uint64(input[i:])
		u2 := binary.LittleEndian.Uint64(output[i:])

		if u1 != u2 {
			t.Fatalf("bytes weren't swapped at offset %d: %v / %v", i, u1, u2)
		}
	}
}

func BenchmarkBSwapQ(b *testing.B) {
	input := make([]byte, 64*1024)
	prng := rand.New(rand.NewSource(0))
	io.ReadFull(prng, input)

	b.SetBytes(int64(len(input)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		BSwapQ(input)
	}
}