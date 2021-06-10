package qsort

import "unsafe"

func unsafeBytesToU64(b []byte) []uint64 {
	return *(*[]uint64)(castBytes(b, 8))
}

func unsafeBytesToU128(b []byte) []uint128 {
	return *(*[]uint128)(castBytes(b, 16))
}

func unsafeBytesToU192(b []byte) []uint192 {
	return *(*[]uint192)(castBytes(b, 24))
}

func unsafeBytesToU256(b []byte) []uint256 {
	return *(*[]uint256)(castBytes(b, 32))
}

func unsafeU128ToBytes(u []uint128) []byte {
	return castSlice(unsafe.Pointer(&u), len(u)*16)
}

func unsafeU256ToBytes(u []uint256) []byte {
	return castSlice(unsafe.Pointer(&u), len(u)*32)
}

func castBytes(b []byte, size int) unsafe.Pointer {
	return unsafe.Pointer(&sliceHeader{
		Data: *(*unsafe.Pointer)(unsafe.Pointer(&b)),
		Len:  len(b) / size,
		Cap:  len(b) / size,
	})
}

func castSlice(ptr unsafe.Pointer, length int) []byte {
	return *(*[]byte)(unsafe.Pointer(&sliceHeader{
		Data: *(*unsafe.Pointer)(ptr),
		Len:  length,
		Cap:  length,
	}))
}

func unsafeU64Addr(slice []uint64) *byte {
	return (*byte)(unsafe.Pointer(&slice[0]))
}

func unsafeU128Addr(slice []uint128) *byte {
	return (*byte)(unsafe.Pointer(&slice[0]))
}

func unsafeU256Addr(slice []uint256) *byte {
	return (*byte)(unsafe.Pointer(&slice[0]))
}

type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}
