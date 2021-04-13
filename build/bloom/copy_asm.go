// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("copyAVX2", NOSPLIT, "func(dst, src *byte, count int)")
	Doc("Copies the one-bits of src to dst, using SIMD instructions as an optimization.")

	dstPtr := Load(Param("dst"), GP64())
	srcPtr := Load(Param("src"), GP64())
	count := Load(Param("count"), GP64())

	// 256 bits registers
	dstReg0 := YMM()
	dstReg1 := YMM()

	Label("loop")
	Comment("Loop until zero bytes remain.")
	CMPQ(count, Imm(0))
	JE(LabelRef("done"))

	Comment("Load operands in registers, apply the OR operation, assign the result.")
	dstMem := Mem{Base: dstPtr}
	srcMem := Mem{Base: srcPtr}
	VMOVUPS(dstMem, dstReg0)
	VMOVUPS(dstMem.Offset(32), dstReg1)

	VPOR(srcMem, dstReg0, dstReg0)
	VPOR(srcMem.Offset(32), dstReg1, dstReg1)

	VMOVUPS(dstReg0, dstMem)
	VMOVUPS(dstReg1, dstMem.Offset(32))

	Comment("Decrement byte count, advance pointers.")
	DECQ(count)
	ADDQ(Imm(64), dstPtr)
	ADDQ(Imm(64), srcPtr)
	JMP(LabelRef("loop"))

	Label("done")
	RET()
	Generate()
}
