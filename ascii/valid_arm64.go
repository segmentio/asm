//go:build !purego && arm64
// +build !purego,arm64

package ascii

// ValidString returns true if s contains only ASCII characters.
func ValidString(s string) bool
