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
	generateIndexPair(indexPair32{})
	Generate()
}

type indexPair interface {
	size() int
	test(a, b Mem)
}

type indexPairAVX interface {
	indexPair
	vpcmpeq(src0, src1, dst VecVirtual)
	vpmovmskb(tmp, src VecVirtual, zero, dst Register)
}

type indexPair1 struct{}

func (indexPair1) size() int                                { return 1 }
func (indexPair1) test(a, b Mem)                            { generateIndexPairTest(MOVB, CMPB, GP8, a, b) }
func (indexPair1) vpcmpeq(a, b, c VecVirtual)               { VPCMPEQB(a, b, c) }
func (indexPair1) vpmovmskb(_, a VecVirtual, _, b Register) { VPMOVMSKB(a, b) }

type indexPair2 struct{}

func (indexPair2) size() int                                { return 2 }
func (indexPair2) test(a, b Mem)                            { generateIndexPairTest(MOVW, CMPW, GP16, a, b) }
func (indexPair2) vpcmpeq(a, b, c VecVirtual)               { VPCMPEQW(a, b, c) }
func (indexPair2) vpmovmskb(_, a VecVirtual, _, b Register) { VPMOVMSKB(a, b) }

type indexPair4 struct{}

func (indexPair4) size() int                                { return 4 }
func (indexPair4) test(a, b Mem)                            { generateIndexPairTest(MOVL, CMPL, GP32, a, b) }
func (indexPair4) vpcmpeq(a, b, c VecVirtual)               { VPCMPEQD(a, b, c) }
func (indexPair4) vpmovmskb(_, a VecVirtual, _, b Register) { VPMOVMSKB(a, b) }

type indexPair8 struct{}

func (indexPair8) size() int                                { return 8 }
func (indexPair8) test(a, b Mem)                            { generateIndexPairTest(MOVQ, CMPQ, GP64, a, b) }
func (indexPair8) vpcmpeq(a, b, c VecVirtual)               { VPCMPEQQ(a, b, c) }
func (indexPair8) vpmovmskb(_, a VecVirtual, _, b Register) { VPMOVMSKB(a, b) }

type indexPair16 struct{}

func (indexPair16) size() int {
	return 16
}
func (indexPair16) test(a, b Mem) {
	r0, r1 := XMM(), XMM()
	MOVOU(a, r0)
	MOVOU(b, r1)
	mask := GP32()
	PCMPEQQ(r0, r1)
	PMOVMSKB(r1, mask)
	CMPL(mask, U32(0xFFFF))
}
func (indexPair16) vpcmpeq(a, b, c VecVirtual) {
	VPCMPEQQ(a, b, c)
}
func (indexPair16) vpmovmskb(tmp, src VecVirtual, _, dst Register) {
	// https://www.felixcloutier.com/x86/vpermq#vpermq--vex-256-encoded-version-
	//
	// Swap each quad word in the lower and upper half of the 32 bytes register,
	// then AND the src and tmp registers to zero each halves that were partial
	// equality; only fully equal 128 bits need to result in setting 1 bits in
	// the destination mask.
	const permutation = (1 << 0) | (0 << 2) | (3 << 4) | (2 << 6)
	VPERMQ(Imm(permutation), src, tmp)
	VPAND(src, tmp, tmp)
	VPMOVMSKB(tmp, dst)
}

type indexPair32 struct{}

