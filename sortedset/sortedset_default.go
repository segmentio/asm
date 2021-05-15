// +build !amd64

package sortedset

func dedupe16(b []byte) int {
	return dedupeGeneric(b, 16)
}

func dedupe32(b []byte) int {
	return dedupeGeneric(b, 32)
}

func intersect16(dst, a, b []byte) int {
	return intersectGeneric(dst, a, b, 16)
}
