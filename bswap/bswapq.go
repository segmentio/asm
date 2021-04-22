package bswap

// BSwapQ performs an in-place byte swap on each qword of the input buffer.
func BSwapQ(b []byte) {
	if len(b) % 8 != 0 {
		panic("bswapq expects full qwords")
	}
	bswapq(b)
}
