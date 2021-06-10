// +build !amd64

package qsort

func insertionsort128NoSwapAsm(data []byte) {
	panic("not implemented")
}

func distributeForward128(data, scratch *byte, limit, lo, hi, pivot int) int {
	panic("not implemented")
}

func distributeBackward128(data, scratch *byte, limit, lo, hi, pivot int) int {
	panic("not implemented")
}

func insertionsort256NoSwapAsm(data []byte) {
	panic("not implemented")
}

func distributeForward256(data, scratch *byte, limit, lo, hi, pivot int) int {
	panic("not implemented")
}

func distributeBackward256(data, scratch *byte, limit, lo, hi, pivot int) int {
	panic("not implemented")
}
