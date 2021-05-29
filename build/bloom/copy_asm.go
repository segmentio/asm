// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("Copy", NOSPLIT, "func(dst, src []byte) int")
	Doc("Copy copies the one-bits of src to dst, returning the number of bytes written.")

	dst := Param("dst")
	src := Param("src")

	d := Load(dst.Base(), GP64())
	s := Load(src.Base(), GP64())

	r := Load(dst.Len(), GP64())
	x := Load(src.Len(), GP64())

	CMPQ(x, r)
	CMOVQGT(x, r)

	n := GP64()
	MOVQ(r, n)

	Label("cmp8")
	CMPQ(n, Imm(8))
	JB(LabelRef("cmp4"))
	gp64 := GP64()
	MOVQ(Mem{Base: s}, gp64)
	ORQ(gp64, Mem{Base: d})
	ADDQ(Imm(8), d)
	ADDQ(Imm(8), s)
	SUBQ(Imm(8), n)
	JMP(LabelRef("cmp8"))

	Label("cmp4")
	CMPQ(n, Imm(4))
	JB(LabelRef("cmp2"))
	gp32 := GP32()
	MOVL(Mem{Base: s}, gp32)
	ORL(gp32, Mem{Base: d})
	ADDQ(Imm(4), d)
	ADDQ(Imm(4), s)
	SUBQ(Imm(4), n)

	Label("cmp2")
	CMPQ(n, Imm(2))
	JB(LabelRef("cmp1"))
	gp16 := GP16()
	MOVW(Mem{Base: s}, gp16)
	ORW(gp16, Mem{Base: d})
	ADDQ(Imm(2), d)
	ADDQ(Imm(2), s)
	SUBQ(Imm(2), n)

	Label("cmp1")
	CMPQ(n, Imm(1))
	JB(LabelRef("done"))
	gp8 := GP8()
	MOVB(Mem{Base: s}, gp8)
	ORB(gp8, Mem{Base: d})

	Label("done")
	Store(r, ReturnIndex(0))
	RET()

	Generate()
}
