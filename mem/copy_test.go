package mem_test

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"testing"

	"github.com/segmentio/asm/mem"
)

func TestCopy(t *testing.T) {
	for _, N := range []int{0, 1, 2, 3, 4, 8, 10, 31, 32, 100, 1024, 4096} {
		t.Run(fmt.Sprintf("N=%d", N), func(t *testing.T) {
			src := make([]byte, N)
			dst := make([]byte, N)
			exp := make([]byte, N)

			prng := rand.New(rand.NewSource(0))
			io.ReadFull(prng, src)
			io.ReadFull(prng, dst)

			copy(exp, src)

			n := mem.Copy(dst, src)
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
	for _, N := range []int{7, 10, 31, 32, 100, 1024, 4096} {
		b.Run(fmt.Sprintf("N=%d", N), func(b *testing.B) {
			dst := make([]byte, N)
			src := make([]byte, N)
			io.ReadFull(rand.New(rand.NewSource(0)), src)
			b.SetBytes(int64(N))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				mem.Copy(dst, src)
			}
		})
	}
}
