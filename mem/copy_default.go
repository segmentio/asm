// +build !amd64

package mem

func Copy(dst, src []byte) int { return copy(dst, src) }
