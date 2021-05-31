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

	// Load the first item from a/b. We know that each has at least
	// one item (this is enforced in the wrapper).
	aItem := XMM()
	bItem := XMM()
	VMOVUPS(Mem{Base: a}, aItem)
	VMOVUPS(Mem{Base: b}, bItem)

	Label("loop")

	// Compare bytes and extract an equality mask.
	result := XMM()
	VPCMPEQB(aItem, bItem, result)
	mask := GP32()
	VPMOVMSKB(result, mask)

	// Check if they're equal firstly.
	CMPL(mask, U32(0xFFFF))
	JNE(LabelRef("compare_byte"))

	// If a==b, copy either and advance both.
	Label("equal")
	VMOVUPS(aItem, Mem{Base: dst})
	ADDQ(Imm(16), dst)
	ADDQ(Imm(16), a)
	ADDQ(Imm(16), b)
	CMPQ(a, aEnd)
	JE(LabelRef("done"))
	CMPQ(b, bEnd)
	JE(LabelRef("done"))
	VMOVUPS(Mem{Base: a}, aItem)
	VMOVUPS(Mem{Base: b}, bItem)
	JMP(LabelRef("loop"))

	// They're not equal, so compare the first unequal byte.
	Label("compare_byte")
	NOTL(mask)
	unequalByteIndex := GP32()
	BSFL(mask, unequalByteIndex)
	aByte := GP8()
	bByte := GP8()
	MOVB(Mem{Base: a, Index: unequalByteIndex, Scale: 1}, aByte)
	MOVB(Mem{Base: b, Index: unequalByteIndex, Scale: 1}, bByte)
	CMPB(aByte, bByte)
	JB(LabelRef("less"))

	// If b>a, copy and advance a.
	Label("greater")
	VMOVUPS(bItem, Mem{Base: dst})
	ADDQ(Imm(16), dst)
	ADDQ(Imm(16), b)
	CMPQ(b, bEnd)
	JE(LabelRef("done"))
	VMOVUPS(Mem{Base: b}, bItem)
	JMP(LabelRef("loop"))

	// If a<b, copy and advance a.
	Label("less")
	VMOVUPS(aItem, Mem{Base: dst})
	ADDQ(Imm(16), dst)
	ADDQ(Imm(16), a)
	CMPQ(a, aEnd)
	JE(LabelRef("done"))
	VMOVUPS(Mem{Base: a}, aItem)
	JMP(LabelRef("loop"))

	// Calculate and return byte offsets of the each pointer.
	Label("done")
	SUBQ(Load(Param("a").Base(), GP64()), a)
	Store(a, Return("i"))
	SUBQ(Load(Param("b").Base(), GP64()), b)
	Store(b, Return("j"))
	SUBQ(Load(Param("dst").Base(), GP64()), dst)
	Store(dst, Return("k"))
	RET()
	Generate()
}
