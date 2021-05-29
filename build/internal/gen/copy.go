package gen

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/x86"

	"github.com/segmentio/asm/cpu"
)

// Copy is a generator for copy-like functions.
type Copy struct {
	// Copy functions generating instructions for various item sizes.
	CopyB func(dst, src Op)
	CopyW func(dst, src Op)
	CopyL func(dst, src Op)
	CopyQ func(dst, src Op)
	// If non-nil, applies the AVX instruction to transform the inputs
	// in src0 and src1 into dst.
	CopyAVX func(src0, src1, dst Op)
}

func (c *Copy) Generate(name, doc string) {
	TEXT(name, NOSPLIT, "func(dst, src []byte) int")
	Doc(name + " " + doc)

	dst := Load(Param("dst").Base(), GP64())
	src := Load(Param("src").Base(), GP64())

	r := Load(Param("dst").Len(), GP64())
	x := Load(Param("src").Len(), GP64())

	CMPQ(x, r)
	CMOVQGT(x, r)

	n := GP64()
	MOVQ(r, n)

	JumpIfFeature("avx2", cpu.AVX2)

	Label("cmp8")
	CMPQ(n, Imm(8))
	JB(LabelRef("cmp4"))
	gp64 := GP64()
	MOVQ(Mem{Base: src}, gp64)
	c.CopyQ(gp64, Mem{Base: dst})
	ADDQ(Imm(8), dst)
	ADDQ(Imm(8), src)
	SUBQ(Imm(8), n)
	JMP(LabelRef("cmp8"))

	Label("cmp4")
	CMPQ(n, Imm(4))
	JB(LabelRef("cmp2"))
	gp32 := GP32()
	MOVL(Mem{Base: src}, gp32)
	c.CopyL(gp32, Mem{Base: dst})
	ADDQ(Imm(4), dst)
	ADDQ(Imm(4), src)
	SUBQ(Imm(4), n)

	Label("cmp2")
	CMPQ(n, Imm(2))
	JB(LabelRef("cmp1"))
	gp16 := GP16()
	MOVW(Mem{Base: src}, gp16)
	c.CopyW(gp16, Mem{Base: dst})
	ADDQ(Imm(2), dst)
	ADDQ(Imm(2), src)
	SUBQ(Imm(2), n)

	Label("cmp1")
	CMPQ(n, Imm(1))
	JB(LabelRef("done"))
	gp8 := GP8()
	MOVB(Mem{Base: src}, gp8)
	c.CopyB(gp8, Mem{Base: dst})

	Label("done")
	Store(r, ReturnIndex(0))
	RET()

	Label("avx2")
	Label("cmp128")
	CMPQ(n, Imm(128))
	JB(LabelRef("cmp64"))

	VMOVDQU(Mem{Base: src}, Y0)
	VMOVDQU((Mem{Base: src}).Offset(32), Y1)
	VMOVDQU((Mem{Base: src}).Offset(64), Y2)
	VMOVDQU((Mem{Base: src}).Offset(96), Y3)
	if c.CopyAVX != nil {
		c.CopyAVX(Mem{Base: dst}, Y0, Y0)
		c.CopyAVX((Mem{Base: dst}).Offset(32), Y1, Y1)
		c.CopyAVX((Mem{Base: dst}).Offset(64), Y2, Y2)
		c.CopyAVX((Mem{Base: dst}).Offset(96), Y3, Y3)
	}
	VMOVDQU(Y0, Mem{Base: dst})
	VMOVDQU(Y1, (Mem{Base: dst}).Offset(32))
	VMOVDQU(Y2, (Mem{Base: dst}).Offset(64))
	VMOVDQU(Y3, (Mem{Base: dst}).Offset(96))

	ADDQ(Imm(128), dst)
	ADDQ(Imm(128), src)
	SUBQ(Imm(128), n)
	JMP(LabelRef("cmp128"))

	Label("cmp64")
	CMPQ(n, Imm(64))
	JB(LabelRef("cmp32"))

	VMOVDQU(Mem{Base: src}, Y0)
	VMOVDQU((Mem{Base: src}).Offset(32), Y1)
	if c.CopyAVX != nil {
		c.CopyAVX(Mem{Base: dst}, Y0, Y0)
		c.CopyAVX((Mem{Base: dst}).Offset(32), Y1, Y1)
	}
	VMOVDQU(Y0, Mem{Base: dst})
	VMOVDQU(Y1, (Mem{Base: dst}).Offset(32))

	ADDQ(Imm(64), dst)
	ADDQ(Imm(64), src)
	SUBQ(Imm(64), n)

	Label("cmp32")
	CMPQ(n, Imm(32))
	JB(LabelRef("cmp8"))

	dstTail := GP64()
	srcTail := GP64()
	MOVQ(dst, dstTail)
	MOVQ(src, srcTail)
	ADDQ(n, dstTail)
	ADDQ(n, srcTail)
	SUBQ(Imm(32), dstTail)
	SUBQ(Imm(32), srcTail)

	VMOVDQU(Mem{Base: src}, Y0)
	VMOVDQU(Mem{Base: srcTail}, Y1)
	if c.CopyAVX != nil {
		c.CopyAVX(Mem{Base: dst}, Y0, Y0)
		c.CopyAVX(Mem{Base: dstTail}, Y1, Y1)
	}
	VMOVDQU(Y0, Mem{Base: dst})
	VMOVDQU(Y1, Mem{Base: dstTail})
	JMP(LabelRef("done"))

	Generate()
}
