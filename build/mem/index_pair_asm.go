// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	//. "github.com/mmcloughlin/avo/reg"
)

func main() {
	TEXT("indexPair1", NOSPLIT, "func(b []byte) int")
	Doc("indexPair1 is the x86 specialization of mem.IndexPair for items of size 1")

	p := Load(Param("b").Base(), GP64())
	n := Load(Param("b").Len(), GP64())

	CMPQ(n, Imm(1)) // n <= 1
	JBE(LabelRef("done"))

	ptr := GP64()
	end := GP64()
	MOVQ(p, ptr)
	MOVQ(p, end)
	ADDQ(n, end)

	Label("loop")
	b0 := GP8()
	b1 := GP8()
	MOVB(Mem{Base: ptr}, b0)
	MOVB((Mem{Base: ptr}).Offset(1), b1)
	CMPB(b0, b1)
	JE(LabelRef("found"))
	INCQ(ptr)
	CMPQ(ptr, end)
	JNE(LabelRef("loop"))

	Label("done")
	Store(n, ReturnIndex(0))
	RET()

	Label("found")
	Comment("The delta between the base pointer and how far we advanced is the index of the pair.")
	i := ptr
	SUBQ(p, i)
	Store(i, ReturnIndex(0))
	RET()

	Generate()
}
