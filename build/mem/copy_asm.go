// +build ignore

package main

import "github.com/segmentio/asm/build/internal/x86"

func main() {
	x86.GenerateCopy("Copy", "copies src to dst, returning the number of bytes written.", nil)
}
