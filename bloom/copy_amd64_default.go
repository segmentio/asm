// +build !amd64

package bloom

func copyAVX2(dst, src *byte, n int) { panic("NOT SUPPORTED") }
