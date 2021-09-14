// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"

	"github.com/segmentio/asm/build/internal/x86"
)

func init() {
	ConstraintExpr("!purego")
}

func main() {
	x86.GenerateCopy("Mask", "set bits of dst to zero and copies the one-bits of src to dst, returning the number of bytes written.",
		x86.BinaryOpTable(ANDB, ANDW, ANDL, ANDQ, PAND, VPAND))
}
