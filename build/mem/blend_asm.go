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
	x86.GenerateCopy("Blend", "copies the one-bits of src to dst, returning the number of bytes written.",
		x86.BinaryOpTable(ORB, ORW, ORL, ORQ, POR, VPOR))
}
