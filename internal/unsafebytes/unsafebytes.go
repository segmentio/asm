package unsafebytes

import "unsafe"

func Pointer(b []byte) *byte {
	return *(**byte)(unsafe.Pointer(&b))
}
