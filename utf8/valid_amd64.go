// Code generated by command: go run valid_asm.go -pkg utf8 -out ../utf8/valid_amd64.s -stubs ../utf8/valid_amd64.go. DO NOT EDIT.

//go:build !purego
// +build !purego

package utf8

// Validate is a more precise version of Valid that also indicates whether the input was valid ASCII.
func Validate(p []byte) (bool, bool)
