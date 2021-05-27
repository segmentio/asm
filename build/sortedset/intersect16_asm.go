// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("intersect16", NOSPLIT, "func(dst, a, b []byte) int")

	// Load all pointers.
	dst := Load(Param("dst").Base(), GP64())
	a := Load(Param("a").Base(), GP64())
	b := Load(Param("b").Base(), GP64())

	// Calculate the end of a/b so we know where to loop until.
	aEnd := Load(Param("a").Len(), GP64())
	ADDQ(a, aEnd)
	bEnd := Load(Param("b").Len(), GP64())
	ADDQ(b, bEnd)

	// Loop until we're at the end of either a or b.
	Label("loop")
	CMPQ(a, aEnd)
	JE(LabelRef("done"))
	CMPQ(b, bEnd)
	JE(LabelRef("done"))

	// Load the next item from one side.
	item := XMM()
	VMOVUPS(Mem{Base: a}, item)

	// Compare bytes from each side and extract an equality mask.
	result := XMM()
	VPCMPEQB(Mem{Base: b}, item, result)
	mask := GP32()
	VPMOVMSKB(result, mask)

	// If a==b, copy either to dst and advance all pointers.
	Label("check_equal")
	CMPL(mask, U32(0xFFFF))
	JNE(LabelRef("check_greater"))
	VMOVUPS(item, Mem{Base: dst})
	ADDQ(Imm(16), dst)
	ADDQ(Imm(16), a)
	ADDQ(Imm(16), b)
	JMP(LabelRef("loop"))

	// Otherwise, if a>b, advance b.
	// Find the first unequal byte and compare.
	Label("check_greater")
	NOTL(mask)
	unequalByteIndex := GP32()
	BSFL(mask, unequalByteIndex)
	aByte := GP8()
	bByte := GP8()
	MOVB(Mem{Base: a, Index: unequalByteIndex, Scale: 1}, aByte)
	MOVB(Mem{Base: b, Index: unequalByteIndex, Scale: 1}, bByte)
	CMPB(aByte, bByte)
	JB(LabelRef("less"))
	ADDQ(Imm(16), b)
	JMP(LabelRef("loop"))

	// Otherwise (if a<b), advance a.
	Label("less")
	ADDQ(Imm(16), a)
	JMP(LabelRef("loop"))

	// Calculate and return byte offset of the dst pointer.
	Label("done")
	SUBQ(Load(Param("dst").Base(), GP64()), dst)
	Store(dst, ReturnIndex(0))
	VZEROUPPER()
	RET()
	Generate()
}
