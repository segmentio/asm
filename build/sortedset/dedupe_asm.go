// +build ignore

package main

import (
	"fmt"
	"math/bits"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/x86"
	"github.com/segmentio/asm/cpu"
)

func main() {
	generateDedupe(new(dedupe1))
	generateDedupe(new(dedupe2))
	generateDedupe(new(dedupe4))
	generateDedupe(new(dedupe8))
	generateDedupe(new(dedupe16))
	generateDedupe(new(dedupe32))
	Generate()
}

type dedupe interface {
	size() int
	init(p, w GPVirtual)
	copy(p, q, w GPVirtual)
}

type dedupeAVX2 interface {
	dedupe
	vec() VecVirtual
	vsize() int
	vlanes() int
	vinit(p, w GPVirtual)
	vcopy(src, dst VecVirtual, off GPVirtual)
}

type dedupe1 struct{}

func (*dedupe1) size() int              { return 1 }
func (*dedupe1) init(p, w GPVirtual)    { move(MOVB, GP8(), p, w) }
func (*dedupe1) copy(p, q, w GPVirtual) { generateDedupeX86(MOVB, CMPB, GP8, p, q, w, 1) }

type dedupe2 struct{}

func (*dedupe2) size() int              { return 2 }
func (*dedupe2) init(p, w GPVirtual)    { move(MOVW, GP16(), p, w) }
func (*dedupe2) copy(p, q, w GPVirtual) { generateDedupeX86(MOVW, CMPW, GP16, p, q, w, 2) }

type dedupe4 struct{}

func (*dedupe4) size() int              { return 4 }
func (*dedupe4) init(p, w GPVirtual)    { move(MOVL, GP32(), p, w) }
func (*dedupe4) copy(p, q, w GPVirtual) { generateDedupeX86(MOVL, CMPL, GP32, p, q, w, 4) }

type dedupe8 struct{}

func (*dedupe8) size() int              { return 8 }
func (*dedupe8) init(p, w GPVirtual)    { move(MOVQ, GP64(), p, w) }
func (*dedupe8) copy(p, q, w GPVirtual) { generateDedupeX86(MOVQ, CMPQ, GP64, p, q, w, 8) }

type dedupe16 struct {
	nop GPVirtual
	inc GPVirtual
}

func (*dedupe16) size() int { return 16 }

func (*dedupe16) init(p, w GPVirtual) { move(MOVOU, XMM(), p, w) }

func (*dedupe16) copy(p, q, w GPVirtual) {
	next := GP64()
	MOVQ(w, next)
	ADDQ(Imm(16), next)
	xp, xq := XMM(), XMM()
	MOVOU(Mem{Base: p}, xp)
	MOVOU(Mem{Base: q}, xq)
	MOVOU(xq, Mem{Base: w})
	mask := GP32()
	PCMPEQQ(xp, xq)
	PMOVMSKB(xq, mask)
	CMPL(mask, U32(0xFFFF))
	CMOVQNE(next, w)
}

func (*dedupe16) vec() VecVirtual { return XMM() }

func (*dedupe16) vsize() int { return 16 }

func (*dedupe16) vlanes() int { return 8 }

func (d *dedupe16) vinit(p, w GPVirtual) {
	move(VMOVDQU, XMM(), p, w)
	d.nop = GP64()
	d.inc = GP64()
	XORQ(d.nop, d.nop)
	MOVQ(U64(16), d.inc)
}

func (d *dedupe16) vcopy(src, dst VecVirtual, off GPVirtual) {
	VPCMPEQQ(dst, src, src)
	// This gives a bitmask with these possible values:
	// * 0b00
	// * 0b01
	// * 0b10
	// * 0b11
	// We only care about the last case, which indicates that both 64 bits lanes
	// of the XMM register were equal.
	VMOVMSKPD(src, off.As32())
	CMPQ(off, Imm(3))
	CMOVQEQ(d.nop, off)
	CMOVQNE(d.inc, off)
}

type dedupe32 struct {
	nop GPVirtual
	inc GPVirtual
}

func (*dedupe32) size() int { return 32 }

