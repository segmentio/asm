// +build !amd64

package qsort

func distributeForward64(data []uint64, scratch []uint64, limit int, lo int, hi int) int {
	panic("not implemented")
}

func distributeBackward64(data []uint64, scratch []uint64, limit int, lo int, hi int) int {
	panic("not implemented")
}

func insertionsort128NoSwap(data []struct {
	hi uint64
	lo uint64
}, base int, swap func(int, int)) {
	panic("not implemented")
}

func distributeForward128(data []struct {
	hi uint64
	lo uint64
}, scratch []struct {
	hi uint64
	lo uint64
}, limit int, lo int, hi int) int {
	panic("not implemented")
}

func distributeBackward128(data []struct {
	hi uint64
	lo uint64
}, scratch []struct {
	hi uint64
	lo uint64
}, limit int, lo int, hi int) int {
	panic("not implemented")
}

func insertionsort256NoSwap(data []struct {
	a uint64
	b uint64
	c uint64
	d uint64
}, base int, swap func(int, int)) {
	panic("not implemented")
}

func distributeForward256(data []struct {
	a uint64
	b uint64
	c uint64
	d uint64
}, scratch []struct {
	a uint64
	b uint64
	c uint64
	d uint64
}, limit int, lo int, hi int) int {
	panic("not implemented")
}

func distributeBackward256(data []struct {
	a uint64
	b uint64
	c uint64
	d uint64
}, scratch []struct {
	a uint64
	b uint64
	c uint64
	d uint64
}, limit int, lo int, hi int) int {
	panic("not implemented")
}
