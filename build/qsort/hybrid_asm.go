// build !amd64

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	insertionsort32()
	distributeForward32()
	distributeBackward32()

	Generate()
}

func insertionsort32() {
	TEXT("insertionsort32", NOSPLIT, "func(data *byte, lo, hi int)")

	data := Load(Param("data"), GP64())
	loIndex := Load(Param("lo"), GP64())
	hiIndex := Load(Param("hi"), GP64())
	SHLQ(Imm(5), loIndex)
	SHLQ(Imm(5), hiIndex)
	lo := GP64()
	hi := GP64()
	LEAQ(Mem{Base: data, Index: loIndex, Scale: 1}, lo)
	LEAQ(Mem{Base: data, Index: hiIndex, Scale: 1}, hi)

	i := GP64()
	MOVQ(lo, i)

	Label("outer")
	ADDQ(Imm(32), i)
	CMPQ(i, hi)
	JA(LabelRef("done"))
	item := YMM()
	VMOVDQU(Mem{Base: i}, item)
	j := GP64()
	MOVQ(i, j)

	Label("inner")
	prev := YMM()
	VMOVDQU(Mem{Base: j, Disp: -32}, prev)

	lte := YMM()
	eq := YMM()
	VPMINUB(item, prev, lte)
	VPCMPEQB(item, prev, eq)
	VPCMPEQB(item, lte, lte)
	eqMask := GP32()
	lteMask := GP32()
	VPMOVMSKB(lte, lteMask)
	VPMOVMSKB(eq, eqMask)
	XORL(U32(0xFFFFFFFF), eqMask)
	JZ(LabelRef("outer"))
	ANDL(eqMask, lteMask)
	BSFL(eqMask, eqMask)
	BSFL(lteMask, lteMask)
	CMPL(eqMask, lteMask)
	JNE(LabelRef("outer"))

	VMOVDQU(prev, Mem{Base: j})
	VMOVDQU(item, Mem{Base: j, Disp: -32})
	SUBQ(Imm(32), j)
	CMPQ(j, lo)
	JA(LabelRef("inner"))
	JMP(LabelRef("outer"))

	Label("done")
	VZEROUPPER()
	RET()
}

func distributeForward32() {
	TEXT("distributeForward32", NOSPLIT, "func(data, scratch *byte, limit, lo, hi, pivot int) int")

	// Load inputs.
	data := Load(Param("data"), GP64())
	scratch := Load(Param("scratch"), GP64())
	limit := Load(Param("limit"), GP64())
	loIndex := Load(Param("lo"), GP64())
	hiIndex := Load(Param("hi"), GP64())
	pivotIndex := Load(Param("pivot"), GP64())

	// Convert indices to offsets (shift left by 5 == multiply by 32).
	SHLQ(Imm(5), limit)
	SHLQ(Imm(5), loIndex)
	SHLQ(Imm(5), hiIndex)
	SHLQ(Imm(5), pivotIndex)

	// Prepare read/cmp pointers.
	lo := GP64()
	hi := GP64()
	tail := GP64()
	LEAQ(Mem{Base: data, Index: loIndex, Scale: 1}, lo)
	LEAQ(Mem{Base: data, Index: hiIndex, Scale: 1}, hi)
	LEAQ(Mem{Base: scratch, Index: limit, Scale: 1, Disp: -32}, tail)

	// Load the pivot item.
	pivot := YMM()
	VMOVDQU(Mem{Base: data, Index: pivotIndex, Scale: 1}, pivot)

	offset := GP64()
	isLess := GP64()
	XORQ(offset, offset)
	XORQ(isLess, isLess)

	// We'll be keeping a negative offset. Negate the limit so we can
	// compare the two in the loop.
	NEGQ(limit)

	Label("loop")

	// Load the next item.
	next := YMM()
	VMOVDQU(Mem{Base: lo}, next)

	// Compare the item with the pivot.
	lte := YMM()
	eq := YMM()
	VPMINUB(next, pivot, lte)
	VPCMPEQB(next, pivot, eq)
	VPCMPEQB(next, lte, lte)
	eqMask := GP32()
	lteMask := GP32()
	VPMOVMSKB(lte, lteMask)
	VPMOVMSKB(eq, eqMask)
	XORL(U32(0xFFFFFFFF), eqMask)
	ANDL(eqMask, lteMask)
	hasUnequalByte := GP8()
	SETNE(hasUnequalByte)
	BSFL(eqMask, eqMask)
	BSFL(lteMask, lteMask)
	CMPL(eqMask, lteMask)
	SETEQ(isLess.As8())
	ANDB(hasUnequalByte, isLess.As8())
	XORB(Imm(1), isLess.As8())

	// Conditionally write to either the beginning of the data slice, or
	// end of the scratch slice.
	dst := GP64()
	MOVQ(lo, dst)
	CMOVQNE(tail, dst)
	VMOVDQU(next, Mem{Base: dst, Index: offset, Scale: 1})
	SHLQ(Imm(5), isLess)
	SUBQ(isLess, offset)
	ADDQ(Imm(32), lo)

	// Loop while we have more input, and enough room in the scratch slice.
	CMPQ(lo, hi)
	JA(LabelRef("done"))
	CMPQ(offset, limit)
	JNE(LabelRef("loop"))

	// Return the number of items written to the data slice.
	Label("done")
	SUBQ(data, lo)
	ADDQ(offset, lo)
	SHRQ(Imm(5), lo)
	DECQ(lo)
	Store(lo, ReturnIndex(0))
	VZEROUPPER()
	RET()
}

