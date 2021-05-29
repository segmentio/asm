package bloom

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"testing"
)

func TestCopy(t *testing.T) {
	for _, N := range []int{0, 1, 2, 3, 4, 8, 10, 32, 100, 1024, 4096} {
		t.Run(fmt.Sprintf("N=%d", N), func(t *testing.T) {
			src := make([]byte, N)
			dst := make([]byte, N)
			exp := make([]byte, N)

			prng := rand.New(rand.NewSource(0))
			io.ReadFull(prng, src)
			io.ReadFull(prng, dst)

			copy(exp, dst)
			for i := range src {
				exp[i] |= src[i]
			}

			n := Copy(dst, src)
			if n != N {
				t.Errorf("copying did not apply to enough bytes: %d != %d", n, N)
			}

			if !bytes.Equal(dst, exp) {
				t.Error("copying produced the wrong output")
				t.Logf("expected: %08b", limit(exp, 8))
				t.Logf("found:    %08b", limit(dst, 8))
				t.Logf("source:   %08b", limit(src, 8))
			}
		})
	}
}

func BenchmarkCopy(b *testing.B) {
	const N = 4096
	dst := make([]byte, N)
	src := make([]byte, N)
	io.ReadFull(rand.New(rand.NewSource(0)), src)
	b.SetBytes(N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Copy(dst, src)
	}
}

func limit(b []byte, n int) []byte {
	if len(b) > n {
		b = b[:n]
	}
	return b
}
