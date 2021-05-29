// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/segmentio/asm/build/internal/x86"
)

func main() {
	gen := Copy{
		CopyB: MOVB,
		CopyW: MOVW,
		CopyL: MOVL,
		CopyQ: MOVQ,
	}

	gen.Generate("Copy", "copies src to dst, returning the number of bytes written.")
}
