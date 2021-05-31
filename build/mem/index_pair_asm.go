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
}

type indexPair1 struct{}

func (indexPair1) size() int   { return 1 }
func (indexPair1) reg() Op     { return GP8() }
func (indexPair1) mov(a, b Op) { MOVB(a, b) }
func (indexPair1) cmp(a, b Op) { CMPB(a, b) }

type indexPair2 struct{}

func (indexPair2) size() int   { return 2 }
func (indexPair2) reg() Op     { return GP16() }
func (indexPair2) mov(a, b Op) { MOVW(a, b) }
func (indexPair2) cmp(a, b Op) { CMPW(a, b) }

type indexPair4 struct{}

func (indexPair4) size() int   { return 4 }
func (indexPair4) reg() Op     { return GP32() }
func (indexPair4) mov(a, b Op) { MOVL(a, b) }
func (indexPair4) cmp(a, b Op) { CMPL(a, b) }

type indexPair8 struct{}

func (indexPair8) size() int   { return 8 }
func (indexPair8) reg() Op     { return GP64() }
func (indexPair8) mov(a, b Op) { MOVQ(a, b) }
func (indexPair8) cmp(a, b Op) { CMPQ(a, b) }

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

func generateIndexPair(code indexPair) {
	size := code.size()
	TEXT(fmt.Sprintf("indexPair%d", size), NOSPLIT, "func(b []byte) int")

	p := Load(Param("b").Base(), GP64())
	n := Load(Param("b").Len(), GP64())

	CMPQ(n, Imm(uint64(size))) // zero or one item
	JBE(LabelRef("done"))

	ptr := GP64()
	end := GP64()
	MOVQ(p, ptr)
	MOVQ(p, end)
	ADDQ(n, end)

	p0 := ptr
	p1 := GP64()
	p2 := GP64()
	p3 := GP64()
	MOVQ(p0, p1)
	ADDQ(Imm(uint64(size)), p1)

	if size == 1 {
		CMPQ(n, Imm(64))
		JBE(LabelRef("generic"))
		JumpIfFeature("avx2", cpu.AVX2)
	}

	Label("generic")
	r0 := code.reg()
	r1 := code.reg()
	code.mov(Mem{Base: p0}, r0)
	code.mov(Mem{Base: p1}, r1)
	code.cmp(r0, r1)
	JE(LabelRef("found"))
	ADDQ(Imm(uint64(size)), p0)
	ADDQ(Imm(uint64(size)), p1)
	CMPQ(p0, end)
	JNE(LabelRef("generic"))

	Label("done")
	Store(n, ReturnIndex(0))
	RET()

	Label("found")
	// The delta between the base pointer and how far we advanced is the index of the pair.
	SUBQ(p, p0)
	Store(p0, ReturnIndex(0))
	RET()

	Label("avx2")
	// limit := end - size
	limit := GP64()
	MOVQ(end, limit)
	SUBQ(Imm(32), limit)

	/*
		MOVQ(p0, p2)
		MOVQ(p1, p3)
		ADDQ(Imm(32), p2)
		ADDQ(Imm(32), p3)
	*/
	_ = p2
	_ = p3

	Label("avx2_loop")
	y0 := YMM()
	y1 := YMM()
	//y2 := YMM()
	//y3 := YMM()
	mask0 := GP64()
	//mask1 := GP64()

	VMOVDQU(Mem{Base: p0}, y0)
	VMOVDQU(Mem{Base: p1}, y1)
	//VMOVDQU(Mem{Base: p2}, y2)
	//VMOVDQU(Mem{Base: p3}, y3)

	VPCMPEQB(y0, y1, y1)
	//VPCMPEQB(y2, y3, y3)

	VPMOVMSKB(y1, mask0.As32())
	//VPMOVMSKB(y3, mask1.As32())

	TZCNTQ(mask0, mask0)
	//TZCNTQ(mask1, mask1)

	//	SHLQ(Imm(32), mask1)
	//	ORQ(mask1, mask0)
	CMPQ(mask0, Imm(64))
	JNE(LabelRef("avx2_found"))

	//ADDQ(Imm(64), p0)
	//ADDQ(Imm(64), p1)
	//ADDQ(Imm(64), p2)
	//ADDQ(Imm(64), p3)
	//CMPQ(p3, limit)
	ADDQ(Imm(32), p0)
	ADDQ(Imm(32), p1)
	CMPQ(p1, limit)
	JBE(LabelRef("avx2_loop"))

	VZEROUPPER()
	//MOVQ(p2, p0)
	//MOVQ(p3, p1)
	CMPQ(p1, end)
	JB(LabelRef("generic"))
	JMP(LabelRef("done"))

	Label("avx2_found")
	i := GP64()
	MOVQ(U32(64), i)
	SUBQ(mask0, i)
	ADDQ(i, p0)
	JMP(LabelRef("found"))
}
