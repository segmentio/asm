// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

const unroll = 4

var avx2ShuffleMask = []byte{
	7, 6, 5, 4, 3, 2, 1, 0,
	15, 14, 13, 12, 11, 10, 9, 8,
	7, 6, 5, 4, 3, 2, 1, 0,
	15, 14, 13, 12, 11, 10, 9, 8,
}

func main() {
	TEXT("bswapq", NOSPLIT, "func(b []byte)")
	Doc("bswapq performs an in-place byte swap on each qword of the input buffer.")

	// Load slice ptr + length, and calculate end ptr.
	ptr := Load(Param("b").Base(), GP64())
	len := Load(Param("b").Len(), GP64())
	end := GP64()
	MOVQ(ptr, end)
	ADDQ(len, end)

	// TODO: jump to x86_loop if there's no AVX2 support

	// Prepare the shuffle mask.
	shuffleMaskData := GLOBL("shuffle_mask", RODATA|NOPTR)
	for i, b := range avx2ShuffleMask {
		DATA(i, U8(b))
	}
	shuffleMask := YMM()
	VMOVDQU(shuffleMaskData, shuffleMask)

	// Loop while we have at least unroll*32 bytes remaining.
	Label("avx2_loop")
	next := GP64()
	MOVQ(ptr, next)
	ADDQ(Imm(unroll*32), next)
	CMPQ(next, end)
	JAE(LabelRef("x86_loop"))

	// Load multiple chunks => byte swap => store.
	var vectors [unroll]reg.VecVirtual
	for i := 0; i < unroll; i++ {
		vectors[i] = YMM()
	}
	for i := 0; i < unroll; i++ {
		VMOVDQU(Mem{Base: ptr}.Offset(i*32), vectors[i])
	}
	for i := 0; i < unroll; i++ {
		VPSHUFB(shuffleMask, vectors[i], vectors[i])
	}
	for i := 0; i < unroll; i++ {
		VMOVDQU(vectors[i], Mem{Base: ptr}.Offset(i*32))
	}

	// Increment ptr and loop.
	MOVQ(next, ptr)
	JMP(LabelRef("avx2_loop"))

	// Loop while we have at least unroll*8 bytes remaining.
	Label("x86_loop")
	next = GP64()
	MOVQ(ptr, next)
	ADDQ(Imm(unroll*8), next)
	CMPQ(next, end)
	JAE(LabelRef("slow_loop"))

	// Load qwords => byte swap => store.
	var chunks [unroll]reg.GPVirtual
	for i := 0; i < unroll; i++ {
		chunks[i] = GP64()
	}
	for i := 0; i < unroll; i++ {
		MOVQ(Mem{Base: ptr}.Offset(i*8), chunks[i])
	}
	for i := 0; i < unroll; i++ {
		BSWAPQ(chunks[i])
	}
	for i := 0; i < unroll; i++ {
		MOVQ(chunks[i], Mem{Base: ptr}.Offset(i*8))
	}

	// Increment ptr and loop.
	MOVQ(next, ptr)
	JMP(LabelRef("x86_loop"))

	// Loop until ptr reaches the end.
	Label("slow_loop")
	CMPQ(ptr, end)
	JAE(LabelRef("done"))

	// Load a qword => byte swap => store.
	qword := GP64()
	MOVQ(Mem{Base: ptr}, qword)
	BSWAPQ(qword)
	MOVQ(qword, Mem{Base: ptr})

	// Increment ptr and loop.
	ADDQ(Imm(8), ptr)
	JMP(LabelRef("slow_loop"))

	Label("done")
	RET()
	Generate()
}
