// +build !amd64

package sortedset

func dedupe1(b []byte) int {
	return dedupeGeneric(b, 1)
}

func dedupe2(b []byte) int {
	return dedupeGeneric(b, 2)
}

func dedupe4(b []byte) int {
	return dedupeGeneric(b, 4)
}

func dedupe8(b []byte) int {
	return dedupeGeneric(b, 8)
}

func dedupe16(b []byte) int {
	return dedupeGeneric(b, 16)
}

func dedupe32(b []byte) int {
	return dedupeGeneric(b, 32)
}
