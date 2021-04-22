package bswap

// Swap64 performs an in-place byte swap on each qword of the input buffer.
func Swap64(b []byte) {
	if len(b)%8 != 0 {
		panic("swap64 expects full qwords")
	}
	swap64(b)
}
