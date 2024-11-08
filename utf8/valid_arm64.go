//go:build !purego || arm64

package utf8

// Optimized version of Validate for inputs of more than 32B.
func validateNEON(p []byte) byte
