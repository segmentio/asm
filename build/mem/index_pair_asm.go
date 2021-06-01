// +build ignore

package main

import (
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/x86"
	"github.com/segmentio/asm/cpu"
)

func main() {
	generateIndexPair(indexPair1{})
	generateIndexPair(indexPair2{})
	generateIndexPair(indexPair4{})
	generateIndexPair(indexPair8{})
	generateIndexPair(indexPair16{})
	Generate()
}

type indexPair interface {
	size() int
	reg() Op
	mov(Op, Op)
	cmp(Op, Op)
	vpcmpeq(Op, Op, Op)
}

type indexPair1 struct{}

func (indexPair1) size() int          { return 1 }
func (indexPair1) reg() Op            { return GP8() }
func (indexPair1) mov(a, b Op)        { MOVB(a, b) }
func (indexPair1) cmp(a, b Op)        { CMPB(a, b) }
func (indexPair1) vpcmpeq(a, b, c Op) { VPCMPEQB(a, b, c) }

type indexPair2 struct{}

func (indexPair2) size() int          { return 2 }
func (indexPair2) reg() Op            { return GP16() }
func (indexPair2) mov(a, b Op)        { MOVW(a, b) }
func (indexPair2) cmp(a, b Op)        { CMPW(a, b) }
func (indexPair2) vpcmpeq(a, b, c Op) { VPCMPEQW(a, b, c) }

type indexPair4 struct{}

func (indexPair4) size() int          { return 4 }
func (indexPair4) reg() Op            { return GP32() }
func (indexPair4) mov(a, b Op)        { MOVL(a, b) }
func (indexPair4) cmp(a, b Op)        { CMPL(a, b) }
func (indexPair4) vpcmpeq(a, b, c Op) { VPCMPEQD(a, b, c) }

type indexPair8 struct{}

func (indexPair8) size() int          { return 8 }
func (indexPair8) reg() Op            { return GP64() }
func (indexPair8) mov(a, b Op)        { MOVQ(a, b) }
func (indexPair8) cmp(a, b Op)        { CMPQ(a, b) }
func (indexPair8) vpcmpeq(a, b, c Op) { VPCMPEQQ(a, b, c) }

type indexPair16 struct{}

func (indexPair16) size() int   { return 16 }
func (indexPair16) reg() Op     { return XMM() }
func (indexPair16) mov(a, b Op) { MOVOU(a, b) }
func (ins indexPair16) cmp(a, b Op) {
	r := GP32()
	PCMPEQQ(a, b)
	PMOVMSKB(b, r)
	CMPL(r, U32(0xFFFF))
}
func (indexPair16) vpcmpeq(a, b, c Op) { panic("NOT IMPLEMENTED") }

func generateIndexPair(code indexPair) {
	size := code.size()
	TEXT(fmt.Sprintf("indexPair%d", size), NOSPLIT, "func(b []byte) int")

	base := Load(Param("b").Base(), GP64())
	count := Load(Param("b").Len(), GP64())

	CMPQ(count, Imm(uint64(size))) // zero or one item
	JBE(LabelRef("done"))

	ptr := GP64()
	end := GP64()
	MOVQ(base, ptr)
	MOVQ(base, end)
	ADDQ(count, end)
	SUBQ(Imm(uint64(size)), end)

	if size < 16 {
		CMPQ(count, Imm(32+uint64(size)))
		JBE(LabelRef("generic"))
		JumpIfFeature("avx2", cpu.AVX2)
	}

	Label("generic")
	r0 := code.reg()
	r1 := code.reg()
	code.mov(Mem{Base: ptr}, r0)
	code.mov((Mem{Base: ptr}).Offset(size), r1)
	code.cmp(r0, r1)
	JE(LabelRef("found"))
	ADDQ(Imm(uint64(size)), ptr)
	CMPQ(ptr, end)
	JNE(LabelRef("generic"))

	Label("done")
	Store(count, ReturnIndex(0))
	RET()

	Label("found")
	// The delta between the base pointer and how far we advanced is the index of the pair.
	index := ptr
	SUBQ(base, index)
	Store(index, ReturnIndex(0))
	RET()

	if size < 16 {
		Label("avx2")
		limit := GP64()
		MOVQ(end, limit)
		SUBQ(Imm(32+uint64(size)), limit)
		mask := GP64()
		MOVQ(U64(0), mask)

		Label("avx2_loop")
		VMOVDQU(Mem{Base: ptr}, Y0)
		VMOVDQU((Mem{Base: ptr}).Offset(size), Y1)
		code.vpcmpeq(Y0, Y1, Y1)
		VPMOVMSKB(Y1, mask.As32())
		TZCNTQ(mask, mask)
		CMPQ(mask, Imm(64))
		JNE(LabelRef("avx2_found"))
		ADDQ(Imm(32), ptr)
		CMPQ(ptr, limit)
		JBE(LabelRef("avx2_loop"))

		VZEROUPPER()
		CMPQ(ptr, end)
		JB(LabelRef("generic"))
		JMP(LabelRef("done"))

		Label("avx2_found")
		VZEROUPPER()
		ADDQ(mask, ptr)
		JMP(LabelRef("found"))
	}
}
