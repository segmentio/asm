// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/segmentio/asm/build/internal/asm"
	. "github.com/segmentio/asm/build/internal/x86"
)

func main() {
	TEXT("encodeAVX2", NOSPLIT, "func(dst, src []byte, lut [16]int8) (int, int)")

	dst := Mem{Base: Load(Param("dst").Base(), GP64()), Index: GP64(), Scale: 1}
	src := Mem{Base: Load(Param("src").Base(), GP64()), Index: GP64(), Scale: 1}
	rem := Load(Param("src").Len(), GP64())
	lut, _ := Param("lut").Index(0).Resolve()

	rsrc := YMM()
	rdst := YMM()
	msrc := YMM()
	shl4 := YMM()
	shl8 := YMM()
	blnd := YMM()
	mult := YMM()
	shfl := YMM()
	subs := YMM()
	cmps := YMM()
	xlat := YMM()
	xtab := YMM()
	xsub := VecBroadcast(U8(51), YMM())
	xcmp := VecBroadcast(U8(25), YMM())

	XORQ(dst.Index, dst.Index)
	XORQ(src.Index, src.Index)

	Comment("Load the 16-byte LUT into both lanes of the register")
	VPERMQ(Imm(1<<6|1<<2), lut.Addr, xtab)

	Comment("Load the first block using a mask to avoid potential fault")
	VMOVDQU(ConstLoadMask32("b64_enc_load",
		0, 1, 1, 1,
		1, 1, 1, 1,
	), rsrc)
	VPMASKMOVD(src.Offset(-4), rsrc, rsrc)

	Label("loop")

	VPSHUFB(ConstBytes("b64_enc_shuf", []byte{
		5, 4, 6, 5, 8, 7, 9, 8, 11, 10, 12, 11, 14, 13, 15, 14,
		1, 0, 2, 1, 4, 3, 5, 4, 7, 6, 8, 7, 10, 9, 11, 10,
	}), rsrc, rsrc)

	VPAND(ConstArray16("b64_enc_mask1",
		0x03f0, 0x003f, 0x03f0, 0x003f, 0x03f0, 0x003f, 0x03f0, 0x003f,
		0x03f0, 0x003f, 0x03f0, 0x003f, 0x03f0, 0x003f, 0x03f0, 0x003f,
	), rsrc, msrc)
	VPSLLW(Imm(8), msrc, shl8)
	VPSLLW(Imm(4), msrc, shl4)
	VPBLENDW(Imm(170), shl8, shl4, blnd)

	VPAND(ConstArray16("b64_enc_mask2",
		0xfc00, 0x0fc0, 0xfc00, 0x0fc0, 0xfc00, 0x0fc0, 0xfc00, 0x0fc0,
		0xfc00, 0x0fc0, 0xfc00, 0x0fc0, 0xfc00, 0x0fc0, 0xfc00, 0x0fc0,
	), rsrc, msrc)
	VPMULHUW(ConstArray16("b64_enc_mult",
		0x0040, 0x0400, 0x0040, 0x0400, 0x0040, 0x0400, 0x0040, 0x0400,
		0x0040, 0x0400, 0x0040, 0x0400, 0x0040, 0x0400, 0x0040, 0x0400,
	), msrc, mult)

	VPOR(mult, blnd, shfl)

	VPSUBUSB(xsub, shfl, subs)
	VPCMPGTB(xcmp, shfl, cmps)
	VPSUBB(cmps, subs, subs)
	VPSHUFB(subs, xtab, xlat)
	VPADDB(shfl, xlat, rdst)
	VMOVDQU(rdst, dst)

	ADDQ(Imm(32), dst.Index)
	ADDQ(Imm(24), src.Index)
	SUBQ(Imm(24), rem)

	CMPQ(rem, Imm(32))
	JB(LabelRef("done"))

	VMOVDQU(src.Offset(-4), rsrc)
	JMP(LabelRef("loop"))

	Label("done")
	Store(dst.Index, ReturnIndex(0))
	Store(src.Index, ReturnIndex(1))
	RET()

	Generate()
}
