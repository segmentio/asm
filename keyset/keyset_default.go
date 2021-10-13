//go:build purego || !amd64
// +build purego !amd64

package keyset

func search16(buffer *byte, lengths []uint32, key []byte) int {
	panic("not implemented")
}
