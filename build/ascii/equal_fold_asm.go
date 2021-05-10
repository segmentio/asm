// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/x86"

	"github.com/segmentio/asm/cpu"
)

func main() {
	TEXT("EqualFoldString", NOSPLIT, "func(a, b string) bool")
	Doc(
		"EqualFoldString is a version of strings.EqualFold designed to work on ASCII",
		"input instead of UTF-8.",
		"",
		"When the program has guarantees that the input is composed of ASCII",
		"characters only, it allows for greater optimizations.",
	)

	// Use index for byte position. We have plenty of registers, and it saves an
	// ADD operation as the memory index is the same for both a and b.
	i := GP64()
	a := Mem{Base: Load(Param("a").Base(), GP64()), Index: i, Scale: 1}
	n := Load(Param("a").Len(), GP64())
	b := Mem{Base: Load(Param("b").Base(), GP64()), Index: i, Scale: 1}
	bn, _ := Param("b").Len().Resolve()
	ret, _ := ReturnIndex(0).Resolve()

	CMPQ(n, bn.Addr)      // if len(a) != len(b):
	JNE(LabelRef("done")) //   return false

	maskG := GP64()
	maskY := YMM()
	maskX := maskY.AsX() // use the lower half of maskY

	XORQ(i, i)                           // i = 0
	MOVQ(U64(0xDFDFDFDFDFDFDFDF), maskG) // maskG = 0xDFDFDFDFDFDFDFDF

	JumpUnlessFeature("eq8", cpu.AVX2)

	PINSRQ(Imm(0), maskG, maskX) // maskX[0:8] = maskG
	VPBROADCASTQ(maskX, maskY)   // maskY[0:32] = [maskX[0:8],maskX[0:8],maskX[0:8],maskX[0:8]]

	// Moving the 128-byte scanning helps the branch predictor for small inputs
	CMPQ(n, U8(128))       // if n >= 128:
	JNB(LabelRef("eq128")) //   goto eq128

	Label("eq64")
	CMPQ(n, U8(64))       // if n < 64:
	JB(LabelRef("eq32"))  //   goto eq32
	EQ64(a, b, n, maskY)  // ZF = [compare 64 bytes]
	JNE(LabelRef("done")) // return if ZF == 0

	Label("eq32")
	CMPQ(n, U8(32))       // if n < 32:
	JB(LabelRef("eq16"))  //   goto eq16
	EQ32(a, b, n, maskY)  // ZF = [compare 32 bytes]
	JNE(LabelRef("done")) // return if ZF == 0

	Label("eq16")
	CMPQ(n, U8(16))       // if n < 16:
	JB(LabelRef("eq8"))   //   goto eq8
	EQ16(a, b, n, maskX)  // ZF = [compare 16 bytes]
	JNE(LabelRef("done")) // return if ZF == 0

	Label("eq8")
	CMPQ(n, U8(8))        // if n < 8:
	JB(LabelRef("eq4"))   //   goto eq4
	EQ8(a, b, n, maskG)   // ZF = [compare 8 bytes]
	JNE(LabelRef("done")) // return if ZF == 0
	JMP(LabelRef("eq8"))  // loop eq8

	Label("eq4")
	CMPQ(n, U8(4))        // if n < 4:
	JB(LabelRef("eq3"))   //   goto eq3
	EQ4(a, b, n)          // ZF = [compare 4 bytes]
	JNE(LabelRef("done")) // return if ZF == 0

	Label("eq3")
	CMPQ(n, U8(3))        // if n < 3:
	JB(LabelRef("eq2"))   //   goto eq2
	EQ3(a, b)             // ZF = [compare 3 bytes]
	JMP(LabelRef("done")) // return ZF

	Label("eq2")
	CMPQ(n, U8(2))        // if n < 2:
	JB(LabelRef("eq1"))   //   goto eq1
	EQ2(a, b)             // ZF = [compare 2 bytes]
	JMP(LabelRef("done")) // return ZF

	Label("eq1")
	CMPQ(n, U8(0))       // if n == 0:
	JE(LabelRef("done")) //   return true
	EQ1(a, b)            // ZF = [compare 1 byte]

	Label("done")
	SETEQ(ret.Addr) // return ZF
	RET()           // ...

	Label("eq128")
	EQ128(a, b, n, maskY)  // ZF = [compare 128 bytes]
	JNE(LabelRef("done"))  // return if ZF == 0
	CMPQ(n, U8(128))       // if n < 128:
	JB(LabelRef("eq64"))   //   goto eq64
	JMP(LabelRef("eq128")) // loop eq128

	Generate()
}