func (indexPair32) size() int {
	return 32
}
func (indexPair32) test(a, b Mem) {
	r0, r1, r2, r3 := XMM(), XMM(), XMM(), XMM()
	MOVOU(a, r0)
	MOVOU(a.Offset(16), r1)
	MOVOU(b, r2)
	MOVOU(b.Offset(16), r3)
	mask0, mask1 := GP32(), GP32()
	PCMPEQQ(r0, r2)
	PCMPEQQ(r1, r3)
	PMOVMSKB(r2, mask0)
	PMOVMSKB(r3, mask1)
	ANDL(mask1, mask0)
	CMPL(mask0, U32(0xFFFF))
}
func (indexPair32) vpcmpeq(a, b, c VecVirtual) {
	VPCMPEQQ(a, b, c)
}
func (indexPair32) vpmovmskb(_, src VecVirtual, zero, dst Register) {
	VPMOVMSKB(src, dst)
	CMPL(dst, U32(0xFFFFFFFF))
	CMOVLNE(zero, dst)
}

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

	if _, ok := code.(indexPairAVX); ok {
		JumpIfFeature("avx2", cpu.AVX2)
	}

	Label("tail")
	CMPQ(n, Imm(0))
	JE(LabelRef("fail"))

	Label("generic")
	code.test(Mem{Base: p}, (Mem{Base: p}).Offset(size))
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

	if avx, ok := code.(indexPairAVX); ok {
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
		zero := GP64()
		MOVQ(U64(0), zero)

		regA := make([]VecVirtual, avxLanes)
		regB := make([]VecVirtual, avxLanes)
		for i := range regA {
			regA[i] = YMM()
			regB[i] = YMM()
		}

		Label(fmt.Sprintf("avx2_loop%d", avxChunk))
		generateIndexPairAVX2(p, regA, regB, masks, zero, avx)
		ADDQ(U32(avxChunk), p)
		SUBQ(U32(avxChunk), n)
		CMPQ(n, U32(avxChunk+uint64(size)))
		JAE(LabelRef(fmt.Sprintf("avx2_loop%d", avxChunk)))

		for chunk := avxChunk / 2; chunk >= 32; chunk /= 2 {
			Label(fmt.Sprintf("avx2_tail%d", chunk))
			CMPQ(n, Imm(uint64(chunk+size)))
			JB(LabelRef(fmt.Sprintf("avx2_tail%d", chunk/2)))
			lanes := chunk / 32
			generateIndexPairAVX2(p, regA[:lanes], regB[:lanes], masks[:lanes], zero, avx)
			ADDQ(U32(uint64(chunk)), p)
			SUBQ(U32(uint64(chunk)), n)
		}

		Label("avx2_tail16")
		if size < 16 {
			CMPQ(n, Imm(uint64(16+size)))
			JB(LabelRef("avx2_tail"))
			generateIndexPairAVX2(p, []VecVirtual{XMM()}, []VecVirtual{XMM()}, masks[:1], zero, avx)
			ADDQ(Imm(16), p)
			SUBQ(Imm(16), n)
		}

		Label("avx2_tail")
		VZEROUPPER()
		JMP(LabelRef("tail"))

		Label("avx2_done")
		VZEROUPPER()
		for i, mask := range masks {
			CMPQ(mask, Imm(0))
			JNE(LabelRef(fmt.Sprintf("avx2_done%d", i)))
		}

		for i, mask := range masks {
			Label(fmt.Sprintf("avx2_done%d", i))
			if i > 0 {
				ADDQ(U32(uint64(i*32)), p)
				SUBQ(U32(uint64(i*32)), n)
			}
			TZCNTQ(mask, mask)
			ADDQ(mask, p)
			SUBQ(mask, n)
			JMP(LabelRef("done"))
		}
	}
}

func generateIndexPairTest(mov func(Op, Op), cmp func(Op, Op), reg func() GPVirtual, a, b Mem) {
	r0, r1 := reg(), reg()
	mov(a, r0)
	mov(b, r1)
	cmp(r0, r1)
}

func generateIndexPairAVX2(p Register, regA, regB []VecVirtual, masks []GPVirtual, zero GPVirtual, code indexPairAVX) {
	size := code.size()
	moves := make(map[int]VecVirtual)

	for i, reg := range regA {
		VMOVDQU((Mem{Base: p}).Offset(i*32), reg)
		moves[i*32] = reg
	}

	for i, reg := range regB {
		// Skip loading from memory a second time if we already loaded the
		// offset in the previous loop. This optimization applies for items
		// of size 32.
		if moves[i*32+size] == nil {
			VMOVDQU((Mem{Base: p}).Offset(i*32+size), reg)
		}
	}

	for i := range regA {
		// The load may have been elided if there was offset overlaps between
		// the two sources.
		if mov := moves[i*32+size]; mov != nil {
			code.vpcmpeq(regA[i], mov, regB[i])
		} else {
			code.vpcmpeq(regA[i], regB[i], regB[i])
		}
	}

	for i := range regB {
		code.vpmovmskb(regA[i], regB[i], zero.As32(), masks[i].As32())
	}

	combinedMask := GP64()
	if len(masks) == 1 {
		combinedMask = masks[0]
	} else {
		MOVQ(U64(0), combinedMask)
		for _, mask := range masks {
			ORQ(mask, combinedMask)
		}
	}

	CMPQ(combinedMask, Imm(0))
	JNE(LabelRef("avx2_done"))
}
