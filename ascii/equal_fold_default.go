// +build !amd64

package ascii

import "unsafe"

// EqualFoldString is a version of strings.EqualFold designed to work on ASCII
// input instead of UTF-8.
//
// When the program has guarantees that the input is composed of ASCII
// characters only, it allows for greater optimizations.
func EqualFoldString(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	n := uintptr(len(a))
	p := *(*unsafe.Pointer)(unsafe.Pointer(&a))
	q := *(*unsafe.Pointer)(unsafe.Pointer(&b))

	for n >= 8 {
		const mask = 0xDFDFDFDFDFDFDFDF

		if (*(*uint64)(p) & mask) != (*(*uint64)(q) & mask) {
			return false
		}

		p = unsafe.Pointer(uintptr(p) + 8)
		q = unsafe.Pointer(uintptr(q) + 8)
		n -= 8
	}

	if n > 4 {
		const mask = 0xDFDFDFDF

		if (*(*uint32)(p) & mask) != (*(*uint32)(q) & mask) {
			return false
		}

		p = unsafe.Pointer(uintptr(p) + 4)
		q = unsafe.Pointer(uintptr(q) + 4)
		n -= 4
	}

	switch n {
	case 4:
		return (*(*uint32)(p) & 0xDFDFDFDF) == (*(*uint32)(q) & 0xDFDFDFDF)
	case 3:
		x := uint32(*(*uint16)(p)) | uint32(*(*uint8)(unsafe.Pointer(uintptr(p) + 2)))<<16
		y := uint32(*(*uint16)(q)) | uint32(*(*uint8)(unsafe.Pointer(uintptr(q) + 2)))<<16
		return (x & 0xDFDFDF) == (y & 0xDFDFDF)
	case 2:
		return (*(*uint16)(p) & 0xDFDF) == (*(*uint16)(q) & 0xDFDF)
	case 1:
		return (*(*uint8)(p) & 0xDF) == (*(*uint8)(q) & 0xDF)
	default:
		return true
	}
}
