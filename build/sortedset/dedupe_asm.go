// +build ignore

package main

import (
	"fmt"
	"math/bits"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	//. "github.com/segmentio/asm/build/internal/x86"
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
	move(a, b Mem)
	test(a, b Mem)
}

type dedupeAVX2 interface {
	dedupe
	vpcmpeq(src0, src1, dst VecVirtual)
	vpmovmskb(tmp, src VecVirtual, dst Register)
}

type dedupe1 struct{}

func (dedupe1) size() int                             { return 1 }
func (dedupe1) test(a, b Mem)                         { generateDedupeTest(MOVB, CMPB, GP8, a, b) }
func (dedupe1) move(a, b Mem)                         { generateDedupeMove(MOVB, GP8, a, b) }
func (dedupe1) vpcmpeq(a, b, c VecVirtual)            { VPCMPEQB(a, b, c) }
func (dedupe1) vpmovmskb(_, a VecVirtual, b Register) { VPMOVMSKB(a, b) }

type dedupe2 struct{}

func (dedupe2) size() int                             { return 2 }
func (dedupe2) test(a, b Mem)                         { generateDedupeTest(MOVW, CMPW, GP16, a, b) }
func (dedupe2) move(a, b Mem)                         { generateDedupeMove(MOVW, GP16, a, b) }
func (dedupe2) vpcmpeq(a, b, c VecVirtual)            { VPCMPEQW(a, b, c) }
func (dedupe2) vpmovmskb(_, a VecVirtual, b Register) { VPMOVMSKB(a, b) }

type dedupe4 struct{}

func (dedupe4) size() int                             { return 4 }
func (dedupe4) test(a, b Mem)                         { generateDedupeTest(MOVL, CMPL, GP32, a, b) }
func (dedupe4) move(a, b Mem)                         { generateDedupeMove(MOVL, GP32, a, b) }
func (dedupe4) vpcmpeq(a, b, c VecVirtual)            { VPCMPEQD(a, b, c) }
func (dedupe4) vpmovmskb(_, a VecVirtual, b Register) { VPMOVMSKB(a, b) }

type dedupe8 struct{}

func (dedupe8) size() int                             { return 8 }
func (dedupe8) test(a, b Mem)                         { generateDedupeTest(MOVQ, CMPQ, GP64, a, b) }
func (dedupe8) move(a, b Mem)                         { generateDedupeMove(MOVQ, GP64, a, b) }
func (dedupe8) vpcmpeq(a, b, c VecVirtual)            { VPCMPEQQ(a, b, c) }
func (dedupe8) vpmovmskb(_, a VecVirtual, b Register) { VPMOVMSKB(a, b) }

type dedupe16 struct{}

