// +build ignore

package main

import (
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func main() {
	generateIndexPair(indexPair{
		size: 1,
		GP:   GP8,
		MOV:  MOVB,
		CMP:  CMPB,
	})

	generateIndexPair(indexPair{
		size: 2,
		GP:   GP16,
		MOV:  MOVW,
		CMP:  CMPW,
	})

	generateIndexPair(indexPair{
		size: 4,
		GP:   GP32,
		MOV:  MOVL,
		CMP:  CMPL,
	})

	generateIndexPair(indexPair{
		size: 8,
		GP:   GP64,
		MOV:  MOVQ,
		CMP:  CMPQ,
	})

	Generate()
}

type indexPair struct {
	size int
	GP   func() GPVirtual
	MOV  func(Op, Op)
	CMP  func(Op, Op)
}

func generateIndexPair(code indexPair) {
	TEXT(fmt.Sprintf("indexPair%d", code.size), NOSPLIT, "func(b []byte) int")

	p := Load(Param("b").Base(), GP64())
	n := Load(Param("b").Len(), GP64())

	CMPQ(n, Imm(uint64(code.size))) // zero or one item
	JBE(LabelRef("done"))

	ptr := GP64()
	end := GP64()
	MOVQ(p, ptr)
	MOVQ(p, end)
	ADDQ(n, end)

	Label("loop")
	r0 := code.GP()
	r1 := code.GP()
	code.MOV(Mem{Base: ptr}, r0)
	code.MOV((Mem{Base: ptr}).Offset(1*code.size), r1)
	code.CMP(r0, r1)
	JE(LabelRef("found"))
	ADDQ(Imm(uint64(code.size)), ptr)
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
}