func (*dedupe32) init(p, w GPVirtual) {
	lo, hi := XMM(), XMM()
	MOVOU(Mem{Base: p}, lo)
	MOVOU(Mem{Base: p}.Offset(16), hi)
	MOVOU(lo, Mem{Base: w})
	MOVOU(hi, Mem{Base: w}.Offset(16))
}

func (*dedupe32) copy(p, q, w GPVirtual) {
	next := GP64()
	MOVQ(w, next)
	ADDQ(Imm(32), next)
	loP, hiP := XMM(), XMM()
	loQ, hiQ := XMM(), XMM()
	MOVOU(Mem{Base: q}, loQ)
	MOVOU(Mem{Base: q}.Offset(16), hiQ)
	MOVOU(Mem{Base: p}, loP)
	MOVOU(Mem{Base: p}.Offset(16), hiP)
	MOVOU(loQ, Mem{Base: w})
	MOVOU(hiQ, Mem{Base: w}.Offset(16))
	mask0, mask1 := GP32(), GP32()
	PCMPEQQ(loP, loQ)
	PCMPEQQ(hiP, hiQ)
	PMOVMSKB(loQ, mask0)
	PMOVMSKB(hiQ, mask1)
	ANDL(mask1, mask0)
	CMPL(mask0, U32(0xFFFF))
	CMOVQNE(next, w)
}

func (*dedupe32) vec() VecVirtual { return YMM() }

func (*dedupe32) vsize() int { return 32 }

func (*dedupe32) vlanes() int { return 8 }

func (d *dedupe32) vinit(p, w GPVirtual) {
	move(VMOVDQU, YMM(), p, w)
	d.nop = GP64()
	d.inc = GP64()
	XORQ(d.nop, d.nop)
	MOVQ(U64(32), d.inc)
}

func (d *dedupe32) vcopy(src, dst VecVirtual, off GPVirtual) {
	VPCMPEQQ(dst, src, src)
	// This gives a bitmask with these possible values:
	// * 0b0000
	// * 0b0001
	// * ...
	// * 0b1111
	//
	// We only care about the last case because it indicates that the full 32
	// bytes are equal.
	//
	// We want to divide by 15, which will either produce a result of 0 or 1.
	// Rather than dividing, we add 1 and shift right by 4.
	VMOVMSKPD(src, off.As32())
	CMPQ(off, Imm(15))
	CMOVQEQ(d.nop, off)
	CMOVQNE(d.inc, off)
}

