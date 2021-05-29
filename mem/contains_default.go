// +build !amd64

package mem

func containsByteAVX2(haystack []byte, needle byte) bool {
	panic("not implemented on !amd64")
}
