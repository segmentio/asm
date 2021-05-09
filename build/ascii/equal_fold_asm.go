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

	a := Load(Param("a").Base(), GP64())
	n := Load(Param("a").Len(), GP64())
	b := Load(Param("b").Base(), GP64())
	l, err := Param("b").Len().Resolve()
	if err != nil {
		panic(err)
	}

	ret, err := ReturnIndex(0).Resolve()
	if err != nil {
		panic(err)
	}

	CMPQ(n, l.Addr)       // if a_len != b_len:
	JNE(LabelRef("done")) //   return false

	eqy := YMM()
	eqx := XMM()
	eq := GP64()
	eqa := GP32()
	eqb := GP32()

	mask64 := GP64()
	mask256 := YMM()

	// Use index for byte position. We have plenty of registers, and it saves an
	// ADD operation as the memory index is the same for both a and b.
	i := GP64()
	p := Mem{Base: a, Index: i, Scale: 1}
	q := Mem{Base: b, Index: i, Scale: 1}

	XORQ(i, i)
	MOVQ(U64(0xDFDFDFDFDFDFDFDF), mask64)

	JumpUnlessFeature("eq8", cpu.AVX2)

	PINSRQ(Imm(0), mask64, mask256.AsX())
	VPBROADCASTQ(mask256.AsX(), mask256)

	Label("eq64")
	CMPQ(n, U8(64))                // if n < 64:
	JB(LabelRef("eq32"))           //   goto eq32
	SIMDEQ(p, q, n, i, mask256, 2) // [compare 64 bytes]
	JMP(LabelRef("eq64"))          // loop eq64

	Label("eq32")
	CMPQ(n, U8(32))       // if n < 32:
	JB(LabelRef("eq16"))  //   goto eq16
	VMOVDQU(p, eqy)       // eqy = a[i:i+32]
	VPXOR(q, eqy, eqy)    // eqy = b[i:i+32] ^ eqy
	ADDQ(U8(32), i)       // i += 32
	SUBQ(U8(32), n)       // n -= 32
	VPTEST(mask256, eqy)  // if !(mask256 & eqy):
	JNE(LabelRef("done")) //   return false

	Label("eq16")
	CMPQ(n, U8(16))            // if n < 16:
	JB(LabelRef("eq8"))        //   goto eq8
	VMOVDQU(p, eqx)            // eqx = a[i:i+16]
	VPXOR(q, eqx, eqx)         // eqx = b[i:i+16] ^ eqx
	ADDQ(U8(16), i)            // i += 16
	SUBQ(U8(16), n)            // n -= 16
	VPTEST(mask256.AsX(), eqx) // if !(mask128 & eqx):
	JNE(LabelRef("done"))      //   return false

	Label("eq8")
	CMPQ(n, U8(8))        // if n < 8:
	JB(LabelRef("eq4"))   //   goto eq4
	MOVQ(p, eq)           // eq = a[i:i+8]
	XORQ(q, eq)           // eq = b[i:i+8] ^ eq
	ADDQ(U8(8), i)        // i += 8
	SUBQ(U8(8), n)        // n -= 8
	TESTQ(mask64, eq)     // if !(mask64 & eq):
	JNE(LabelRef("done")) //   return false
	JMP(LabelRef("eq8"))  // loop eq8

	Label("eq4")
	CMPQ(n, U8(4))                    // if n < 4:
	JB(LabelRef("eq3"))               //   goto eq3
	MOVL(p, eq.As32())                // eq = a[i:i+4]
	XORL(q, eq.As32())                // eq = b[i:i+4] ^ eq
	ADDQ(U8(4), i)                    // i += 4
	SUBQ(U8(4), n)                    // n -= 4
	TESTL(U32(0xDFDFDFDF), eq.As32()) // if !(mask & eq):
	JNE(LabelRef("done"))             //   return false

	Label("eq3")
	CMPQ(n, U8(3))            // if n < 3:
	JB(LabelRef("eq2"))       //   goto eq2
	MOVWLZX(p, eq.As32())     // eq = a[i:i+2]
	MOVBLZX(p.Offset(2), eqa) // eqa = a[i+2:i+3]
	SHLL(U8(16), eqa)         // eqa <<= 16
	ORL(eq.As32(), eqa)       // eqa = eq | eqa
	MOVWLZX(q, eq.As32())     // eq = b[i:i+2]
	MOVBLZX(q.Offset(2), eqb) // eqb = b[i+2:i+3]
	SHLL(U8(16), eqb)         // eqb <<= 16
	ORL(eq.As32(), eqb)       // eqb = eq | eqb
	XORL(eqa, eqb)            // eqb = eqa ^ eqb
	TESTL(U32(0xDFDFDF), eqb) // return (mask & eqb)
	JMP(LabelRef("done"))     // ...

	Label("eq2")
	CMPQ(n, U8(2))                // if n < 2:
	JB(LabelRef("eq1"))           //   goto eq1
	MOVW(p, eq.As16())            // eq = a[i:i+2]
	XORW(q, eq.As16())            // eq = b[i:i+2] ^ eq
	TESTW(U16(0xDFDF), eq.As16()) // return (mask & eq)
	JMP(LabelRef("done"))         // ...

	Label("eq1")
	CMPQ(n, U8(0))            // if n == 0:
	JE(LabelRef("done"))      //   return true
	MOVB(p, eq.As8())         // eq = a[i:i+1]
	XORB(q, eq.As8())         // eq = b[i:i+1] ^ eq
	TESTB(U8(0xDF), eq.As8()) // return (mask & eq)

	Label("done")
	SETEQ(ret.Addr)
	RET()

	Generate()
}

func SIMDEQ(p, q Mem, n, i Register, mask256 VecVirtual, lanes int) {
	eq := GP32()
	and := make([]VecPhysical, 0)
	ymm := []VecPhysical{Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7, Y8, Y9, Y10, Y11, Y12, Y13, Y14, Y15}

	for i := 0; i < lanes; i++ {
		y0 := ymm[2*i]
		y1 := ymm[2*i+1]
		and = append(and, y0)

		VPAND(p.Offset(32*i), mask256, y0) // y0 = a[i:i+32] & mask256
		VPAND(q.Offset(32*i), mask256, y1) // y1 = b[i:i+32] & mask256
		VPCMPEQB(y1, y0, y0)               // y0 = y1 == y0
	}

	for len(and) > 1 {
		y0 := and[0]
		y1 := and[1]
		and = append(and[2:], y0)

		VPAND(y1, y0, y0)
	}

	VPMOVMSKB(and[0], eq)          // eq[0,1,2,...] = y0[0,8,16,...]
	ADDQ(Imm(32*uint64(lanes)), i) // i += 32*lanes
	SUBQ(Imm(32*uint64(lanes)), n) // n -= 32*lanes
	CMPL(eq, U32(0xFFFFFFFF))      // if eq != 0xFFFFFFFF:
	JNE(LabelRef("done"))          //   return false
}
