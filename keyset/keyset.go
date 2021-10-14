package keyset

import (
	"bytes"

	"github.com/segmentio/asm/cpu"
)

// New returns a Lookup function that returns the index of a particular key in
// the array of input keys.
//
// An optimized routine is used if the processor supports AVX instructions and
// the maximum length of any of the keys is less than or equal to 16. Otherwise,
// a pure go routine that uses runtime.memequal() is used.
func New(keys [][]byte) Lookup {
	if len(keys) == 0 {
		return emptySetLookup
	}

	maxWidth := 0
	for _, k := range keys {
		if len(k) > maxWidth {
			maxWidth = len(k)
		}
	}

	if maxWidth <= 16 && cpu.X86.Has(cpu.AVX) {
		var b bytes.Buffer
		n := len(keys) * 16
		b.Grow(n)
		buffer := b.Bytes()[:n]
		lengths := make([]uint32, len(keys))

		for i, k := range keys {
			lengths[i] = uint32(len(k))
			copy(buffer[i*16:], k)
		}

		return func(k []byte) int {
			return searchAVX(&buffer[0], lengths, k)
		}
	}

	return func(k []byte) (i int) {
		for ; i < len(keys); i++ {
			if string(k) == string(keys[i]) {
				break
			}
		}
		return
	}
}

// Lookup is a function that searches for a key in the array of keys provided to
// New(). If the key is found, the function returns the first index of that key.
// If the key can't be found, the length of the input array is returned.
type Lookup func(key []byte) int

func emptySetLookup([]byte) int {
	return 0
}
