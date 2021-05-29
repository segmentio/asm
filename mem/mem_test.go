package mem_test

func limit(b []byte, n int) []byte {
	if len(b) > n {
		b = b[:n]
	}
	return b
}
