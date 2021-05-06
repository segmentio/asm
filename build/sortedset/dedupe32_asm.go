// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("dedupe32", NOSPLIT, "func(b []byte) (pos int)")

	src := Load(Param("b").Base(), GP64())
	length := Load(Param("b").Len(), GP64())

	// Calculate end of the slice.
	end := GP64()
	MOVQ(src, end)
	ADDQ(length, end)

	// Load the first item (which is never a duplicate).
	prev := YMM()
	VMOVUPS(Mem{Base: src}, prev)

	// Advance to the second item.
	ADDQ(Imm(32), src)

	// Keep a separate write pointer.
	dst := GP64()
	MOVQ(src, dst)

	// Loop until we're done.
	Label("loop")
	CMPQ(src, end)
	JE(LabelRef("done"))

	// Load the item at the input pointer.
	item := YMM()
	VMOVUPS(Mem{Base: src}, item)
	ADDQ(Imm(32), src)

	// Compare item == prev by comparing each byte.
	result := YMM()
	VPCMPEQB(prev, item, result)

	// Extract the equality mask, where 0xFFFFFFFF indicates that all bytes are equal.
	mask := GP32()
	VPMOVMSKB(result, mask)
	CMPL(mask, U32(0xFFFFFFFF))
	JE(LabelRef("loop")) // skip the write if they're equal

	// Write item to dst and advance the write pointer.
	VMOVUPS(item, Mem{Base: dst})
	ADDQ(Imm(32), dst)
	VMOVUPS(item, prev)
	JMP(LabelRef("loop"))

	Label("done")

	// Calculate and return byte offset of the dst pointer.
	base := Load(Param("b").Base(), GP64())
	SUBQ(base, dst)
	Store(dst, Return("pos"))

	RET()
	Generate()
}