func generateDedupe(dedupe dedupe) {
	size := dedupe.size()
	TEXT(fmt.Sprintf("dedupe%d", size), NOSPLIT, "func(dst, src []byte) int")

	n := Load(Param("src").Len(), GP64())
	CMPQ(n, Imm(0))
	JE(LabelRef("short"))

	dst := Load(Param("dst").Base(), GP64())
	src := Load(Param("src").Base(), GP64())
	// `p` and `q` are two read pointers that will be advanced through the
	// input array testing for equal pairs.
	//
	// `w` points to the position in the output buffer where the next item
	// is to be written.
	p := GP64()
	q := GP64()
	w := GP64()
	MOVQ(src, p)
	MOVQ(src, q)
	MOVQ(dst, w)
	ADDQ(Imm(uint64(size)), q)
	SUBQ(Imm(uint64(size)), n)

	if _, ok := dedupe.(dedupeAVX2); ok {
		CMPQ(n, Imm(32))
		JL(LabelRef("init"))
		JumpIfFeature("avx2", cpu.AVX2)
	}

	Label("init")
	dedupe.init(p, w)
	ADDQ(Imm(uint64(size)), w)

	Label("tail")
	CMPQ(n, Imm(0))
	JE(LabelRef("done"))

	Label("generic")
	dedupe.copy(p, q, w)
	ADDQ(Imm(uint64(size)), p)
	ADDQ(Imm(uint64(size)), q)
	SUBQ(Imm(uint64(size)), n)
	CMPQ(n, Imm(0))
	JG(LabelRef("generic"))

	Label("done")
	SUBQ(dst, w)
	Store(w, ReturnIndex(0))
	RET()

	Label("short")
	Store(n, ReturnIndex(0))
	RET()

	if avx, ok := dedupe.(dedupeAVX2); ok {
		avxLanes := avx.vlanes()
		avxChunk := avx.vsize() * avxLanes
		Label("avx2")

		src := make([]VecVirtual, avxLanes)
		dst := make([]VecVirtual, avxLanes)
		off := make([]GPVirtual, avxLanes)
		for i := range src {
			src[i] = avx.vec()
			dst[i] = avx.vec()
			off[i] = GP64()
		}

		avx.vinit(p, w)
		ADDQ(Imm(uint64(size)), w)

		// This bit of magic aligns the tail chunk size on the first power of
		// two smaller than the chunk size used in the loop.
		//
		// This is useful when the number of lanes in not a power of two.
		tailChunk := 1 << (63 - bits.LeadingZeros(uint(avxChunk)))
		if tailChunk == avxChunk {
			tailChunk /= 2
		}

		CMPQ(n, U32(avxChunk))
		if tailChunk >= avx.size() {
			JL(LabelRef(fmt.Sprintf("avx2_tail%d", tailChunk)))
		} else {
			JL(LabelRef("avx2_tail"))
		}

		Label(fmt.Sprintf("avx2_loop%d", avxChunk))
		generateDedupeAVX2(p, q, w, src, dst, off, avx)
		ADDQ(U32(uint64(avxChunk)), p)
		ADDQ(U32(uint64(avxChunk)), q)
		SUBQ(U32(uint64(avxChunk)), n)
		CMPQ(n, U32(avxChunk))
		JGE(LabelRef(fmt.Sprintf("avx2_loop%d", avxChunk)))

		for chunk := tailChunk; chunk >= avx.vsize(); chunk /= 2 {
			Label(fmt.Sprintf("avx2_tail%d", chunk))
			CMPQ(n, Imm(uint64(chunk)))
			if next := chunk / 2; next >= avx.vsize() {
				JL(LabelRef(fmt.Sprintf("avx2_tail%d", chunk/2)))
			} else {
				JL(LabelRef("avx2_tail"))
			}
			lanes := chunk / avx.vsize()
			generateDedupeAVX2(p, q, w, src[:lanes], dst[:lanes], off[:lanes], avx)
			ADDQ(Imm(uint64(chunk)), p)
			ADDQ(Imm(uint64(chunk)), q)
			SUBQ(Imm(uint64(chunk)), n)
		}

		Label("avx2_tail")
		VZEROUPPER()
		JMP(LabelRef("tail"))
	}
}

func generateDedupeX86(mov func(Op, Op), cmp func(Op, Op), reg func() GPVirtual, p, q, w GPVirtual, size int) {
	next := GP64()
	MOVQ(w, next)
	ADDQ(Imm(uint64(size)), next)
	pv, qv := reg(), reg()
	mov(Mem{Base: p}, pv)
	mov(Mem{Base: q}, qv)
	mov(qv, Mem{Base: w})
	cmp(pv, qv)
	CMOVQNE(next, w)
}

func generateDedupeAVX2(p, q, w GPVirtual, src, dst []VecVirtual, off []GPVirtual, dedupe dedupeAVX2) {
	size := dedupe.vsize()
	for i := range src {
		VMOVDQU(Mem{Base: p}.Offset(i*size), src[i])
	}
	for i := range dst {
		VMOVDQU(Mem{Base: q}.Offset(i*size), dst[i])
	}
	for i := range src {
		dedupe.vcopy(src[i], dst[i], off[i])
	}

	// Compute the cumulative offsets so we can use indexes relative to the
	// write pointer, which allows the CPU to pipeline the writes to memory.
	//
	// There are still strong data dependencies between these instructions,
	// but I'm not sure there is a great alternative. Moving the values to a
	// vector register and using SIMD seems like a lost of heavy lifting for
	// the limited number of registers we have.
	for i := range off[1:] {
		ADDQ(off[i], off[i+1])
	}

	for i := range dst {
		if i == 0 {
			VMOVDQU(dst[i], Mem{Base: w})
		} else {
			VMOVDQU(dst[i], Mem{Base: w}.Idx(off[i-1], 1))
		}
	}

	ADDQ(off[len(off)-1], w)
}

func move(mov func(Op, Op), tmp Register, src, dst GPVirtual) {
	mov(Mem{Base: src}, tmp)
	mov(tmp, Mem{Base: dst})
}
