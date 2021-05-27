// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("union16", NOSPLIT, "func(dst, a, b []byte) (i, j, k int)")

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
	aItem := XMM()
	bItem := XMM()
	VMOVUPS(Mem{Base: a}, aItem)
	VMOVUPS(Mem{Base: b}, bItem)

	// Compare bytes from each side and extract an equality mask.
	result := XMM()
	VPCMPEQB(aItem, bItem, result)
	mask := GP32()
	VPMOVMSKB(result, mask)

	// If a==b, copy either to dst and advance all pointers.
	Label("check_equal")
	CMPL(mask, U32(0xFFFF))
	JNE(LabelRef("check_greater"))
	VMOVUPS(aItem, Mem{Base: dst})
	ADDQ(Imm(16), dst)
	ADDQ(Imm(16), a)
	ADDQ(Imm(16), b)
	JMP(LabelRef("loop"))

	// Otherwise, if a>b, copy and advance b.
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
	VMOVUPS(bItem, Mem{Base: dst})
	ADDQ(Imm(16), dst)
	ADDQ(Imm(16), b)
	JMP(LabelRef("loop"))

	// Otherwise (if a<b), copy and advance a.
	Label("less")
	VMOVUPS(aItem, Mem{Base: dst})
	ADDQ(Imm(16), dst)
	ADDQ(Imm(16), a)
	JMP(LabelRef("loop"))

	// Calculate and return byte offsets of the each pointer.
	Label("done")
	SUBQ(Load(Param("a").Base(), GP64()), a)
	Store(a, Return("i"))
	SUBQ(Load(Param("b").Base(), GP64()), b)
	Store(b, Return("j"))
	SUBQ(Load(Param("dst").Base(), GP64()), dst)
	Store(dst, Return("k"))
	VZEROUPPER()
	RET()
	Generate()
}
