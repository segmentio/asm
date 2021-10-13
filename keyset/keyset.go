package keyset

// New returns a Lookup function that returns the index
// of a particular key in the array of input keys.
func New(keys [][]byte) Lookup {
	if len(keys) == 0 {
		return emptySetLookup
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
