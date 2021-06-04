// +build !amd64

package main

import (
	"fmt"
	"math"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func main() {
	insertionsort(32, YMM)
	distributeForward(32, YMM)
	distributeBackward(32, YMM)

	insertionsort(16, XMM)
	distributeForward(16, XMM)
	distributeBackward(16, XMM)

	Generate()
}

func shiftForSize(size uint64) uint64 {
	return uint64(math.Log2(float64(size)))
}

func less(size uint64, register func() VecVirtual, a, b, ones Op, withMasks func (neMask, ltMask Op)) {
	ne := register()
	lt := register()
	VPCMPEQB(a, b, ne)
	VPXOR(ne, ones, ne)
	VPMINUB(a, b, lt)
	VPCMPEQB(a, lt, lt)
	VPAND(lt, ne, lt)
	neMask := GP32()
	ltMask := GP32()
	VPMOVMSKB(ne, neMask)
	VPMOVMSKB(lt, ltMask)
	withMasks(neMask, ltMask)
	unequalByteIndex := GP32()
	BSFL(neMask, unequalByteIndex)
	BTSL(unequalByteIndex, ltMask)
}

func insertionsort(size uint64, register func() VecVirtual) {
	TEXT(fmt.Sprintf("insertionsort%d", size), NOSPLIT, "func(data *byte, lo, hi int)")

	shift := shiftForSize(size)

	data := Load(Param("data"), GP64())
	loIndex := Load(Param("lo"), GP64())
	hiIndex := Load(Param("hi"), GP64())
	SHLQ(Imm(shift), loIndex)
	SHLQ(Imm(shift), hiIndex)
	lo := GP64()
	hi := GP64()
	LEAQ(Mem{Base: data, Index: loIndex, Scale: 1}, lo)
	LEAQ(Mem{Base: data, Index: hiIndex, Scale: 1}, hi)

	ones := register()
	VPCMPEQB(ones, ones, ones)

	i := GP64()
	MOVQ(lo, i)

	Label("outer")
	ADDQ(Imm(size), i)
	CMPQ(i, hi)
	JA(LabelRef("done"))
	item := register()
	VMOVDQU(Mem{Base: i}, item)
	j := GP64()
	MOVQ(i, j)

	Label("inner")
	prev := register()
	VMOVDQU(Mem{Base: j, Disp: -int(size)}, prev)

	less(size, register, item, prev, ones, func (neMask, ltMask Op) {
		TESTL(neMask, neMask)
		JZ(LabelRef("outer"))
	})
	JCC(LabelRef("outer"))

	VMOVDQU(prev, Mem{Base: j})
	VMOVDQU(item, Mem{Base: j, Disp: -int(size)})
	SUBQ(Imm(size), j)
	CMPQ(j, lo)
	JA(LabelRef("inner"))
	JMP(LabelRef("outer"))

	Label("done")
	if size > 16 {
		VZEROUPPER()
	}
	RET()
}

func distributeForward(size uint64, register func() VecVirtual) {
	TEXT(fmt.Sprintf("distributeForward%d", size), NOSPLIT, "func(data, scratch *byte, limit, lo, hi, pivot int) int")

	shift := shiftForSize(size)

	// Load inputs.
	data := Load(Param("data"), GP64())
	scratch := Load(Param("scratch"), GP64())
	limit := Load(Param("limit"), GP64())
	loIndex := Load(Param("lo"), GP64())
	hiIndex := Load(Param("hi"), GP64())
	pivotIndex := Load(Param("pivot"), GP64())

	// Convert indices to byte offsets.
	SHLQ(Imm(shift), limit)
	SHLQ(Imm(shift), loIndex)
	SHLQ(Imm(shift), hiIndex)
	SHLQ(Imm(shift), pivotIndex)

	// Prepare read/cmp pointers.
	lo := GP64()
	hi := GP64()
	tail := GP64()
	LEAQ(Mem{Base: data, Index: loIndex, Scale: 1}, lo)
	LEAQ(Mem{Base: data, Index: hiIndex, Scale: 1}, hi)
	LEAQ(Mem{Base: scratch, Index: limit, Scale: 1, Disp: -int(size)}, tail)

	ones := register()
	VPCMPEQB(ones, ones, ones)

	// Load the pivot item.
	pivot := register()
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
	next := register()
	VMOVDQU(Mem{Base: lo}, next)

	// Compare the item with the pivot.
	hasUnequalByte := GP8()
	less(size, register, next, pivot, ones, func (neMask, ltMask Op) {
		TESTL(neMask, neMask)
		SETNE(hasUnequalByte)
	})
	SETCS(isLess.As8())
	ANDB(hasUnequalByte, isLess.As8())
	XORB(Imm(1), isLess.As8())

	// Conditionally write to either the beginning of the data slice, or
	// end of the scratch slice.
	dst := GP64()
	MOVQ(lo, dst)
	CMOVQNE(tail, dst)
	VMOVDQU(next, Mem{Base: dst, Index: offset, Scale: 1})
	SHLQ(Imm(shift), isLess)
	SUBQ(isLess, offset)
	ADDQ(Imm(size), lo)

	// Loop while we have more input, and enough room in the scratch slice.
	CMPQ(lo, hi)
	JA(LabelRef("done"))
	CMPQ(offset, limit)
	JNE(LabelRef("loop"))

	// Return the number of items written to the data slice.
	Label("done")
	SUBQ(data, lo)
	ADDQ(offset, lo)
	SHRQ(Imm(shift), lo)
	DECQ(lo)
	Store(lo, ReturnIndex(0))
	if size > 16 {
		VZEROUPPER()
	}
	RET()
}

func distributeBackward(size uint64, register func() VecVirtual) {
	TEXT(fmt.Sprintf("distributeBackward%d", size), NOSPLIT, "func(data, scratch *byte, limit, lo, hi, pivot int) int")

	shift := shiftForSize(size)

	// Load inputs.
	data := Load(Param("data"), GP64())
	scratch := Load(Param("scratch"), GP64())
	limit := Load(Param("limit"), GP64())
	loIndex := Load(Param("lo"), GP64())
	hiIndex := Load(Param("hi"), GP64())
	pivotIndex := Load(Param("pivot"), GP64())

	// Convert indices to byte offsets.
	SHLQ(Imm(shift), limit)
	SHLQ(Imm(shift), loIndex)
	SHLQ(Imm(shift), hiIndex)
	SHLQ(Imm(shift), pivotIndex)

	// Prepare read/cmp pointers.
	lo := GP64()
	hi := GP64()
	LEAQ(Mem{Base: data, Index: loIndex, Scale: 1}, lo)
	LEAQ(Mem{Base: data, Index: hiIndex, Scale: 1}, hi)

	ones := register()
	VPCMPEQB(ones, ones, ones)

	// Load the pivot item.
	pivot := register()
	VMOVDQU(Mem{Base: data, Index: pivotIndex, Scale: 1}, pivot)

	offset := GP64()
	isLess := GP64()
	XORQ(offset, offset)
	XORQ(isLess, isLess)

	CMPQ(hi, lo)
	JBE(LabelRef("done"))

	Label("loop")

	// Load the next item.
	next := register()
	VMOVDQU(Mem{Base: hi}, next)

	// Compare the item with the pivot.
	hasUnequalByte := GP8()
	less(size, register, next, pivot, ones, func (neMask, ltMask Op) {
		TESTL(neMask, neMask)
		SETNE(hasUnequalByte)
	})
	SETCS(isLess.As8())
	ANDB(hasUnequalByte, isLess.As8())

	// Conditionally write to either the end of the data slice, or
	// beginning of the scratch slice.
	dst := GP64()
	MOVQ(scratch, dst)
	CMOVQEQ(hi, dst)
	VMOVDQU(next, Mem{Base: dst, Index: offset, Scale: 1})
	SHLQ(Imm(shift), isLess)
	ADDQ(isLess, offset)
	SUBQ(Imm(size), hi)

	// Loop while we have more input, and enough room in the scratch slice.
	CMPQ(hi, lo)
	JBE(LabelRef("done"))
	CMPQ(offset, limit)
	JNE(LabelRef("loop"))

	// Return the number of items written to the data slice.
	Label("done")
	SUBQ(data, hi)
	ADDQ(offset, hi)
	SHRQ(Imm(shift), hi)
	Store(hi, ReturnIndex(0))
	if size > 16 {
		VZEROUPPER()
	}
	RET()
}
