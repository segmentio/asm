// +build !amd64

package bloom

// Copy performs a OR of src and dst into dst, returning the number of bytes
// written to dst.
func Copy(dst, src []byte) int {
	switch {
	case len(dst) < len(src):
		src = src[:len(dst)]
	case len(dst) > len(src):
		dst = dst[:len(src)]
	}

	for i := range dst {
		dst[i] |= src[i]
	}

	return len(dst)
}
