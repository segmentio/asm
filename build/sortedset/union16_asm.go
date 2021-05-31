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

	ones := XMM()
	VPCMPEQB(ones, ones, ones)

	// Load the first item from a/b. We know that each has at least
	// one item (this is enforced in the wrapper).
	aItem := XMM()
	bItem := XMM()
	VMOVUPS(Mem{Base: a}, aItem)
	VMOVUPS(Mem{Base: b}, bItem)

	Label("loop")

	// Compare bytes and extract two masks.
	// ne = mask of bytes where a!=b
	// lt = mask of bytes where a<b
	ne := XMM()
	lt := XMM()
	VPCMPEQB(aItem, bItem, ne)
	VPXOR(ne, ones, ne)
	VPMINUB(aItem, bItem, lt)
	VPCMPEQB(aItem, lt, lt)
	VPAND(lt, ne, lt)
	unequalMask := GP32()
	lessMask := GP32()
	VPMOVMSKB(ne, unequalMask)
	VPMOVMSKB(lt, lessMask)

	// Branch based on whether a==b, or a<b.
	CMPL(unequalMask, U32(0))
	JE(LabelRef("equal"))
	unequalByteIndex := GP32()
	BSFL(unequalMask, unequalByteIndex)
	BTSL(unequalByteIndex, lessMask)
	JCS(LabelRef("less"))

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
