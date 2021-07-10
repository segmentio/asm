// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

const unroll = 8

func main() {
	TEXT("sumUint64", NOSPLIT, "func(x, y []uint64)")
	Doc("Sum uint64s using avx2 instructions, results stored in x")
	xPtr := Load(Param("x").Base(), GP64())
	yPtr := Load(Param("y").Base(), GP64())
	xLen := Load(Param("x").Len(), GP64())
	yLen := Load(Param("y").Len(), GP64())
	xEnd := GP64()
	yEnd := GP64()
	MOVQ(xPtr, xEnd)
	MOVQ(yPtr, yEnd)

	// Little math here to calculate the memory address of our last value
	// 64bit uints so lex * 8 from our original ptr address
	z := GP64()
	MOVQ(U64(8), z)
	MOVQ(xLen, reg.RAX)
	MULQ(z)
	ADDQ(reg.RAX, xEnd)

	MOVQ(yLen, reg.RAX)
	MULQ(z)
	ADDQ(reg.RAX, yEnd)

	// Hitting an error..
	// github.com/segmentio/asm/zip.SumUint64: relocation target github.com/segmentio/asm/cpu.X86 not defined
	// Once resolved we can add this AVX check
	//JumpUnlessFeature("x86_loop", cpu.AVX2)

	Label("avx2_loop")
	xNext := GP64()
	yNext := GP64()
	MOVQ(xPtr, xNext)
	MOVQ(yPtr, yNext)
	ADDQ(Imm(unroll*16), xNext)
	ADDQ(Imm(unroll*16), yNext)
	CMPQ(xNext, xEnd)
	JAE(LabelRef("x86_loop"))
	CMPQ(yNext, yEnd)
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
		VMOVDQU(Mem{Base: xPtr}.Offset(i*32), vectors[offset])
		VMOVDQU(Mem{Base: yPtr}.Offset(i*32), vectors[offset+1])
		offset += 2
	}

	for offset, i := 0, 0; i < unroll/2; i++ {
		VPADDQ(vectors[offset], vectors[offset+1], vectors[offset])
		offset += 2
	}

	for offset, i := 0, 0; i < unroll/2; i++ {
		VMOVDQU(vectors[offset], Mem{Base: xPtr}.Offset(i*32))
		offset += 2
	}
	// Increment ptrs and loop.
	MOVQ(xNext, xPtr)
	MOVQ(yNext, yPtr)
	JMP(LabelRef("avx2_loop"))

	// Here's we're just going to manually bump our pointers
	// and do a the addition on the remaining integers (if any)
	Label("x86_loop")
	CMPQ(xPtr, xEnd)
	JAE(LabelRef("return"))
	CMPQ(yPtr, yEnd)
	JAE(LabelRef("return"))

	qword := GP64()
	MOVQ(Mem{Base: xPtr}, qword)
	ADDQ(Mem{Base: yPtr}, qword)
	MOVQ(qword, Mem{Base: xPtr})
	// Increment ptrs and loop.
	ADDQ(Imm(8), xPtr)
	ADDQ(Imm(8), yPtr)
	JMP(LabelRef("x86_loop"))

	Label("return")
	RET()
	Generate()
}
