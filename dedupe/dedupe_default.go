// +build !amd64

package dedupe

func dedupe16(b []byte) (pos int) {
	return dedupeGeneric(b, 16)
}
