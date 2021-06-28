package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/segmentio/asm/build/internal/asm"
	. "github.com/segmentio/asm/build/internal/x86"
)

func main() {
	TEXT("decodeAVX2", NOSPLIT, "func(dst, src []byte, lut [16]int8) (int, int)")

	dst := Mem{Base: Load(Param("dst").Base(), GP64()), Index: GP64(), Scale: 1}
	src := Mem{Base: Load(Param("src").Base(), GP64()), Index: GP64(), Scale: 1}
	rem := Load(Param("src").Len(), GP64())
	lut, _ := Param("lut").Index(0).Resolve()

	rsrc := YMM()
	rdst := YMM()
	nibh := YMM()
	nibl := YMM()
	emsk := YMM()
	roll := YMM()
	shfl := YMM()
	lutl := YMM()
	luth := YMM()
	xtab := YMM()
	zero := YMM()
	lo := YMM()
	hi := YMM()
	mask := VecBroadcast(U8(47), YMM())

	XORQ(dst.Index, dst.Index)
	XORQ(src.Index, src.Index)
	VPXOR(zero, zero, zero)

	Comment("Load the 16-byte LUT into both lanes of the register")
	VPERMQ(Imm(1<<6|1<<2), lut.Addr, xtab)

	VMOVDQA(ConstBytes("b64_dec_lut_lo", []byte{
		21, 17, 17, 17, 17, 17, 17, 17, 17, 17, 19, 26, 27, 27, 27, 26,
		21, 17, 17, 17, 17, 17, 17, 17, 17, 17, 19, 26, 27, 27, 27, 26,
	}), lutl)
	VMOVDQA(ConstBytes("b64_dec_lut_hi", []byte{
		16, 16, 1, 2, 4, 8, 4, 8, 16, 16, 16, 16, 16, 16, 16, 16,
		16, 16, 1, 2, 4, 8, 4, 8, 16, 16, 16, 16, 16, 16, 16, 16,
	}), luth)

	Label("loop")

	VMOVDQU(src, rsrc)

	VPSRLD(Imm(4), rsrc, nibh)
	VPAND(mask, rsrc, nibl)
	VPSHUFB(nibl, lutl, lo)
	VPAND(mask, nibh, nibh)
	VPSHUFB(nibh, luth, hi)

	VPTEST(hi, lo)
	JNE(LabelRef("done"))

	VPCMPEQB(mask, rsrc, emsk)
	VPADDB(emsk, nibh, roll)

	VPSHUFB(roll, xtab, roll)

	VPADDB(rsrc, roll, shfl)
	VPMADDUBSW(ConstBytes("b64_dec_maddub", []byte{
		64, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64, 1,
		64, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64, 1,
	}), shfl, shfl)

	VPMADDWD(ConstArray16("b64_dec_madd",
		4096, 1, 4096, 1, 4096, 1, 4096, 1,
		4096, 1, 4096, 1, 4096, 1, 4096, 1,
	), shfl, shfl)

	VEXTRACTI128(Imm(1), shfl, rdst.AsX())
	VPSHUFB(ConstBytes("b64_dec_shuf_lo", []byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 1, 0, 6,
	}), rdst.AsX(), rdst.AsX())

	VPSHUFB(ConstBytes("b64_dec_shuf", []byte{
		2, 1, 0, 6, 5, 4, 10, 9, 8, 14, 13, 12, 0, 0, 0, 0,
		5, 4, 10, 9, 8, 14, 13, 12, 0, 0, 0, 0, 0, 0, 0, 0,
	}), shfl, shfl)

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
	RET()

	Generate()
}
