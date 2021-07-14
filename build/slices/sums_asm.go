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

func main() {
	TEXT("sumUint64", NOSPLIT, "func(x, y []uint64)")
	Doc("Sum uint64s using avx2 instructions, results stored in x")
	// Little math here to calculate the memory address of our last value
	// 64bit uints so len * 8 from our original ptr address
	idx := GP64()
	XORQ(idx, idx)
	xPtr := Mem{Base: Load(Param("x").Base(), GP64()), Index: idx, Scale: 8}
	yPtr := Mem{Base: Load(Param("y").Base(), GP64()), Index: idx, Scale: 8}
	len := Load(Param("x").Len(), GP64())
	yLen := Load(Param("y").Len(), GP64())
	// len = min(len(x), len(y))
	CMPQ(yLen, len)
	CMOVQLT(yLen, len)

	JumpUnlessFeature("x86_loop", cpu.AVX2)

	Label("avx2_loop")
	next := GP64()
	MOVQ(idx, next)
	ADDQ(Imm(unroll*2), next)
	CMPQ(next, len)
	JAE(LabelRef("x86_loop"))

	// Create unroll num vector registers
	var vectors [unroll]reg.VecVirtual
	for i := 0; i < unroll; i++ {
		vectors[i] = YMM()
	}

	// So here essentially what we're doing is populating pairs
	// of vector registers with 4, 64 bit uints, like...
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

	for offset, i := 0, 0; i < unroll/2; i++ {
		VPADDQ(vectors[offset], vectors[offset+1], vectors[offset])
		offset += 2
	}

	for offset, i := 0, 0; i < unroll/2; i++ {
		VMOVDQU(vectors[offset], xPtr.Offset(i*32))
		offset += 2
	}
	// Increment ptrs and loop.
	MOVQ(next, idx)
	JMP(LabelRef("avx2_loop"))

	// Here's we're just going to manually bump our pointers
	// and do a the addition on the remaining integers (if any)
	Label("x86_loop")
	CMPQ(idx, len)
	JAE(LabelRef("return"))

	qword := GP64()
	MOVQ(yPtr, qword)
	ADDQ(qword, xPtr)
	// Increment ptrs and loop.
	ADDQ(Imm(1), idx)
	JMP(LabelRef("x86_loop"))

	Label("return")
	RET()
	Generate()
}
