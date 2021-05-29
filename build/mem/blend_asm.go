// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/segmentio/asm/build/internal/x86"
)

func main() {
	gen := Copy{
		CopyB:   ORB,
		CopyW:   ORW,
		CopyL:   ORL,
		CopyQ:   ORQ,
		CopyAVX: VPOR,
	}

	gen.Generate("Blend", "copies the one-bits of src to dst, returning the number of bytes written.")
}