func EQ1(a, b Mem) {
	eq := GP8()

	MOVB(a, eq)         // eq = a[i:i+1]
	XORB(b, eq)         // eq = b[i:i+1] ^ eq
	TESTB(U8(0xDF), eq) // ZF = (mask & eq)
}

func EQ2(a, b Mem) {
	eq := GP16()

	MOVW(a, eq)            // eq = a[i:i+2]
	XORW(b, eq)            // eq = b[i:i+2] ^ eq
	TESTW(U16(0xDFDF), eq) // ZF = (mask & eq)
}

func EQ3(a, b Mem) {
	eq := GP32()
	eqa := GP32()
	eqb := GP32()

	MOVWLZX(a, eq)            // eq = a[i:i+2]
	MOVBLZX(a.Offset(2), eqa) // eqa = a[i+2:i+3]
	SHLL(U8(16), eqa)         // eqa <<= 16
	ORL(eq, eqa)              // eqa = eq | eqa
	MOVWLZX(b, eq)            // eq = b[i:i+2]
	MOVBLZX(b.Offset(2), eqb) // eqb = b[i+2:i+3]
	SHLL(U8(16), eqb)         // eqb <<= 16
	ORL(eq, eqb)              // eqb = eq | eqb
	XORL(eqa, eqb)            // eqb = eqa ^ eqb
	TESTL(U32(0xDFDFDF), eqb) // ZF = (mask & eqb)
}

func EQ4(a, b Mem, n Register) {
	eq := GP32()

	MOVL(a, eq)                // eq = a[i:i+4]
	XORL(b, eq)                // eq = b[i:i+4] ^ eq
	ADDQ(U8(4), a.Index)       // i += 4
	SUBQ(U8(4), n)             // n -= 4
	TESTL(U32(0xDFDFDFDF), eq) // ZF = (mask & eq)
}

func EQ8(a, b Mem, n, mask Register) {
	eq := GP64()

	MOVQ(a, eq)          // eq = a[i:i+8]
	XORQ(b, eq)          // eq = b[i:i+8] ^ eq
	ADDQ(U8(8), a.Index) // i += 8
	SUBQ(U8(8), n)       // n -= 8
	TESTQ(mask, eq)      // ZF = (mask & eq)
}

func EQ16(a, b Mem, n, mask Register) {
	EQAVX(a, b, n, XMM(), mask, 16)
}

func EQ32(a, b Mem, n, mask Register) {
	EQAVX(a, b, n, YMM(), mask, 32)
}

func EQ64(a, b Mem, n, mask Register) {
	EQAVXSIMD(a, b, n, mask, 2)
}

func EQ128(a, b Mem, n, mask Register) {
	EQAVXSIMD(a, b, n, mask, 4)
}

func EQAVX(a, b Mem, n, eq, mask Register, size uint8) {
	VMOVDQU(a, eq)          // eq = a[i:i+32]
	VPXOR(b, eq, eq)        // eq = b[i:i+32] ^ eq
	ADDQ(U8(size), a.Index) // i += size
	SUBQ(U8(size), n)       // n -= size
	VPTEST(mask, eq)        // ZF = (mask & eq)
}

func EQAVXSIMD(a, b Mem, n, mask Register, lanes int) {
	eq := GP32()
	and := make([]VecPhysical, 0)
	ymm := []VecPhysical{Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7, Y8, Y9, Y10, Y11, Y12, Y13, Y14, Y15}

	for i := 0; i < lanes; i++ {
		y0 := ymm[2*i]
		y1 := ymm[2*i+1]
		and = append(and, y0)

		VPAND(a.Offset(32*i), mask, y0) // y0 = a[i:i+32] & mask
		VPAND(b.Offset(32*i), mask, y1) // y1 = b[i:i+32] & mask
		VPCMPEQB(y1, y0, y0)            // y0 = y1 == y0
	}

	for len(and) > 1 {
		y0 := and[0]
		y1 := and[1]
		and = append(and[2:], y0)

		VPAND(y1, y0, y0)
	}

	VPMOVMSKB(and[0], eq)                // eq[0,1,2,...] = y0[0,8,16,...]
	ADDQ(Imm(32*uint64(lanes)), a.Index) // i += 32*lanes
	SUBQ(Imm(32*uint64(lanes)), n)       // n -= 32*lanes
	CMPL(eq, U32(0xFFFFFFFF))            // ZF = (eq == 0xFFFFFFFF)
}
