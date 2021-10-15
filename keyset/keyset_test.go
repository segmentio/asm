package keyset

import (
	"fmt"
	"math/rand"
	"strconv"
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
		keyset := New(keys)
		if keyset == nil {
			t.Skip("Lookup is not implemented")
		}

		for j := range keys {
			if n := Lookup(keyset, keys[j]); n != j {
				t.Errorf("unexpected index for known key: %d, expected %d", n, j)
			}
		}
		if n := Lookup(keyset, []byte(fmt.Sprintf("key-%d", i+1))); n != len(keys) {
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

	copy(head, hex)
	copy(tail, hex)

	for i := 0; i <= 16; i++ {
		key := head[:i]
		keyset := New([][]byte{[]byte("foo"), []byte("bar"), key})
		if keyset == nil {
			t.Skip("Lookup is not implemented")
		}
		if n := Lookup(keyset, key); n != 2 {
			t.Errorf("unexpected lookup result %d", n)
		}
	}

	for i := 0; i <= 16; i++ {
		key := tail[i:]
		keyset := New([][]byte{[]byte("foo"), []byte("bar"), key})
		if n := Lookup(keyset, key); n != 2 {
			t.Errorf("unexpected lookup result for i=%d: %d", i, n)
		}
	}
}

func BenchmarkKeySet(b *testing.B) {
	keys := make([][]byte, 32)
	m := map[string]int{}
	for i := range keys {
		k := "key-" + strconv.Itoa(i)
		// k := strings.Repeat(strconv.Itoa(i), i)
		if len(k) > 16 {
			k = k[:16]
		}
		keys[i] = []byte(k)
		m[k] = i
	}

	prng := rand.New(rand.NewSource(0))

	const permutations = 1000 // enough to throw off the branch predictor hopeully
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

	keyset := New(keys)
	if keyset == nil {
		b.Skip("Lookup is not implemented")
	}

	b.Run("map-lookup-first", func(b *testing.B) {
		first := keys[0]
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = m[string(first)]
		}
	})
	b.Run("keyset-lookup-first", func(b *testing.B) {
		first := keys[0]
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Lookup(keyset, first)
		}
	})

	b.Run("map-lookup-last", func(b *testing.B) {
		last := keys[len(keys)-1]
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = m[string(last)]
		}
	})
	b.Run("keyset-lookup-last", func(b *testing.B) {
		last := keys[len(keys)-1]
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Lookup(keyset, last)
		}
	})

	b.Run("map-ordered-iteration", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, k := range keys {
				_ = m[string(k)]
			}
		}
	})
	b.Run("keyset-ordered-iteration", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, k := range keys {
				Lookup(keyset, k)
			}
		}
	})

	b.Run("map-random-iteration", func(b *testing.B) {
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
	b.Run("keyset-random-iteration", func(b *testing.B) {
		prng := rand.New(rand.NewSource(0))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p := prng.Intn(permutations)
			permutation := r[p*len(keys):][:len(keys)]
			for _, i := range permutation {
				Lookup(keyset, keys[i])
			}
		}
	})
}
