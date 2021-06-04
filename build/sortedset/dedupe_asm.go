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
	generateDedupe(dedupe1{})
	generateDedupe(dedupe2{})
	generateDedupe(dedupe4{})
	generateDedupe(dedupe8{})
	generateDedupe(dedupe16{})
	generateDedupe(dedupe32{})
	Generate()
}

type dedupe interface {
	size() int
	copy(r, w, x GPVirtual)
}

type dedupeAVX2 interface {
	dedupe
}

type dedupe1 struct{}

func (dedupe1) size() int              { return 1 }
func (dedupe1) copy(r, w, x GPVirtual) { generateDedupeX86(MOVB, CMPB, GP8, r, w, x) }

type dedupe2 struct{}

func (dedupe2) size() int              { return 2 }
func (dedupe2) copy(r, w, x GPVirtual) { generateDedupeX86(MOVW, CMPW, GP16, r, w, x) }

type dedupe4 struct{}

func (dedupe4) size() int              { return 4 }
func (dedupe4) copy(r, w, x GPVirtual) { generateDedupeX86(MOVL, CMPL, GP32, r, w, x) }

type dedupe8 struct{}

func (dedupe8) size() int              { return 8 }
func (dedupe8) copy(r, w, x GPVirtual) { generateDedupeX86(MOVQ, CMPQ, GP64, r, w, x) }

type dedupe16 struct{}

func (dedupe16) size() int { return 16 }
func (dedupe16) copy(r, w, x GPVirtual) {
	a, b := XMM(), XMM()
	MOVOU(Mem{Base: r}, a)
	MOVOU(Mem{Base: w}, b)
	mask := GP32()
	PCMPEQQ(a, b)
	PMOVMSKB(b, mask)
	CMPL(mask, U32(0xFFFF))
	CMOVQNE(x, w)
	MOVOU(a, Mem{Base: w})
}

type dedupe32 struct{}

func (dedupe32) size() int { return 32 }
func (dedupe32) copy(r, w, x GPVirtual) {
	a, b, c, d := XMM(), XMM(), XMM(), XMM()
	MOVOU(Mem{Base: r}, a)
	MOVOU(Mem{Base: r}.Offset(16), b)
	MOVOU(Mem{Base: w}, c)
	MOVOU(Mem{Base: w}.Offset(16), d)
	mask0, mask1 := GP32(), GP32()
	PCMPEQQ(a, c)
	PCMPEQQ(b, d)
	PMOVMSKB(c, mask0)
	PMOVMSKB(d, mask1)
	ANDL(mask1, mask0)
	CMPL(mask0, U32(0xFFFF))
	CMOVQNE(x, w)
	MOVOU(a, Mem{Base: w})
	MOVOU(b, Mem{Base: w}.Offset(16))
}

func generateDedupe(dedupe dedupe) {
	size := dedupe.size()
	TEXT(fmt.Sprintf("dedupe%d", size), NOSPLIT, "func(b []byte) int")

	p := Load(Param("b").Base(), GP64())
	n := Load(Param("b").Len(), GP64())
	CMPQ(n, Imm(uint64(size)))
	JG(LabelRef("init"))
	Store(n, ReturnIndex(0))
	RET()

	Label("init")
	r := GP64()
	w := GP64()
	MOVQ(p, r)
	MOVQ(p, w)
	ADDQ(Imm(uint64(size)), r)
	SUBQ(Imm(uint64(size)), n)

	if _, ok := dedupe.(dedupeAVX2); ok && false {
		JumpIfFeature("avx2", cpu.AVX2)
	}

	Label("tail")
	CMPQ(n, Imm(0))
	JLE(LabelRef("done"))

	Label("generic")
	x := GP64()
	MOVQ(w, x)
	ADDQ(Imm(uint64(size)), x)
	dedupe.copy(r, w, x)
	ADDQ(Imm(uint64(size)), r)
	SUBQ(Imm(uint64(size)), n)
	CMPQ(n, Imm(0))
	JG(LabelRef("generic"))

	Label("done")
	ADDQ(Imm(uint64(size)), w)
	SUBQ(p, w)
	Store(w, ReturnIndex(0))
	RET()

	/*
		if avx, ok := dedupe.(dedupeAVX2); ok {
				const avxChunk = 256
				const avxLanes = avxChunk / 32
				Label("avx2")
				r := GP64()
				MOVQ(n, r)
				SUBQ(Imm(uint64(size)), r)
				CMPQ(r, U32(avxChunk))
				JL(LabelRef(fmt.Sprintf("avx2_tail%d", avxChunk/2)))

				masks := make([]GPVirtual, avxLanes)
				for i := range masks {
					masks[i] = GP64()
					//XORQ(masks[i], masks[i])
				}

				regA := make([]VecVirtual, avxLanes)
				regB := make([]VecVirtual, avxLanes)
				for i := range regA {
					regA[i] = YMM()
					regB[i] = YMM()
				}

				Label(fmt.Sprintf("avx2_loop%d", avxChunk))
				generateDedupeAVX2(p, i, j, regA, regB, masks, avx)
				ADDQ(U32(uint64(avxChunk)), j)
				SUBQ(U32(uint64(avxChunk)), r)
				CMPQ(r, U32(avxChunk))
				JGE(LabelRef(fmt.Sprintf("avx2_loop%d", avxChunk)))

				for chunk := avxChunk / 2; chunk >= 32; chunk /= 2 {
					Label(fmt.Sprintf("avx2_tail%d", chunk))
					CMPQ(r, Imm(uint64(chunk)))
					JL(LabelRef(fmt.Sprintf("avx2_tail%d", chunk/2)))
					lanes := chunk / 32
					generateDedupeAVX2(p, i, j, regA[:lanes], regB[:lanes], masks, avx)
					ADDQ(Imm(uint64(chunk)), j)
					SUBQ(Imm(uint64(chunk)), r)
				}

				Label("avx2_tail16")
				if size < 16 {
					CMPQ(r, Imm(uint64(16+size)))
					JL(LabelRef("avx2_tail"))
					generateDedupeAVX2(p, i, j, []VecVirtual{XMM()}, []VecVirtual{XMM()}, masks, avx)
					ADDQ(Imm(16), p)
					SUBQ(Imm(16), n)
				}

				Label("avx2_tail")
				VZEROUPPER()
				JMP(LabelRef("tail"))
		}
	*/
}

func generateDedupeX86(mov func(Op, Op), cmp func(Op, Op), reg func() GPVirtual, r, w, x GPVirtual) {
	tmp := reg()
	mov(Mem{Base: r}, tmp)
	cmp(tmp, Mem{Base: w})
	CMOVQNE(x, w)
	mov(tmp, Mem{Base: w})
}

func generateDedupeAVX2(p, i, j Register, regA, regB []VecVirtual, masks []GPVirtual, dedupe dedupeAVX2) {
	//size := dedupe.size()

	//for off, reg := range regA {
	//VMOVDQU(Mem{Base: p}.Idx(i, 1).Offset(off*32), reg)
	//}

	for off, reg := range regB {
		VMOVDQU(Mem{Base: p}.Idx(j, 1).Offset(off*32), reg)
	}

	for off, reg := range regB {
		VMOVDQU(reg, Mem{Base: p}.Idx(i, 1).Offset(off*32))
	}

	ADDQ(U32(uint64(32*len(regB))), i)
}
