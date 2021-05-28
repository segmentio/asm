// +build !amd64

package ascii

import "unsafe"

// ValidString returns true if s contains only printable ASCII characters.
func ValidPrintString(s string) bool {
	p := *(*unsafe.Pointer)(unsafe.Pointer(&s))
	n := uintptr(len(s))

	for n >= 8 {
		if hasLess64(*(*uint64)(p), 0x20) || hasMore64(*(*uint64)(p), 0x7e) {
			return false
		}
		p = unsafe.Pointer(uintptr(p) + 8)
		n -= 8
	}

	if n >= 4 {
		if hasLess32(*(*uint32)(p), 0x20) || hasMore32(*(*uint32)(p), 0x7e) {
			return false
		}
		p = unsafe.Pointer(uintptr(p) + 4)
		n -= 4
	}

	var x uint32
	switch n {
	case 3:
		x = 0x20000000 | uint32(*(*uint16)(p)) | uint32(*(*uint8)(unsafe.Pointer(uintptr(p) + 2)))<<16
	case 2:
		x = 0x20200000 | uint32(*(*uint16)(p))
	case 1:
		x = 0x20202000 | uint32(*(*uint8)(p))
	default:
		return true
	}
	return !(hasLess32(x, 0x20) || hasMore32(x, 0x7e))
}
