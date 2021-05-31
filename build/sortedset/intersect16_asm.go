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

	// Load the first item from one side.
	item := XMM()
	VMOVUPS(Mem{Base: a}, item)

	Label("loop")

	// Compare bytes and extract an equality mask.
	result := XMM()
	VPCMPEQB(Mem{Base: b}, item, result)
	mask := GP32()
	VPMOVMSKB(result, mask)

	// Check if they're equal firstly.
	CMPL(mask, U32(0xFFFF))
	JNE(LabelRef("compare_byte"))

	// If a==b, copy either and advance both.
	VMOVUPS(item, Mem{Base: dst})
	ADDQ(Imm(16), dst)
	ADDQ(Imm(16), a)
	ADDQ(Imm(16), b)
	CMPQ(a, aEnd)
	JE(LabelRef("done"))
	CMPQ(b, bEnd)
	JE(LabelRef("done"))
	VMOVUPS(Mem{Base: a}, item)
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

	// If b<a, advance b.
	Label("greater")
	ADDQ(Imm(16), b)
	CMPQ(b, bEnd)
	JE(LabelRef("done"))
	JMP(LabelRef("loop"))

	// If a<b, advance a.
	Label("less")
	ADDQ(Imm(16), a)
	CMPQ(a, aEnd)
	JE(LabelRef("done"))
	VMOVUPS(Mem{Base: a}, item)
	JMP(LabelRef("loop"))

	// Calculate and return byte offset of the dst pointer.
	Label("done")
	SUBQ(Load(Param("dst").Base(), GP64()), dst)
	Store(dst, ReturnIndex(0))
	RET()
	Generate()
}