func distributeBackward32() {
	TEXT("distributeBackward32", NOSPLIT, "func(data, scratch *byte, limit, lo, hi, pivot int) int")

	// Load inputs.
	data := Load(Param("data"), GP64())
	scratch := Load(Param("scratch"), GP64())
	limit := Load(Param("limit"), GP64())
	loIndex := Load(Param("lo"), GP64())
	hiIndex := Load(Param("hi"), GP64())
	pivotIndex := Load(Param("pivot"), GP64())

	// Convert indices to offsets (shift left by 5 == multiply by 32).
	SHLQ(Imm(5), limit)
	SHLQ(Imm(5), loIndex)
	SHLQ(Imm(5), hiIndex)
	SHLQ(Imm(5), pivotIndex)

	// Prepare read/cmp pointers.
	lo := GP64()
	hi := GP64()
	LEAQ(Mem{Base: data, Index: loIndex, Scale: 1}, lo)
	LEAQ(Mem{Base: data, Index: hiIndex, Scale: 1}, hi)

	// Load the pivot item.
	pivot := YMM()
	VMOVDQU(Mem{Base: data, Index: pivotIndex, Scale: 1}, pivot)

	offset := GP64()
	isLess := GP64()
	XORQ(offset, offset)
	XORQ(isLess, isLess)

	CMPQ(hi, lo)
	JBE(LabelRef("done"))

	Label("loop")

	// Compare the item with the pivot.
	next := YMM()
	VMOVDQU(Mem{Base: hi}, next)
	lte := YMM()
	eq := YMM()
	VPMINUB(next, pivot, lte)
	VPCMPEQB(next, pivot, eq)
	VPCMPEQB(next, lte, lte)
	eqMask := GP32()
	lteMask := GP32()
	VPMOVMSKB(lte, lteMask)
	VPMOVMSKB(eq, eqMask)
	XORL(U32(0xFFFFFFFF), eqMask)
	ANDL(eqMask, lteMask)
	hasUnequalByte := GP8()
	SETNE(hasUnequalByte)
	BSFL(eqMask, eqMask)
	BSFL(lteMask, lteMask)
	CMPL(eqMask, lteMask)
	SETEQ(isLess.As8())
	ANDB(hasUnequalByte, isLess.As8())

	// Conditionally write to either the end of the data slice, or
	// beginning of the scratch slice.
	dst := GP64()
	MOVQ(scratch, dst)
	CMOVQEQ(hi, dst)
	VMOVDQU(next, Mem{Base: dst, Index: offset, Scale: 1})
	SHLQ(Imm(5), isLess)
	ADDQ(isLess, offset)
	SUBQ(Imm(32), hi)

	// Loop while we have more input, and enough room in the scratch slice.
	CMPQ(hi, lo)
	JBE(LabelRef("done"))
	CMPQ(offset, limit)
	JNE(LabelRef("loop"))

	// Return the number of items written to the data slice.
	Label("done")
	SUBQ(data, hi)
	ADDQ(offset, hi)
	SHRQ(Imm(5), hi)
	Store(hi, ReturnIndex(0))
	VZEROUPPER()
	RET()
}
