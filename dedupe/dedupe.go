package dedupe

import (
	"bytes"
	"unsafe"
)

func Dedupe(b []byte, size int) []byte {
	if size <= 0 || len(b)%size != 0 {
		panic("len(b) % size != 0")
	}
	if len(b) <= size {
		return b
	}

	var pos int
	switch size {
	case 16:
		pos = dedupeUnsafe16(b)
	default:
		pos = dedupeGeneric(b, size)
	}

	return b[:pos]
}

func dedupeGeneric(b []byte, size int) int {
	pos := size
	prev := b[:size]
	for i := size; i < len(b); i += size {
		item := b[i : i+size]
		if !bytes.Equal(prev, item) {
			copy(b[pos:], item)
			pos += size
			prev = item
		}
	}
	return pos
}

func dedupeUnsafe16(values []byte) int {
	n := (len(values) / 16) * 16
	ptr := *(*unsafe.Pointer)(unsafe.Pointer(&values))
	end := unsafe.Pointer(uintptr(ptr) + uintptr(n))
	rdp := unsafe.Pointer(uintptr(ptr) + uintptr(16))
	wrp := ptr

	for uintptr(rdp) < uintptr(end) {
		u := (*[16]byte)(wrp)
		v := (*[16]byte)(rdp)

		if *u != *v {
			if wrp = unsafe.Pointer(uintptr(wrp) + 16); wrp != rdp {
				*(*[16]byte)(wrp) = *v
			}
		}

		rdp = unsafe.Pointer(uintptr(rdp) + 16)
	}

	return int(uintptr(wrp)-uintptr(ptr)) + 16
}
