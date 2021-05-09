package unsafebytes

import "unsafe"

func Pointer(b []byte) *byte {
	return *(**byte)(unsafe.Pointer(&b))
}

func String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
