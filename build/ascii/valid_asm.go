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
	TEXT("ValidString", NOSPLIT, "func(s string) bool")
	Doc("ValidString returns true if s contains only ASCII characters.")

	p := Mem{Base: Load(Param("s").Base(), GP64())}
	n := Load(Param("s").Len(), GP64())
	ret, _ := ReturnIndex(0).Resolve()

	v := GP32()
	vl := GP32()
	maskG := GP64()
	maskY := YMM()
	maskX := maskY.AsX() // use the lower half of maskY

	MOVQ(U64(0x8080808080808080), maskG) // maskG = 0x8080808080808080

	JumpUnlessFeature("cmp8", cpu.AVX2)

	PINSRQ(Imm(0), maskG, maskX) // maskX[0:8] = maskG
	VPBROADCASTQ(maskX, maskY)   // maskY[0:32] = [maskX[0:8],maskX[0:8],maskX[0:8],maskX[0:8]]

	// Moving the 256-byte scanning helps the branch predictor for small inputs
	CMPQ(n, U32(256))       // if n >= 256:
	JNB(LabelRef("cmp256")) //   goto cmp256

	Label("cmp64")
	CMPQ(n, U8(64))            // if n < 64:
	JB(LabelRef("cmp32"))      //   goto cmp32
	VMOVDQU(p, Y0)             // Y0 = p[0:32]
	VPOR(p.Offset(32), Y0, Y0) // Y0 = p[32:64] | Y0
	VPTEST(Y0, maskY)          // if (Y0 & maskY) != 0:
	JNZ(LabelRef("invalid"))   //   return false
	ADDQ(U8(64), p.Base)       // p += 64
	SUBQ(U8(64), n)            // n -= 64

	Label("cmp32")
	CMPQ(n, U8(32))          // if n < 32:
	JB(LabelRef("cmp16"))    //   goto cmp16
	VPTEST(p, maskY)         // if (p[0:32] & maskY) != 0:
	JNZ(LabelRef("invalid")) //   return false
	ADDQ(U8(32), p.Base)     // p += 32
	SUBQ(U8(32), n)          // n -= 32

	Label("cmp16")
	CMPQ(n, U8(16))          // if n < 16:
	JB(LabelRef("cmp8"))     //   goto cmp8
	VPTEST(p, maskX)         // if (p[0:16] & maskX) != 0:
	JNZ(LabelRef("invalid")) //   return false
	ADDQ(U8(16), p.Base)     // p += 16
	SUBQ(U8(16), n)          // n -= 16

	Label("cmp8")
	CMPQ(n, U8(8))           // if n < 8:
	JB(LabelRef("cmp4"))     //   goto cmp4
	TESTQ(maskG, p)          // if (p[0:8] & 0x8080808080808080) != 0:
	JNZ(LabelRef("invalid")) //   return false
	ADDQ(U8(8), p.Base)      // p += 8
	SUBQ(U8(8), n)           // n -= 8
	JMP(LabelRef("cmp8"))    // loop cmp8

	Label("cmp4")
	CMPQ(n, U8(4))            // if n < 4:
	JB(LabelRef("cmp3"))      //   goto cmp3
	TESTL(U32(0x80808080), p) // if (p[0:4] & 0x80808080) != 0:
	JNZ(LabelRef("invalid"))  //   return false
	ADDQ(U8(4), p.Base)       // p += 4
	SUBQ(U8(4), n)            // n -= 4

	Label("cmp3")
	CMPQ(n, U8(3))            // if n < 3:
	JB(LabelRef("cmp2"))      //   goto cmp2
	MOVWLZX(p, vl)            // vl = p[i:i+2]
	MOVBLZX(p.Offset(2), v)   // v = p[i+2:i+3]
	SHLL(U8(16), v)           // v <<= 16
	ORL(vl, v)                // v = vl | v
	TESTL(U32(0x80808080), v) // ZF = (v & 0x80808080) == 0
	JMP(LabelRef("done"))     // return ZF

	Label("cmp2")
	CMPQ(n, U8(2))        // if n < 2:
	JB(LabelRef("cmp1"))  //   goto cmp1
	TESTW(U16(0x8080), p) // ZF = (p[0:2] & 0x8080) == 0
	JMP(LabelRef("done")) // return ZF

	Label("cmp1")
	CMPQ(n, U8(0))       // if n == 0:
	JE(LabelRef("done")) //   return true
	TESTB(U8(0x80), p)   // ZF = (p[0:1] & 0x80) == 0

	Label("done")
	SETEQ(ret.Addr) // return ZF
	RET()           // ...

	Label("invalid")
	MOVB(U8(0), ret.Addr)
	RET()

	Label("cmp256")
	VMOVDQU(p.Offset(0), Y0)    // Y0 = p[0:32]
	VPOR(p.Offset(32), Y0, Y0)  // Y0 = p[32:64] | Y0
	VMOVDQU(p.Offset(64), Y1)   // Y1 = p[64:96]
	VPOR(p.Offset(96), Y1, Y1)  // Y1 = p[96:126] | Y1
	VMOVDQU(p.Offset(128), Y2)  // Y2 = p[128:160]
	VPOR(p.Offset(160), Y2, Y2) // Y2 = p[160:192] | Y2
	VMOVDQU(p.Offset(192), Y3)  // Y3 = p[192:224]
	VPOR(p.Offset(224), Y3, Y3) // Y3 = p[224:256] | Y3
	VPOR(Y1, Y0, Y0)            // Y0 = Y1 | Y0
	VPOR(Y3, Y2, Y2)            // Y2 = Y3 | Y2
	VPOR(Y2, Y0, Y0)            // Y0 = Y2 | Y0
	VPTEST(Y0, maskY)           // if (Y0 & maskY) != 0:
	JNZ(LabelRef("invalid"))    //   return false
	ADDQ(U32(256), p.Base)      // p += 256
	SUBQ(U32(256), n)           // n -= 256
	CMPQ(n, U32(256))           // if n < 256:
	JB(LabelRef("cmp128"))      //   goto cmp128
	JMP(LabelRef("cmp256"))     // loop cmp256

	Label("cmp128")
	CMPQ(n, U8(128))           // if n < 128:
	JB(LabelRef("cmp64"))      //   goto cmp64
	VMOVDQU(p.Offset(0), Y0)   // Y0 = p[0:32]
	VPOR(p.Offset(32), Y0, Y0) // Y0 = p[32:64] | Y0
	VMOVDQU(p.Offset(64), Y1)  // Y1 = p[64:96]
	VPOR(p.Offset(96), Y1, Y1) // Y1 = p[96:126] | Y1
	VPOR(Y1, Y0, Y0)           // Y0 = Y1 | Y0
	VPTEST(Y0, maskY)          // if (Y0 & maskY) != 0:
	JNZ(LabelRef("invalid"))   //   return false
	ADDQ(U8(128), p.Base)      // p += 128
	SUBQ(U8(128), n)           // n -= 128
	JMP(LabelRef("cmp64"))     // goto cmp64

	Generate()
}
