// +build ignore

package main

import (
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/segmentio/asm/build/internal/x86"

	"github.com/mmcloughlin/avo/reg"
	"github.com/segmentio/asm/cpu"
)

const unroll = 8

type Processor struct {
	name      string
	typ       string
	scale     uint8
	avxOffset uint64
	avxAdd    func(...Op)
	x86Mov    func(imr, mr Op)
	x86Add    func(imr, amr Op)
	x86Reg    reg.GPVirtual
}

func init() {
	ConstraintExpr("!purego")
}

func main() {
	generate(Processor{
		name:      "sumUint64",
		typ:       "uint64",
		scale:     8,
		avxOffset: 2,
		avxAdd:    VPADDQ,
		x86Mov:    MOVQ,
		x86Add:    ADDQ,
		x86Reg:    GP64(),
	})

	generate(Processor{
		name:      "sumUint32",
		typ:       "uint32",
		scale:     4,
		avxOffset: 4,
		avxAdd:    VPADDD,
		x86Mov:    MOVL,
		x86Add:    ADDL,
		x86Reg:    GP32(),
	})

	generate(Processor{
		name:      "sumUint16",
		typ:       "uint16",
		scale:     2,
		avxOffset: 8,
		avxAdd:    VPADDW,
		x86Mov:    MOVW,
		x86Add:    ADDW,
		x86Reg:    GP16(),
	})

	generate(Processor{
		name:      "sumUint8",
		typ:       "uint8",
		scale:     1,
		avxOffset: 16,
		avxAdd:    VPADDB,
		x86Mov:    MOVB,
		x86Add:    ADDB,
		x86Reg:    GP8(),
	})

	Generate()
}

func generate(p Processor) {
	TEXT(p.name, NOSPLIT, fmt.Sprintf("func(x, y []%s)", p.typ))
	Doc(fmt.Sprintf("Sum %ss using avx2 instructions, results stored in x", p.typ))
	idx := GP64()
	XORQ(idx, idx)
	xPtr := Mem{Base: Load(Param("x").Base(), GP64()), Index: idx, Scale: p.scale}
	yPtr := Mem{Base: Load(Param("y").Base(), GP64()), Index: idx, Scale: p.scale}
	len := Load(Param("x").Len(), GP64())
	yLen := Load(Param("y").Len(), GP64())
	// len = min(len(x), len(y))
	CMPQ(yLen, len)
	CMOVQLT(yLen, len)

	JumpUnlessFeature("x86_loop", cpu.AVX2)

	Label("avx2_loop")
	next := GP64()
	MOVQ(idx, next)
	ADDQ(Imm(unroll*p.avxOffset), next)
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

	// AVX intrinsics to sum 64 bit integers/quad words
	for offset, i := 0, 0; i < unroll/2; i++ {
		p.avxAdd(vectors[offset], vectors[offset+1], vectors[offset])
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

	// Delegate to specific computation
	//calc()
	p.x86Mov(yPtr, p.x86Reg)
	p.x86Add(p.x86Reg, xPtr)

	// Increment ptrs and loop.
	ADDQ(Imm(1), idx)
	JMP(LabelRef("x86_loop"))

	Label("return")
	RET()
}
