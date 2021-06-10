// +build !amd64

package qsort

func distributeForward64(data, scratch []uint64, limit, lo, hi int) int {
	panic("not implemented")
}

func distributeBackward64(data, scratch []uint64, limit, lo, hi int) int {
	panic("not implemented")
}

func insertionsort128NoSwapAsm(data []uint128, base int, swap func(int, int)) {
	panic("not implemented")
}

func distributeForward128(data, scratch []uint128, limit, lo, hi int) int {
	panic("not implemented")
}

func distributeBackward128(data, scratch []uint128, limit, lo, hi int) int {
	panic("not implemented")
}

func insertionsort256NoSwapAsm(data []uint256, base int, swap func(int, int)) {
	panic("not implemented")
}

func distributeForward256(data, scratch []uint256, limit, lo, hi int) int {
	panic("not implemented")
}

func distributeBackward256(data, scratch []uint256, limit, lo, hi int) int {
	panic("not implemented")
}
