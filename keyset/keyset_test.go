package keyset

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/segmentio/asm/internal/buffer"
)

func TestKeySet(t *testing.T) {
	const max = 23

	keys := make([][]byte, max)
	for i := 0; i < max; i++ {
		keys = keys[:i]
		for j := range keys {
			keys[j] = []byte(strconv.Itoa(i - j))
		}
		lookup := New(keys)

		for j := range keys {
			if n := lookup(keys[j]); n != j {
				t.Errorf("unexpected index for known key: %d, expected %d", n, j)
			}
		}
		if n := lookup([]byte(fmt.Sprintf("key-%d", i+1))); n != len(keys) {
			t.Errorf("unexpected index for unknown key: %d", n)
		}
	}
}

const hex = "0123456789abcdef"

func TestPageBoundary(t *testing.T) {
	buf, err := buffer.New(16)
	if err != nil {
		t.Fatal(err)
	}
	defer buf.Release()

	head := buf.ProtectHead()
	tail := buf.ProtectTail()

	var chars [16]byte
	for i := 0; i < 16; i++ {
		chars[i] = hex[i]
	}
	copy(head, chars[:])
	copy(tail, chars[:])

	for i := 0; i < 16; i++ {
		key := head[:i]
		lookup := New([][]byte{[]byte("foo"), []byte("bar"), key})
		if n := lookup(key); n != 2 {
			t.Errorf("unexpected lookup result %d", n)
		}
	}

	for i := 0; i < 16; i++ {
		key := tail[i:]
		lookup := New([][]byte{[]byte("foo"), []byte("bar"), key})
		if n := lookup(key); n != 2 {
			t.Errorf("unexpected lookup result for i=%d: %d", i, n)
		}
	}
}

func BenchmarkIteration(b *testing.B) {
	keys := make([][]byte, 8)
	m := map[string]int{}
	for i := range keys {
		k := strings.Repeat("x", i)
		keys[i] = []byte(k)
		m[k] = i
	}

	prng := rand.New(rand.NewSource(0))

	const permutations = 1000
	r := make([]int, len(keys)*permutations)
	for i := 0; i < permutations; i++ {
		x := r[i*len(keys):][:len(keys)]
		for j := range x {
			x[j] = j
		}
		prng.Shuffle(len(keys), func(a, b int) {
			x[a], x[b] = x[b], x[a]
		})
	}

	lookup := New(keys)

	b.Run("map-ordered", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, k := range keys {
				_ = m[string(k)]
			}
		}
	})

	b.Run("keyset-ordered", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, k := range keys {
				lookup(k)
			}
		}
	})

	b.Run("map-random", func(b *testing.B) {
		prng := rand.New(rand.NewSource(0))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p := prng.Intn(permutations)
			permutation := r[p*len(keys):][:len(keys)]
			for _, i := range permutation {
				_ = m[string(keys[i])]
			}
		}
	})

	b.Run("keyset-random", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p := rand.Intn(permutations)
			permutation := r[p*len(keys):][:len(keys)]
			for _, i := range permutation {
				lookup(keys[i])
			}
		}
	})
}
