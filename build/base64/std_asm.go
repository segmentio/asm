// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/segmentio/asm/build/internal/asm"
	. "github.com/segmentio/asm/build/internal/x86"
)

func main() {
	stdEncMask := ConstLoadMask32("stdEncMask",
		0, 1, 1, 1,
		1, 1, 1, 1,
	)

	stdEncShfl := ConstBytes("stdEncShfl", []byte{
		0x05, 0x04, 0x06, 0x05, 0x08, 0x07, 0x09, 0x08,
		0x0b, 0x0a, 0x0c, 0x0b, 0x0e, 0x0d, 0x0f, 0x0e,
		0x01, 0x00, 0x02, 0x01, 0x04, 0x03, 0x05, 0x04,
		0x07, 0x06, 0x08, 0x07, 0x0a, 0x09, 0x0b, 0x0a,
	})

	stdEncAnd1 := ConstArray16("stdEncAnd1",
		0x03f0, 0x003f, 0x03f0, 0x003f, 0x03f0, 0x003f, 0x03f0, 0x003f,
		0x03f0, 0x003f, 0x03f0, 0x003f, 0x03f0, 0x003f, 0x03f0, 0x003f,
	)

	stdEncAnd2 := ConstArray16("stdEncAnd2",
		0xfc00, 0x0fc0, 0xfc00, 0x0fc0, 0xfc00, 0x0fc0, 0xfc00, 0x0fc0,
		0xfc00, 0x0fc0, 0xfc00, 0x0fc0, 0xfc00, 0x0fc0, 0xfc00, 0x0fc0,
	)

	stdEncMult := ConstArray16("stdEncMult",
		0x0040, 0x0400, 0x0040, 0x0400, 0x0040, 0x0400, 0x0040, 0x0400,
		0x0040, 0x0400, 0x0040, 0x0400, 0x0040, 0x0400, 0x0040, 0x0400,
	)

	stdEncTabl := ConstBytes("stdEncTabl", []byte{
		0x41, 0x47, 0xfc, 0xfc, 0xfc, 0xfc, 0xfc, 0xfc,
		0xfc, 0xfc, 0xfc, 0xfc, 0xed, 0xf0, 0x00, 0x00,
		0x41, 0x47, 0xfc, 0xfc, 0xfc, 0xfc, 0xfc, 0xfc,
		0xfc, 0xfc, 0xfc, 0xfc, 0xed, 0xf0, 0x00, 0x00,
	})

	TEXT("stdEncodeAVX2", NOSPLIT, "func(dst, src []byte) (int, int)")

	dst := Mem{Base: Load(Param("dst").Base(), GP64()), Index: GP64(), Scale: 1}
	src := Mem{Base: Load(Param("src").Base(), GP64()), Index: GP64(), Scale: 1}
	rem := Load(Param("src").Len(), GP64())

	ymsk := YMM()
	ysrc := YMM()
	shfl := YMM()
	xlat := YMM()
	ycmp := YMM()
	ysub := YMM()
	yout := YMM()
	tmp0 := YMM()
	tmp1 := YMM()

	XORQ(dst.Index, dst.Index)
	XORQ(src.Index, src.Index)

	//CMPQ(rem, Imm(28))
	//JB(LabelRef("done"))

	xlatSub := VecBroadcast(U8(51), YMM())
	xlatCmp := VecBroadcast(U8(25), YMM())
	xlatTbl := YMM()
	VMOVDQU(stdEncTabl, xlatTbl)

	Comment("Load the first block using a mask to avoid potential fault")
	VMOVDQU(stdEncMask, ymsk)
	VPMASKMOVD(src.Offset(-4), ymsk, ysrc)

	Label("loop")

	VPSHUFB(stdEncShfl, ysrc, ysrc)

	VPAND(stdEncAnd1, ysrc, tmp0)
	VPSLLW(Imm(8), tmp0, tmp1)
	VPSLLW(Imm(4), tmp0, tmp0)
	VPBLENDW(Imm(170), tmp1, tmp0, tmp0)

	VPAND(stdEncAnd2, ysrc, tmp1)
	VPMULHUW(stdEncMult, tmp1, tmp1)

	VPOR(tmp1, tmp0, shfl)

	VPSUBUSB(xlatSub, shfl, ysub)
	VPCMPGTB(xlatCmp, shfl, ycmp)
	VPSUBB(ycmp, ysub, ysub)
	VPSHUFB(ysub, xlatTbl, xlat)
	VPADDB(shfl, xlat, yout)
	VMOVDQU(yout, dst)

	ADDQ(Imm(32), dst.Index)
	ADDQ(Imm(24), src.Index)
	SUBQ(Imm(24), rem)

	CMPQ(rem, Imm(32))
	JB(LabelRef("done"))

	VMOVDQU(src.Offset(-4), ysrc)
	JMP(LabelRef("loop"))

	Label("done")
	Store(dst.Index, ReturnIndex(0))
	Store(src.Index, ReturnIndex(1))
	RET()

	Generate()
}
