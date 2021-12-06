//go:build purego || !amd64
// +build purego !amd64

package keyset

var purego = true

func Lookup(keyset []byte, key []byte) int {
	panic("not implemented")
}
