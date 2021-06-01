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

	p := Load(Param("b").Base(), GP64())
	n := Load(Param("b").Len(), GP64())

	base := GP64()
	MOVQ(p, base)

	CMPQ(n, Imm(0))
	JLE(LabelRef("done"))
	SUBQ(Imm(uint64(size)), n)

	if size < 16 {
		JumpIfFeature("avx2", cpu.AVX2)
	}

	Label("tail")
	CMPQ(n, Imm(0))
	JE(LabelRef("fail"))

	Label("generic")
	r0 := code.reg()
	r1 := code.reg()
	code.mov(Mem{Base: p}, r0)
	code.mov((Mem{Base: p}).Offset(size), r1)
	code.cmp(r0, r1)
	JE(LabelRef("done"))
	ADDQ(Imm(uint64(size)), p)
	SUBQ(Imm(uint64(size)), n)
	CMPQ(n, Imm(0))
	JA(LabelRef("generic"))

	Label("fail")
	ADDQ(Imm(uint64(size)), p)

	Label("done")
	// The delta between the base pointer and how far we advanced is the index of the pair.
	index := p
	SUBQ(base, index)
	Store(index, ReturnIndex(0))
	RET()

	switch size {
	case 1, 2, 4, 8:
	default:
		return
	}

	const avxChunk = 256
	const avxLanes = avxChunk / 32
	Label("avx2")
	CMPQ(n, U32(avxChunk+uint64(size)))
	JB(LabelRef(fmt.Sprintf("avx2_tail%d", avxChunk/2)))

	masks := make([]GPVirtual, avxLanes)
	for i := range masks {
		masks[i] = GP64()
		MOVQ(U64(0), masks[i])
	}

	regA := make([]VecVirtual, avxLanes)
	regB := make([]VecVirtual, avxLanes)
	for i := range regA {
		regA[i] = YMM()
		regB[i] = YMM()
	}

	Label(fmt.Sprintf("avx2_loop%d", avxChunk))
	/*
		switch size {
		case 16:
				VPCMPEQQ(Y0, Y1, Y1)
				VPMOVMSKB(Y1, mask.As32())

				hi := GP64()
				lo := GP64()
				MOVQ(mask, hi)
				MOVQ(mask, lo)
				SHRQ(Imm(16), hi)
				ANDQ(U32(0xFFFF), lo)

				MOVQ(U64(0), mask)
				CMPQ(lo, U32(0xFFFF))
				JE(LabelRef("avx2_found"))

				MOVQ(U64(16), mask)
				CMPQ(hi, U32(0xFFFF))
				JE(LabelRef("avx2_found"))
		}
	*/

	generateIndexPairAVX2(p, regA, regB, masks, code)
	ADDQ(U32(avxChunk), p)
	SUBQ(U32(avxChunk), n)
	CMPQ(n, U32(avxChunk+uint64(size)))
	JAE(LabelRef(fmt.Sprintf("avx2_loop%d", avxChunk)))

	for chunk := avxChunk / 2; chunk >= 32; chunk /= 2 {
		Label(fmt.Sprintf("avx2_tail%d", chunk))
		CMPQ(n, Imm(uint64(chunk+size)))
		JB(LabelRef(fmt.Sprintf("avx2_tail%d", chunk/2)))
		lanes := chunk / 32
		generateIndexPairAVX2(p, regA[:lanes], regB[:lanes], masks[:lanes], code)
		ADDQ(U32(uint64(chunk)), p)
		SUBQ(U32(uint64(chunk)), n)
	}

	Label("avx2_tail16")
	CMPQ(n, Imm(uint64(16+size)))
	JB(LabelRef("avx2_tail"))
	generateIndexPairAVX2(p, []VecVirtual{XMM()}, []VecVirtual{XMM()}, masks[:1], code)
	ADDQ(Imm(16), p)
	SUBQ(Imm(16), n)

	Label("avx2_tail")
	VZEROUPPER()
	JMP(LabelRef("tail"))

	for i, mask := range masks {
		Label(fmt.Sprintf("avx2_done%d", i))
		if i > 0 {
			ADDQ(U32(uint64(i*32)), p)
			SUBQ(U32(uint64(i*32)), n)
		}
		ADDQ(mask, p)
		SUBQ(mask, n)
		VZEROUPPER()
		JMP(LabelRef("done"))
	}
}

func generateIndexPairAVX2(p Register, regA, regB []VecVirtual, masks []GPVirtual, code indexPair) {
	size := code.size()
	for i, reg := range regA {
		VMOVDQU((Mem{Base: p}).Offset(i*32), reg)
	}
	for i, reg := range regB {
		VMOVDQU((Mem{Base: p}).Offset(i*32+size), reg)
	}
	for i := range regA {
		code.vpcmpeq(regA[i], regB[i], regB[i])
	}
	for i := range regB {
		VPMOVMSKB(regB[i], masks[i].As32())
	}
	for _, mask := range masks {
		TZCNTQ(mask, mask)
	}
	for i, mask := range masks {
		CMPQ(mask, Imm(64))
		JNE(LabelRef(fmt.Sprintf("avx2_done%d", i)))
	}
}
