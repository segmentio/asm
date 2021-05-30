package x86

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"

	"github.com/segmentio/asm/cpu"
)

func CopyB(src, dst Mem, store func(Op, Op)) {
	LoadAndStore(src, dst, GP8(), MOVB, store)
}

func CopyW(src, dst Mem, store func(Op, Op)) {
	LoadAndStore(src, dst, GP16(), MOVW, store)
}

func CopyL(src, dst Mem, store func(Op, Op)) {
	LoadAndStore(src, dst, GP32(), MOVL, store)
}

func CopyQ(src, dst Mem, store func(Op, Op)) {
	LoadAndStore(src, dst, GP64(), MOVQ, store)
}

func LoadAndStore(src, dst Mem, tmp Register, load func(Op, Op), store func(Op, Op)) {
	load(src, tmp)
	store(tmp, dst)
}

// Copy is a generator for copy-like functions.
type Copy struct {
	// Copy functions generating instructions for various item sizes.
	CopyB func(dst, src Op)
	CopyW func(dst, src Op)
	CopyL func(dst, src Op)
	CopyQ func(dst, src Op)
	// If non-nil, applies the SSE instruction to transform the input
	// in src into dst.
	CopySSE func(src, dst Op)
	// If non-nil, applies the AVX instruction to transform the inputs
	// in src0 and src1 into dst.
	CopyAVX func(src0, src1, dst Op)
}

