// +build ignore

package main

import (
	"fmt"
	"math/bits"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/asm"
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
	vmask() Mem
	vcopy(shuff Mem, tmp, src, dst VecVirtual, off GPVirtual)
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

func (dedupe32) vmask() Mem {
	return ConstBytes("dedupe32_blend_mask", []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,

		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	})
}

func (dedupe32) vcopy(shuff Mem, tmp, src, dst VecVirtual, off GPVirtual) {
	VPCMPEQQ(src, dst, tmp)
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
	VMOVMSKPD(tmp, off.As32())
	INCQ(off)
	SHRQ(Imm(4), off)
	// The off register now has the value 0 or 1, the former indicates that
	// items were not equal (advance), the latter that they were. We flip the
	// bit so 1 indicate that we want to eventually increment the write pointer.
	NOTQ(off)
	ANDQ(Imm(1), off)
	// There are two 32 bytes blend masks collocated in the global variable,
	// the offset is off*32, so we shift left by 5 to determine the offset,
	// which also turns out to be the number of bytes that we copy so off
	// contains the number of bytes to increment by.
	SHLQ(Imm(5), off)
	// Pick the part of the mask that corresponds to either a copy or a no-op,
	// depending on whether the source and destination would be the same or not.
	VMOVDQU(shuff.Idx(off, 1), tmp)
	VBLENDVPD(tmp, src, dst, dst)
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

	if _, ok := dedupe.(dedupeAVX2); ok {
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

	if avx, ok := dedupe.(dedupeAVX2); ok {
		const avxChunk = 32
		const avxLanes = avxChunk / 32
		Label("avx2")

		off := make([]GPVirtual, avxLanes)
		for i := range off {
			off[i] = GP64()
			XORQ(off[i], off[i])
		}

		src := make([]VecVirtual, avxLanes)
		dst := make([]VecVirtual, avxLanes)
		tmp := make([]VecVirtual, avxLanes)
		for i := range src {
			src[i] = YMM()
			dst[i] = YMM()
			tmp[i] = YMM()
		}

		shuff := GP64()
		LEAQ(avx.vmask(), shuff)

		CMPQ(n, U32(avxChunk))
		JL(LabelRef(fmt.Sprintf("avx2_tail%d", avxChunk/2)))

		Label(fmt.Sprintf("avx2_loop%d", avxChunk))
		generateDedupeAVX2(r, w, shuff, tmp, src, dst, off, avx)
		ADDQ(U32(uint64(avxChunk)), r)
		SUBQ(U32(uint64(avxChunk)), n)
		CMPQ(n, U32(avxChunk))
		JGE(LabelRef(fmt.Sprintf("avx2_loop%d", avxChunk)))

		for chunk := avxChunk / 2; chunk >= 32; chunk /= 2 {
			Label(fmt.Sprintf("avx2_tail%d", chunk))
			CMPQ(n, Imm(uint64(chunk)))
			JL(LabelRef(fmt.Sprintf("avx2_tail%d", chunk/2)))
			lanes := chunk / 32
			generateDedupeAVX2(r, w, shuff, tmp[:lanes], src[:lanes], dst[:lanes], off[:lanes], avx)
			ADDQ(Imm(uint64(chunk)), r)
			SUBQ(Imm(uint64(chunk)), n)
		}

		Label("avx2_tail16")
		if size < 16 {
			CMPQ(n, Imm(uint64(16+size)))
			JL(LabelRef("avx2_tail"))
			tmp := []VecVirtual{XMM()}
			src := []VecVirtual{XMM()}
			dst := []VecVirtual{XMM()}
			generateDedupeAVX2(r, w, shuff, tmp, src, dst, off[:1], avx)
			ADDQ(Imm(16), r)
			SUBQ(Imm(16), n)
		}

		Label("avx2_tail")
		VZEROUPPER()
		JMP(LabelRef("tail"))
	}
}

func generateDedupeX86(mov func(Op, Op), cmp func(Op, Op), reg func() GPVirtual, r, w, x GPVirtual) {
	tmp := reg()
	mov(Mem{Base: r}, tmp)
	cmp(tmp, Mem{Base: w})
	CMOVQNE(x, w)
	mov(tmp, Mem{Base: w})
}

func generateDedupeAVX2(r, w, shuff GPVirtual, tmp, src, dst []VecVirtual, off []GPVirtual, dedupe dedupeAVX2) {
	for i, reg := range src {
		VMOVDQU(Mem{Base: r}.Offset(i*32), reg)
	}
	for i, reg := range dst {
		VMOVDQU(Mem{Base: w}.Offset(i*32), reg)
	}
	for i := range src {
		dedupe.vcopy(Mem{Base: shuff}, tmp[i], src[i], dst[i], off[i])
	}
	for i := range dst {
		VMOVDQU(dst[i], Mem{Base: w}.Idx(off[i], 1))
		ADDQ(off[i], w)
	}
}

func shift(size int) int {
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
