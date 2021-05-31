// +build ignore

package main

import (
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"

	//. "github.com/mmcloughlin/avo/reg"
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
		CMPQ(count, Imm(64))
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
		// limit := end - size
		limit := GP64()
		MOVQ(end, limit)
		SUBQ(Imm(32+uint64(size)), limit)
		//SUBQ(Imm(64), limit)

		Label("avx2_loop")
		y0 := YMM()
		y1 := YMM()
		//y2 := YMM()
		//y3 := YMM()
		mask0 := GP64()
		//mask1 := GP64()

		VMOVDQU(Mem{Base: ptr}, y0)
		VMOVDQU((Mem{Base: ptr}).Offset(size), y1)

		code.vpcmpeq(y0, y1, y1)
		//code.vpcmpeq(y2, y3, y3)

		//XORQ(mask0, mask0)
		//XORQ(mask1, mask1)
		VPMOVMSKB(y1, mask0.As32())
		//VPMOVMSKB(y3, mask1.As32())
		//SHLQ(Imm(32), mask1)
		//ORQ(mask1, mask0)

		TZCNTQ(mask0, mask0)
		CMPQ(mask0, Imm(64))
		JNE(LabelRef("avx2_found"))

		//ADDQ(Imm(64), ptr)
		//ADDQ(Imm(64), p1)
		ADDQ(Imm(32), ptr)
		//ADDQ(Imm(32), p1)
		CMPQ(ptr, limit)
		JBE(LabelRef("avx2_loop"))

		VZEROUPPER()
		CMPQ(ptr, end)
		JB(LabelRef("generic"))
		JMP(LabelRef("done"))

		Label("avx2_found")
		ADDQ(mask0, ptr) // ptr += trailingZeros
		JMP(LabelRef("found"))
	}
}
