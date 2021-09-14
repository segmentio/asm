// +build purego !amd64

package ascii

import "unsafe"

// ValidString returns true if s contains only ASCII characters.
func ValidString(s string) bool {
	p := *(*unsafe.Pointer)(unsafe.Pointer(&s))
	n := uintptr(len(s))

	for n >= 8 {
		if (*(*uint64)(p) & 0x8080808080808080) != 0 {
			return false
		}
		p = unsafe.Pointer(uintptr(p) + 8)
		n -= 8
	}

	if n > 4 {
		if (*(*uint32)(p) & 0x80808080) != 0 {
			return false
		}
		p = unsafe.Pointer(uintptr(p) + 4)
		n -= 4
	}

	var x uint32
	switch n {
	case 4:
		x = *(*uint32)(p)
	case 3:
		x = uint32(*(*uint16)(p)) | uint32(*(*uint8)(unsafe.Pointer(uintptr(p) + 2)))<<16
	case 2:
		x = uint32(*(*uint16)(p))
	case 1:
		x = uint32(*(*uint8)(p))
	default:
		return true
	}
	return (x & 0x80808080) == 0
}
