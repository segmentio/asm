// +build !amd64

package bloom

func copyAVX2(dst, src *bytes, n int) { panic("NOT SUPPORTED") }
