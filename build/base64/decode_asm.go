// +build ignore
//
// This code is a go assembly implementation of:
//
// Muła, Wojciech, & Lemire, Daniel (Thu, 14 Jun 2018).
//   Faster Base64 Encoding and Decoding Using AVX2 Instructions.
//   [arXiv:1704.00605](https://arxiv.org/abs/1704.00605)
//
// ...with changes to support multiple encodings.
package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/gotypes"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/asm"
	. "github.com/segmentio/asm/build/internal/x86"
)

var lutHi = ConstBytes("b64_dec_lut_hi", []byte{
	16, 16, 1, 2, 4, 8, 4, 8, 16, 16, 16, 16, 16, 16, 16, 16,
	16, 16, 1, 2, 4, 8, 4, 8, 16, 16, 16, 16, 16, 16, 16, 16,
})

var madd1 = ConstBytes("b64_dec_madd1", []byte{
	64, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64, 1,
	64, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64, 1,
})

var madd2 = ConstArray16("b64_dec_madd2",
	4096, 1, 4096, 1, 4096, 1, 4096, 1,
	4096, 1, 4096, 1, 4096, 1, 4096, 1,
)

var shufLo = ConstBytes("b64_dec_shuf_lo", []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 1, 0, 6,
})

var shuf = ConstBytes("b64_dec_shuf", []byte{
	2, 1, 0, 6, 5, 4, 10, 9, 8, 14, 13, 12, 0, 0, 0, 0,
	5, 4, 10, 9, 8, 14, 13, 12, 0, 0, 0, 0, 0, 0, 0, 0,
})

func init() {
	ConstraintExpr("!purego")
}

func main() {
	TEXT("decodeAVX2", NOSPLIT, "func(dst, src []byte, lut [32]int8) (int, int)")
	createDecode(Param("dst"), Param("src"), Param("lut"), func(m Mem, r VecVirtual) {
		VMOVDQU(m, r)
	})

	TEXT("decodeAVX2URI", NOSPLIT, "func(dst, src []byte, lut [32]int8) (int, int)")
	slash := VecBroadcast(U8('/'), YMM())
	underscore := VecBroadcast(U8('_'), YMM())
	createDecode(Param("dst"), Param("src"), Param("lut"), func(m Mem, r VecVirtual) {
		eq := YMM()
		VMOVDQU(m, r)
		VPCMPEQB(r, underscore, eq)
		VPBLENDVB(eq, slash, r, r)
	})

	Generate()
}

func createDecode(pdst, psrc, plut Component, load func(m Mem, r VecVirtual)) {
	dst := Mem{Base: Load(pdst.Base(), GP64()), Index: GP64(), Scale: 1}
	src := Mem{Base: Load(psrc.Base(), GP64()), Index: GP64(), Scale: 1}
	rem := Load(psrc.Len(), GP64())
	lut, err := plut.Index(0).Resolve()
	if err != nil {
		panic(err)
	}

	rsrc := YMM()
	rdst := YMM()
	nibh := YMM()
	nibl := YMM()
	emsk := YMM()
	roll := YMM()
	shfl := YMM()
	lutl := YMM()
	luth := YMM()
	lutr := YMM()
	zero := YMM()
	lo := YMM()
	hi := YMM()
	mask := VecBroadcast(U8(0x2f), YMM())

	XORQ(dst.Index, dst.Index)
	XORQ(src.Index, src.Index)
	VPXOR(zero, zero, zero)

	VPERMQ(Imm(1<<6|1<<2), lut.Addr, lutr)
	VPERMQ(Imm(1<<6|1<<2), lut.Addr.Offset(16), lutl)
	VMOVDQA(lutHi, luth)

	Label("loop")

	load(src, rsrc)

	VPSRLD(Imm(4), rsrc, nibh)
	VPAND(mask, rsrc, nibl)
	VPSHUFB(nibl, lutl, lo)
	VPAND(mask, nibh, nibh)
	VPSHUFB(nibh, luth, hi)

	VPTEST(hi, lo)
	JNE(LabelRef("done"))

	VPCMPEQB(mask, rsrc, emsk)
	VPADDB(emsk, nibh, roll)

	VPSHUFB(roll, lutr, roll)

	VPADDB(rsrc, roll, shfl)
	VPMADDUBSW(madd1, shfl, shfl)
	VPMADDWD(madd2, shfl, shfl)

	VEXTRACTI128(Imm(1), shfl, rdst.AsX())
	VPSHUFB(shufLo, rdst.AsX(), rdst.AsX())
	VPSHUFB(shuf, shfl, shfl)

	VPBLENDD(Imm(8), rdst, shfl, rdst)
	VPBLENDD(Imm(192), zero, rdst, rdst)
	VMOVDQU(rdst, dst)

	ADDQ(Imm(24), dst.Index)
	ADDQ(Imm(32), src.Index)
	SUBQ(Imm(32), rem)

	CMPQ(rem, Imm(45))
	JB(LabelRef("done"))
	JMP(LabelRef("loop"))

	Label("done")
	Store(dst.Index, ReturnIndex(0))
	Store(src.Index, ReturnIndex(1))
	VZEROUPPER()
	RET()
}
