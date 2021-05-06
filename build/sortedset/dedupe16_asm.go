// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("dedupe16", NOSPLIT, "func(b []byte) (pos int)")

	src := Load(Param("b").Base(), GP64())
	length := Load(Param("b").Len(), GP64())

	// Calculate end of the slice.
	end := GP64()
	MOVQ(src, end)
	ADDQ(length, end)

	// Load the first item (which is never a duplicate).
	prev := XMM()
	MOVUPS(Mem{Base: src}, prev)

	// Advance to the second item.
	ADDQ(Imm(16), src)

	// Keep a separate write pointer.
	dst := GP64()
	MOVQ(src, dst)

	// Loop until we're done.
	Label("loop")
	CMPQ(src, end)
	JE(LabelRef("done"))

	// Load the item at the input pointer.
	item := XMM()
	MOVUPS(Mem{Base: src}, item)

	// Compare item == prev by comparing the two qwords that make up the 16 byte item.
	result := XMM()
	MOVUPS(item, result)
	PCMPEQQ(prev, result)

	// Extract the equality mask, where 0x3 indicates that both qword components are equal.
	mask := GP32()
	MOVMSKPD(result, mask)
	CMPL(mask, Imm(3))
	JE(LabelRef("next")) // skip the write if they're equal

	// Write item to dst and advance the write pointer.
	MOVUPS(item, Mem{Base: dst})
	ADDQ(Imm(16), dst)
	MOVUPS(item, prev)

	// Advance the input pointer and loop.
	Label("next")
	ADDQ(Imm(16), src)
	JMP(LabelRef("loop"))

	Label("done")

	// Calculate and return byte offset of the dst pointer.
	base := Load(Param("b").Base(), GP64())
	SUBQ(base, dst)
	Store(dst, Return("pos"))

	RET()
	Generate()
}
