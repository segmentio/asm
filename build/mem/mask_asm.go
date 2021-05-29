// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/segmentio/asm/build/internal/gen"
)

func main() {
	gen := Copy{
		CopyB:   ANDB,
		CopyW:   ANDW,
		CopyL:   ANDL,
		CopyQ:   ANDQ,
		CopyAVX: VPAND,
	}

	gen.Generate("Mask", "set bits of dst to zero and copies the one-bits of src to dst, returning the number of bytes written.")
}
