package qsort

import (
	"unsafe"
)

type b8 [8]byte
type b16 [16]byte
type b24 [24]byte
type b32 [32]byte

func unsafeBytesTo8(b []byte) []b8 {
	return *(*[]b8)(unsafe.Pointer(cast(b, 8)))
}

func unsafeBytesTo16(b []byte) []b16 {
	return *(*[]b16)(unsafe.Pointer(cast(b, 16)))
}

func unsafeBytesTo24(b []byte) []b24 {
	return *(*[]b24)(unsafe.Pointer(cast(b, 24)))
}

func unsafeBytesTo32(b []byte) []b32 {
	return *(*[]b32)(unsafe.Pointer(cast(b, 32)))
}

func cast(b []byte, size int) *sliceHeader {
	return &sliceHeader{
		Data: *(*unsafe.Pointer)(unsafe.Pointer(&b)),
		Len:  len(b) / size,
		Cap:  len(b) / size,
	}
}

type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}
