// +build ignore

package main

import (
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	//. "github.com/mmcloughlin/avo/reg"
)

func main() {
	generateIndexPair(indexPair1{})
	generateIndexPair(indexPair2{})
	generateIndexPair(indexPair4{})
	generateIndexPair(indexPair8{})
	generateIndexPair(indexPair16{})
	Generate()
}

type indexPair interface {
	size() int
	reg() Op
	mov(Op, Op)
	cmp(Op, Op)
	//test(...[2]Op)
}

type indexPair1 struct{}

func (indexPair1) size() int   { return 1 }
func (indexPair1) reg() Op     { return GP8() }
func (indexPair1) mov(a, b Op) { MOVB(a, b) }
func (indexPair1) cmp(a, b Op) { CMPB(a, b) }

type indexPair2 struct{}

func (indexPair2) size() int   { return 2 }
func (indexPair2) reg() Op     { return GP16() }
func (indexPair2) mov(a, b Op) { MOVW(a, b) }
func (indexPair2) cmp(a, b Op) { CMPW(a, b) }

type indexPair4 struct{}

func (indexPair4) size() int   { return 4 }
func (indexPair4) reg() Op     { return GP32() }
func (indexPair4) mov(a, b Op) { MOVL(a, b) }
func (indexPair4) cmp(a, b Op) { CMPL(a, b) }

type indexPair8 struct{}

func (indexPair8) size() int   { return 8 }
func (indexPair8) reg() Op     { return GP64() }
func (indexPair8) mov(a, b Op) { MOVQ(a, b) }
func (indexPair8) cmp(a, b Op) { CMPQ(a, b) }

type indexPair16 struct{}

func (indexPair16) size() int   { return 16 }
func (indexPair16) reg() Op     { return XMM() }
func (indexPair16) mov(a, b Op) { MOVOU(a, b) }
func (indexPair16) cmp(a, b Op) {
	r := GP32()
	PCMPEQQ(a, b)
	PMOVMSKB(b, r)
	CMPL(r, U32(0xFFFF))
}

func generateIndexPair(code indexPair) {
	size := code.size()
	TEXT(fmt.Sprintf("indexPair%d", size), NOSPLIT, "func(b []byte) int")

	p := Load(Param("b").Base(), GP64())
	n := Load(Param("b").Len(), GP64())

	CMPQ(n, Imm(uint64(size))) // zero or one item
	JBE(LabelRef("done"))

	ptr := GP64()
	end := GP64()
	MOVQ(p, ptr)
	MOVQ(p, end)
	ADDQ(n, end)

	//CMPQ(n, Imm(4*uint64(size)))
	//JE(LabelRef("loop4"))

	p0 := ptr
	p1 := GP64()
	MOVQ(p0, p1)
	ADDQ(Imm(uint64(size)), p1)

	Label("loop1")
	r0 := code.reg()
	r1 := code.reg()
	code.mov(Mem{Base: p0}, r0)
	code.mov(Mem{Base: p1}, r1)
	code.cmp(r0, r1)
	JE(LabelRef("found"))
	ADDQ(Imm(uint64(size)), p0)
	ADDQ(Imm(uint64(size)), p1)
	CMPQ(ptr, end)
	JNE(LabelRef("loop1"))
	JMP(LabelRef("done"))

	/*
		Label("loop4")
		r0 := code.reg()
		r1 := code.reg()
		r2 := code.reg()
		r3 := code.reg()
		code.mov(Mem{Base: ptr}, r0)
		code.mov((Mem{Base: ptr}).Offset(1*size), r1)
		code.mov((Mem{Base: ptr}).Offset(2*size), r2)
		code.mov((Mem{Base: ptr}).Offset(3*size), r3)
		code.test(
			[2]Op{r0, r1},
			[2]Op{r0, r1},
			[2]Op{r0, r1},
			[2]Op{r0, r1},
		)
		JE(LabelRef("found"))
		ADDQ(Imm(4*uint64(size)), ptr)
		CMPQ(ptr, end)
		JNE(LabelRef("loop4"))
		JMP(LabelRef("done"))
	*/

	Label("done")
	Store(n, ReturnIndex(0))
	RET()

	Label("found")
	// The delta between the base pointer and how far we advanced is the index of the pair.
	i := ptr
	SUBQ(p, i)
	//SUBQ(Imm(uint64(size)), i)
	Store(i, ReturnIndex(0))
	RET()
}
