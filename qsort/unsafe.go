package qsort

import "unsafe"

func unsafeBytesToU64(b []byte) []uint64 {
	return *(*[]uint64)(unsafe.Pointer(castBytes(b, 8)))
}

func unsafeBytesToU128(b []byte) []uint128 {
	return *(*[]uint128)(unsafe.Pointer(castBytes(b, 16)))
}

func unsafeBytesToU192(b []byte) []uint192 {
	return *(*[]uint192)(unsafe.Pointer(castBytes(b, 24)))
}

func unsafeBytesToU256(b []byte) []uint256 {
	return *(*[]uint256)(unsafe.Pointer(castBytes(b, 32)))
}

func castBytes(b []byte, size int) *sliceHeader {
	return &sliceHeader{
		Data: *(*unsafe.Pointer)(unsafe.Pointer(&b)),
		Len:  len(b) / size,
		Cap:  len(b) / size,
	}
}

func unsafeU64ToBytes(u []uint64) []byte {
	return *(*[]byte)(unsafe.Pointer(&sliceHeader{
		Data: *(*unsafe.Pointer)(unsafe.Pointer(&u)),
		Len:  len(u) * 8,
		Cap:  len(u) * 8,
	}))
}

func unsafeU128ToBytes(u []uint128) []byte {
	return *(*[]byte)(unsafe.Pointer(&sliceHeader{
		Data: *(*unsafe.Pointer)(unsafe.Pointer(&u)),
		Len:  len(u) * 16,
		Cap:  len(u) * 16,
	}))
}

func unsafeU192ToBytes(u []uint192) []byte {
	return *(*[]byte)(unsafe.Pointer(&sliceHeader{
		Data: *(*unsafe.Pointer)(unsafe.Pointer(&u)),
		Len:  len(u) * 24,
		Cap:  len(u) * 24,
	}))
}

func unsafeU256ToBytes(u []uint256) []byte {
	return *(*[]byte)(unsafe.Pointer(&sliceHeader{
		Data: *(*unsafe.Pointer)(unsafe.Pointer(&u)),
		Len:  len(u) * 32,
		Cap:  len(u) * 32,
	}))
}

type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}
