//go:build purego || !amd64
// +build purego !amd64

package keyset

func searchAVX(buffer *byte, lengths []uint32, key []byte) int {
	panic("not implemented")
}