func (c *Copy) Generate(name, doc string) {
	TEXT(name, NOSPLIT, "func(dst, src []byte) int")
	Doc(name + " " + doc)

	dst := Load(Param("dst").Base(), GP64())
	src := Load(Param("src").Base(), GP64())

	n := Load(Param("dst").Len(), GP64())
	x := Load(Param("src").Len(), GP64())

	CMPQ(x, n)
	CMOVQGT(x, n)
	Store(n, ReturnIndex(0))

	Comment("Tail copy with special cases for each possible item size.")
	Label("tail")

	CMPQ(n, Imm(0))
	JE(LabelRef("done"))

	CMPQ(n, Imm(2))
	JBE(LabelRef("copy1to2"))

	CMPQ(n, Imm(3))
	JE(LabelRef("copy3"))

	CMPQ(n, Imm(4))
	JE(LabelRef("copy4"))

	CMPQ(n, Imm(8))
	JB(LabelRef("copy5to7"))
	JE(LabelRef("copy8"))

	CMPQ(n, Imm(16))
	JBE(LabelRef("copy9to16"))

	CMPQ(n, Imm(32))
	JBE(LabelRef("copy17to32"))

	CMPQ(n, Imm(64))
	JBE(LabelRef("copy33to64"))

	JumpIfFeature("avx2", cpu.AVX2)

	Comment("Generic copy for targets without AVX instructions.")
	Label("generic")
	CopyQ(Mem{Base: src}, Mem{Base: dst}, c.CopyQ)
	ADDQ(Imm(8), src)
	ADDQ(Imm(8), dst)
	SUBQ(Imm(8), n)
	CMPQ(n, Imm(8))
	JBE(LabelRef("tail"))
	JMP(LabelRef("generic"))

	Label("done")
	RET()

	Label("copy1to2")
	copy1to2Reg0 := GP8()
	copy1to2Reg1 := GP8()
	MOVB(Mem{Base: src}, copy1to2Reg0)
	MOVB((Mem{Base: src}).Idx(n, 1).Offset(-1), copy1to2Reg1)
	c.CopyB(copy1to2Reg0, Mem{Base: dst})
	c.CopyB(copy1to2Reg1, (Mem{Base: dst}).Idx(n, 1).Offset(-1))
	RET()

	Label("copy3")
	CopyW(Mem{Base: src}, Mem{Base: dst}, c.CopyW)
	CopyB((Mem{Base: src}).Offset(2), (Mem{Base: dst}).Offset(2), c.CopyB)
	RET()

	Label("copy4")
	CopyL(Mem{Base: src}, Mem{Base: dst}, c.CopyL)
	RET()

	Label("copy5to7")
	copy5to7Reg0 := GP32()
	copy5to7Reg1 := GP32()
	MOVL(Mem{Base: src}, copy5to7Reg0)
	MOVL((Mem{Base: src}).Idx(n, 1).Offset(-4), copy5to7Reg1)
	c.CopyL(copy5to7Reg0, Mem{Base: dst})
	c.CopyL(copy5to7Reg1, (Mem{Base: dst}).Idx(n, 1).Offset(-4))
	RET()

	Label("copy8")
	CopyQ(Mem{Base: src}, Mem{Base: dst}, c.CopyQ)
	RET()

	Label("copy9to16")
	copy9to16Reg0 := GP64()
	copy9to16Reg1 := GP64()
	MOVQ(Mem{Base: src}, copy9to16Reg0)
	MOVQ((Mem{Base: src}).Idx(n, 1).Offset(-8), copy9to16Reg1)
	c.CopyQ(copy9to16Reg0, Mem{Base: dst})
	c.CopyQ(copy9to16Reg1, (Mem{Base: dst}).Idx(n, 1).Offset(-8))
	RET()

	Label("copy17to32")
	MOVOU(Mem{Base: src}, X0)
	MOVOU((Mem{Base: src}).Idx(n, 1).Offset(-16), X1)
	if c.CopySSE != nil {
		MOVOU(Mem{Base: dst}, X2)
		MOVOU((Mem{Base: dst}).Idx(n, 1).Offset(-16), X3)
		c.CopySSE(X2, X0)
		c.CopySSE(X3, X1)
	}
	MOVOU(X0, Mem{Base: dst})
	MOVOU(X1, (Mem{Base: dst}).Idx(n, 1).Offset(-16))
	RET()

	Label("copy33to64")
	MOVOU(Mem{Base: src}, X0)
	MOVOU((Mem{Base: src}).Offset(16), X1)
	MOVOU((Mem{Base: src}).Idx(n, 1).Offset(-32), X2)
	MOVOU((Mem{Base: src}).Idx(n, 1).Offset(-16), X3)
	if c.CopySSE != nil {
		MOVOU(Mem{Base: dst}, X4)
		MOVOU((Mem{Base: dst}).Offset(16), X5)
		MOVOU((Mem{Base: dst}).Idx(n, 1).Offset(-32), X6)
		MOVOU((Mem{Base: dst}).Idx(n, 1).Offset(-16), X7)
		c.CopySSE(X4, X0)
		c.CopySSE(X5, X1)
		c.CopySSE(X6, X2)
		c.CopySSE(X7, X3)
	}
	MOVOU(X0, Mem{Base: dst})
	MOVOU(X1, (Mem{Base: dst}).Offset(16))
	MOVOU(X2, (Mem{Base: dst}).Idx(n, 1).Offset(-32))
	MOVOU(X3, (Mem{Base: dst}).Idx(n, 1).Offset(-16))
	RET()

	Comment("AVX optimized version for medium to large size inputs.")
	Label("avx2")
	CMPQ(n, Imm(128))
	JB(LabelRef("avx2_tail"))

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
	JMP(LabelRef("avx2"))

	Label("avx2_tail")
	JZ(LabelRef("done")) // check flags from last CMPQ

	CMPQ(n, Imm(32)) // n > 0 && n <= 32
	JBE(LabelRef("avx2_tail_1to32"))

	CMPQ(n, Imm(64)) // n > 32 && n <= 64
	JBE(LabelRef("avx2_tail_33to64"))

	CMPQ(n, Imm(96)) // n > 64 && n <= 96
	JBE(LabelRef("avx2_tail_65to96"))

	VMOVDQU(Mem{Base: src}, Y0)
	VMOVDQU((Mem{Base: src}).Offset(32), Y1)
	VMOVDQU((Mem{Base: src}).Offset(64), Y2)
	VMOVDQU((Mem{Base: src}).Idx(n, 1).Offset(-32), Y3)
	if c.CopyAVX != nil {
		c.CopyAVX(Mem{Base: dst}, Y0, Y0)
		c.CopyAVX((Mem{Base: dst}).Offset(32), Y1, Y1)
		c.CopyAVX((Mem{Base: dst}).Offset(64), Y2, Y2)
		c.CopyAVX((Mem{Base: dst}).Idx(n, 1).Offset(-32), Y3, Y3)
	}
	VMOVDQU(Y0, Mem{Base: dst})
	VMOVDQU(Y1, (Mem{Base: dst}).Offset(32))
	VMOVDQU(Y2, (Mem{Base: dst}).Offset(64))
	VMOVDQU(Y3, (Mem{Base: dst}).Idx(n, 1).Offset(-32))
	RET()

	Label("avx2_tail_65to96")
	VMOVDQU(Mem{Base: src}, Y0)
	VMOVDQU((Mem{Base: src}).Offset(32), Y1)
	VMOVDQU((Mem{Base: src, Index: n, Scale: 1}).Offset(-32), Y3)
	if c.CopyAVX != nil {
		c.CopyAVX(Mem{Base: dst}, Y0, Y0)
		c.CopyAVX((Mem{Base: dst}).Offset(32), Y1, Y1)
		c.CopyAVX((Mem{Base: dst}).Idx(n, 1).Offset(-32), Y3, Y3)
	}
	VMOVDQU(Y0, Mem{Base: dst})
	VMOVDQU(Y1, (Mem{Base: dst}).Offset(32))
	VMOVDQU(Y3, (Mem{Base: dst}).Idx(n, 1).Offset(-32))
	RET()

	Label("avx2_tail_33to64")
	VMOVDQU(Mem{Base: src}, Y0)
	VMOVDQU((Mem{Base: src, Index: n, Scale: 1}).Offset(-32), Y3)
	if c.CopyAVX != nil {
		c.CopyAVX(Mem{Base: dst}, Y0, Y0)
		c.CopyAVX((Mem{Base: dst}).Idx(n, 1).Offset(-32), Y3, Y3)
	}
	VMOVDQU(Y0, Mem{Base: dst})
	VMOVDQU(Y3, (Mem{Base: dst}).Idx(n, 1).Offset(-32))
	RET()

	Label("avx2_tail_1to32")
	VMOVDQU((Mem{Base: src}).Idx(n, 1).Offset(-32), Y3)
	if c.CopyAVX != nil {
		c.CopyAVX((Mem{Base: dst}).Idx(n, 1).Offset(-32), Y3, Y3)
	}
	VMOVDQU(Y3, (Mem{Base: dst}).Idx(n, 1).Offset(-32))
	RET()

	Generate()
}
