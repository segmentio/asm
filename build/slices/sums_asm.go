// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/segmentio/asm/build/internal/x86"

	"github.com/mmcloughlin/avo/reg"
	"github.com/segmentio/asm/cpu"
)

const unroll = 8

type Processor struct {
	xPtr    Mem
	yPtr    Mem
	len     reg.Register
	idx     reg.Register
	next    reg.Register
	vectors [unroll]reg.VecVirtual
}

func main() {
	generateSumUint64()
	generateSumUint32()
	generateSumUint16()
	generateSumUint8()
	Generate()
}

func generateSumUint64() {
	TEXT("sumUint64", NOSPLIT, "func(x, y []uint64)")
	Doc("Sum uint64s using avx2 instructions, results stored in x")
	p := genAvxTop(8, 2)

	// AVX intrinsics to sum 64 bit integers/quad words
	for offset, i := 0, 0; i < unroll/2; i++ {
		VPADDQ(p.vectors[offset], p.vectors[offset+1], p.vectors[offset])
		offset += 2
	}

	genAVXBottom(p)
	genX86Loop(p, func() {
		qword := GP64()
		MOVQ(p.yPtr, qword)
		ADDQ(qword, p.xPtr)
	})
}

func generateSumUint32() {
	TEXT("sumUint32", NOSPLIT, "func(x, y []uint32)")
	Doc("Sum uint32s using avx2 instructions, results stored in x")
	p := genAvxTop(4, 4)

	// AVX intrinsics to sum 32 bit integers/double words
	for offset, i := 0, 0; i < unroll/2; i++ {
		VPADDD(p.vectors[offset], p.vectors[offset+1], p.vectors[offset])
		offset += 2
	}

	genAVXBottom(p)
	genX86Loop(p, func() {
		dword := GP32()
		MOVL(p.yPtr, dword)
		ADDL(dword, p.xPtr)
	})
}

func generateSumUint16() {
	TEXT("sumUint16", NOSPLIT, "func(x, y []uint16)")
	Doc("Sum uint16s using avx2 instructions, results stored in x")
	p := genAvxTop(2, 8)

	// AVX intrinsics to sum 32 bit integers/double words
	for offset, i := 0, 0; i < unroll/2; i++ {
		VPADDW(p.vectors[offset], p.vectors[offset+1], p.vectors[offset])
		offset += 2
	}

	genAVXBottom(p)
	genX86Loop(p, func() {
		word := GP16()
		MOVW(p.yPtr, word)
		ADDW(word, p.xPtr)
	})
}

func generateSumUint8() {
	TEXT("sumUint8", NOSPLIT, "func(x, y []uint8)")
	Doc("Sum uint8s using avx2 instructions, results stored in x")
	p := genAvxTop(1, 16)

	// AVX intrinsics to sum 32 bit integers/double words
	for offset, i := 0, 0; i < unroll/2; i++ {
		VPADDB(p.vectors[offset], p.vectors[offset+1], p.vectors[offset])
		offset += 2
	}

	genAVXBottom(p)
	genX86Loop(p, func() {
		byter := GP8()
		MOVB(p.yPtr, byter)
		ADDB(byter, p.xPtr)
	})
}

func genAVXBottom(p *Processor) {
	for offset, i := 0, 0; i < unroll/2; i++ {
		VMOVDQU(p.vectors[offset], p.xPtr.Offset(i*32))
		offset += 2
	}
	// Increment ptrs and loop.
	MOVQ(p.next, p.idx)
	JMP(LabelRef("avx2_loop"))
}

func genX86Loop(p *Processor, calc func()) {
	// Here's we're just going to manually bump our pointers
	// and do a the addition on the remaining integers (if any)
	Label("x86_loop")
	CMPQ(p.idx, p.len)
	JAE(LabelRef("return"))

	// Delegate to specific computation
	calc()

	// Increment ptrs and loop.
	ADDQ(Imm(1), p.idx)
	JMP(LabelRef("x86_loop"))

	Label("return")
	RET()
}

func genAvxTop(scale uint8, avxOffset uint64) *Processor {
	// Little math here to calculate the memory address of our last value
	// 64bit uints so len * 8 from our original ptr address
	idx := GP64()
	XORQ(idx, idx)
	xPtr := Mem{Base: Load(Param("x").Base(), GP64()), Index: idx, Scale: scale}
	yPtr := Mem{Base: Load(Param("y").Base(), GP64()), Index: idx, Scale: scale}
	len := Load(Param("x").Len(), GP64())
	yLen := Load(Param("y").Len(), GP64())
	// len = min(len(x), len(y))
	CMPQ(yLen, len)
	CMOVQLT(yLen, len)

	JumpUnlessFeature("x86_loop", cpu.AVX2)

	Label("avx2_loop")
	next := GP64()
	MOVQ(idx, next)
	ADDQ(Imm(unroll*avxOffset), next)
	CMPQ(next, len)
	JAE(LabelRef("x86_loop"))

	// Create unroll num vector registers
	var vectors [unroll]reg.VecVirtual
	for i := 0; i < unroll; i++ {
		vectors[i] = YMM()
	}
	// So here essentially what we're doing is populating pairs
	// of vector registers with 256 bits of integer data, so as an example
	// for uint64s, it would look like...
	// YMM0 [ x0, x1, x2, x3 ]
	// YMM1 [ y0, y1, y2, y3 ]
	// ...
	// YMM(N) ...
	//
	// We then use VPADDQ to perform a SIMD addition operation
	// on the pairs and the result is stored in even register (0,2,4...).
	// Finally we copy the results back out to the slice pointed to by x
	for offset, i := 0, 0; i < unroll/2; i++ {
		VMOVDQU(xPtr.Offset(i*32), vectors[offset])
		VMOVDQU(yPtr.Offset(i*32), vectors[offset+1])
		offset += 2
	}

	return &Processor{
		xPtr:    xPtr,
		yPtr:    yPtr,
		len:     len,
		idx:     idx,
		next:    next,
		vectors: vectors,
	}
}
