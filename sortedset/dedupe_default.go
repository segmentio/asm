// +build !amd64

package sortedset

func dedupe16(b []byte) (pos int) {
	return dedupeGeneric(b, 16)
}

func dedupe32(b []byte) (pos int) {
	return dedupeGeneric(b, 32)
}
