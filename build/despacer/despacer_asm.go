// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/segmentio/asm/build/internal/x86"
)

func main() {
	TEXT("Despace", NOSPLIT, "func(data []byte)")
	Doc("remove spaces (in-place) from string bytes (UTF-8 or ASCII)")
	ptr := Mem{Base: Load(Param("data").Base(), GP64())}
	len := Load(Param("data").Len(), GP64())

	idx := GP64()
	XORQ(idx, idx)
	spaces := VecBroadcast(U8(' '), YMM())
	newline := VecBroadcast(U8('\n'), YMM())
	carriage := VecBroadcast(U8('\r'), YMM())

	Label("avx2_loop")
	next := GP64()
	MOVQ(idx, next)
	ahead := GP64()
	Comment("just created ahead")
	MOVQ(U64(32), ahead)
	ADDQ(ahead, next)
	CMPQ(next, len)
	JAE(LabelRef("x86_loop"))

	y := YMM()
	yspaces := YMM()
	ynewlines := YMM()
	ycarriages := YMM()
	VLDDQU(ptr, y)
	VPCMPEQB(y, spaces, yspaces)
	VPCMPEQB(y, newline, ynewlines)
	VPCMPEQB(y, carriage, ycarriages)
	results := YMM()
	VPOR(yspaces, ynewlines, results)
	VPOR(ycarriages, results, results)
	zero := YMM()
	VPOR(zero, zero, zero)
	VPTEST(results, results)
	JNZ(LabelRef("remove_space"))
	// Increment ptrs and loop.
	MOVQ(next, idx)
	JMP(LabelRef("avx2_loop"))

	// Increment ptrs and loop.
	Label("remove_space")
	swapIdx := GP64()
	MOVQ(next, swapIdx)
	CMPQ(last, ahead)
	//CMOVQLT(yLen, len)

	Label("x86_loop")
	// Do something
	JAE(LabelRef("return"))

	Label("return")
	RET()
	Generate()

}
