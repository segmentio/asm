package x86

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"

	"github.com/segmentio/asm/cpu"
)

func CopyB(src, dst Register, store func(Op, Op)) {
	LoadAndStore(src, dst, GP8(), MOVB, store)
}

func CopyW(src, dst Register, store func(Op, Op)) {
	LoadAndStore(src, dst, GP16(), MOVW, store)
}

func CopyL(src, dst Register, store func(Op, Op)) {
	LoadAndStore(src, dst, GP32(), MOVL, store)
}

func CopyQ(src, dst Register, store func(Op, Op)) {
	LoadAndStore(src, dst, GP64(), MOVQ, store)
}

func LoadAndStore(src, dst, tmp Register, load func(Op, Op), store func(Op, Op)) {
	load(Mem{Base: src}, tmp)
	store(tmp, Mem{Base: dst})
}

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

	Doc("Generic copy for small inputs or targets without AVX instructions.")
	Label("cmp8")
	CMPQ(n, Imm(8))
	JB(LabelRef("cmp4"))
	CopyQ(src, dst, c.CopyQ)
	ADDQ(Imm(8), dst)
	ADDQ(Imm(8), src)
	SUBQ(Imm(8), n)
	JMP(LabelRef("cmp8"))

	Label("cmp4")
	CMPQ(n, Imm(4))
	JB(LabelRef("cmp2"))
	CopyL(src, dst, c.CopyL)
	ADDQ(Imm(4), dst)
	ADDQ(Imm(4), src)
	SUBQ(Imm(4), n)

	Label("cmp2")
	CMPQ(n, Imm(2))
	JB(LabelRef("cmp1"))
	CopyW(src, dst, c.CopyW)
	ADDQ(Imm(2), dst)
	ADDQ(Imm(2), src)
	SUBQ(Imm(2), n)

	Label("cmp1")
	CMPQ(n, Imm(1))
	JB(LabelRef("done"))
	CopyB(src, dst, c.CopyB)

	Label("done")
	Store(r, ReturnIndex(0))
	RET()

	Doc("AVX optimized version for medium to large size inputs.")
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