func (dedupe16) size() int {
	return 16
}
func (dedupe16) test(a, b Mem) {
	r0, r1 := XMM(), XMM()
	MOVOU(a, r0)
	MOVOU(b, r1)
	mask := GP32()
	PCMPEQQ(r0, r1)
	PMOVMSKB(r1, mask)
	CMPL(mask, U32(0xFFFF))
}
func (dedupe16) move(a, b Mem) {
	r := XMM()
	MOVOU(a, r)
	MOVOU(r, b)
}
func (dedupe16) vpcmpeq(a, b, c VecVirtual) {
	VPCMPEQQ(a, b, c)
}
func (dedupe16) vpmovmskb(tmp, src VecVirtual, dst Register) {
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

type dedupe32 struct{}

func (dedupe32) size() int {
	return 32
}
func (dedupe32) test(a, b Mem) {
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
func (dedupe32) move(a, b Mem) {
	r0, r1 := XMM(), XMM()
	MOVOU(a, r0)
	MOVOU(a.Offset(16), r1)
	MOVOU(r0, b)
	MOVOU(r1, b.Offset(16))
}
func (dedupe32) vpcmpeq(a, b, c VecVirtual) {
	VPCMPEQQ(a, b, c)
}
func (dedupe32) vpmovmskb(_, src VecVirtual, dst Register) {
	VPMOVMSKB(src, dst)
}

func generateDedupe(code dedupe) {
	size := code.size()
	TEXT(fmt.Sprintf("dedupe%d", size), NOSPLIT, "func(b []byte) int")

	p := Load(Param("b").Base(), GP64())
	n := Load(Param("b").Len(), GP64())
	CMPQ(n, Imm(uint64(size)))
	JLE(LabelRef("none"))

	i := GP64()
	j := GP64()
	XORQ(i, i)
	MOVQ(U64(uint64(size)), j)

	if _, ok := code.(dedupeAVX2); ok {
		//JumpIfFeature("avx2", cpu.AVX2)
	}

	Label("tail")
	CMPQ(j, n)
	JGE(LabelRef("none"))

	Label("generic")
	k := GP64()
	MOVQ(i, k)
	ADDQ(Imm(uint64(size)), k)
	code.test(Mem{Base: p}.Idx(i, 1), Mem{Base: p}.Idx(j, 1))
	CMOVQNE(k, i)
	code.move(Mem{Base: p}.Idx(j, 1), Mem{Base: p}.Idx(i, 1))
	ADDQ(Imm(uint64(size)), j)
	CMPQ(j, n)
	JL(LabelRef("generic"))

	ADDQ(Imm(uint64(size)), i)
	Store(i, ReturnIndex(0))
	RET()

	Label("none")
	Store(n, ReturnIndex(0))
	RET()

	/*
		if avx, ok := code.(dedupeAVX2); ok {
			const avxChunk = 256
			const avxLanes = avxChunk / 32
			Label("avx2")
			CMPQ(n, U32(avxChunk+uint64(size)))
			JL(LabelRef(fmt.Sprintf("avx2_tail%d", avxChunk/2)))

			masks := make([]GPVirtual, avxLanes)
			for i := range masks {
				masks[i] = GP64()
				XORQ(masks[i], masks[i])
			}

			regA := make([]VecVirtual, avxLanes)
			regB := make([]VecVirtual, avxLanes)
			for i := range regA {
				regA[i] = YMM()
				regB[i] = YMM()
			}

			Label(fmt.Sprintf("avx2_loop%d", avxChunk))
			generateDedupeAVX2(r, p, regA, regB, masks, avx)
			ADDQ(U32(avxChunk), p)
			SUBQ(U32(avxChunk), n)
			CMPQ(n, U32(avxChunk+uint64(size)))
			JGE(LabelRef(fmt.Sprintf("avx2_loop%d", avxChunk)))

			for chunk := avxChunk / 2; chunk >= 32; chunk /= 2 {
				Label(fmt.Sprintf("avx2_tail%d", chunk))
				CMPQ(n, Imm(uint64(chunk+size)))
				JL(LabelRef(fmt.Sprintf("avx2_tail%d", chunk/2)))
				lanes := chunk / 32
				generateDedupeAVX2(r, p, regA[:lanes], regB[:lanes], masks[:lanes], avx)
				ADDQ(U32(uint64(chunk)), p)
				SUBQ(U32(uint64(chunk)), n)
			}

			Label("avx2_tail16")
			if size < 16 {
				CMPQ(n, Imm(uint64(16+size)))
				JL(LabelRef("avx2_tail"))
				generateDedupeAVX2(r, p, []VecVirtual{XMM()}, []VecVirtual{XMM()}, masks[:1], avx)
				ADDQ(Imm(16), p)
				SUBQ(Imm(16), n)
			}

			Label("avx2_tail")
			VZEROUPPER()
			if size < 32 {
				if shift := divideShift(size); shift > 0 {
					SHRQ(Imm(uint64(shift)), r)
				}
			}
			JMP(LabelRef("tail"))
		}
	*/
}

func generateDedupeTest(mov func(Op, Op), cmp func(Op, Op), reg func() GPVirtual, a, b Mem) {
	r0, r1 := reg(), reg()
	mov(a, r0)
	mov(b, r1)
	cmp(r0, r1)
}

func generateDedupeMove(mov func(Op, Op), reg func() GPVirtual, a, b Mem) {
	r := reg()
	mov(a, r)
	mov(r, b)
}

func generateDedupeAVX2(r, p Register, regA, regB []VecVirtual, masks []GPVirtual, code dedupeAVX2) {
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
			lo := moves[i*32+(size-16)]
			hi := moves[i*32+(size+16)]

			if lo != nil && hi != nil {
				// https://www.felixcloutier.com/x86/vperm2i128#vperm2i128
				//
				// The data was already loaded, but split across two registers.
				// We recompose it using a permutation of the upper and lower
				// halves of the registers holding the contiguous data.
				//
				// Note that in Go assembly the arguments are reversed;
				// SRC1 is `lo` and SRC2 is `hi`, but we pass them in the
				// reverse order.
				const permutation = (1 << 0) | (2 << 4)
				VPERM2I128(Imm(permutation), hi, lo, reg)
			} else {
				VMOVDQU((Mem{Base: p}).Offset(i*32+size), reg)
			}
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
		code.vpmovmskb(regA[i], regB[i], masks[i].As32())
	}

	for _, mask := range masks {
		POPCNTQ(mask, mask)
		if size == 32 {
			SHRQ(Imm(uint64(divideShift(size))), mask)
		}
	}

	ADDQ(divideAndConquerSum(masks), r)
}

func divideShift(size int) int {
	return bits.TrailingZeros(uint(size))
}

func divideAndConquerSum(regs []GPVirtual) GPVirtual {
	switch len(regs) {
	case 1:
		return regs[0]

	case 2:
		r0, r1 := regs[0], regs[1]
		ADDQ(r1, r0)
		return r0

	default:
		i := len(regs) / 2
		r0 := divideAndConquerSum(regs[:i])
		r1 := divideAndConquerSum(regs[i:])
		ADDQ(r1, r0)
		return r0
	}
}
