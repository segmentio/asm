// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("copyAVX2", NOSPLIT, "func(dst, src *byte, n int)")
	Doc("Copies the one-bits of src to dst, using SIMD instructions as an optimization.")

	dst := Load(Param("dst"), GP64())
	src := Load(Param("src"), GP64())
	n := Load(Param("n"), GP64())

	end := GP64()
	MOVQ(dst, end)
	ADDQ(n, end)

	ymm0 := YMM()
	ymm1 := YMM()

	Label("loop")
	Comment("Loop until we reach the end.")
	CMPQ(dst, end)
	JE(LabelRef("done"))

	Comment("Load operands in registers, apply the OR operation, assign the result.")
	dstMem := Mem{Base: dst}
	srcMem := Mem{Base: src}
	VMOVUPS(dstMem, ymm0)
	VMOVUPS(dstMem.Offset(32), ymm1)

	VPOR(srcMem, ymm0, ymm0)
	VPOR(srcMem.Offset(32), ymm1, ymm1)

	VMOVUPS(ymm0, dstMem)
	VMOVUPS(ymm1, dstMem.Offset(32))

	Comment("Advance pointers.")
	ADDQ(Imm(64), dst)
	ADDQ(Imm(64), src)
	JMP(LabelRef("loop"))

	Label("done")
	RET()
	Generate()
}
