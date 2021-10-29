//go:build ignore
// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/x86"

	"github.com/segmentio/asm/cpu"
)

func init() {
	ConstraintExpr("!purego")
}

func main() {
	TEXT("ValidString", NOSPLIT, "func(s string) bool")
	Doc("ValidString returns true if s contains only ASCII characters.")

	p := Mem{Base: Load(Param("s").Base(), GP64())}
	n := Load(Param("s").Len(), GP64())
	ret, _ := ReturnIndex(0).Resolve()

	v := GP32()
	vl := GP32()
	maskG := GP64()

	MOVQ(U64(0x8080808080808080), maskG) // maskG = 0x8080808080808080

	CMPQ(n, U8(16))      // if n < 16:
	JB(LabelRef("cmp8")) //   goto cmp8
	JumpIfFeature("init_avx", cpu.AVX2)

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

	Label("init_avx")

	maskY := VecBroadcast(maskG, YMM())
	maskX := maskY.(Vec).AsX()

	vec := NewVectorizer(15, func(l VectorLane) Register {
		r := l.Alloc()
		VMOVDQU(l.Offset(p), r)
		VPOR(l.Offset(p), r, r)
		return r
	}).Reduce(ReduceOr)

	Label("cmp256")
	CMPQ(n, U32(256))                      // if n < 256:
	JB(LabelRef("cmp128"))                 //   goto cmp128
	VPTEST(vec.Compile(S256, 4)[0], maskY) // if (OR & maskY) != 0:
	JNZ(LabelRef("invalid"))               //   return false
	ADDQ(U32(256), p.Base)                 // p += 256
	SUBQ(U32(256), n)                      // n -= 256
	JMP(LabelRef("cmp256"))                // loop cmp256

	Label("cmp128")
	CMPQ(n, U8(128))                       // if n < 128:
	JB(LabelRef("cmp64"))                  //   goto cmp64
	VPTEST(vec.Compile(S256, 2)[0], maskY) // if (OR & maskY) != 0:
	JNZ(LabelRef("invalid"))               //   return false
	ADDQ(U8(128), p.Base)                  // p += 128
	SUBQ(U8(128), n)                       // n -= 128
	JMP(LabelRef("cmp64"))                 // goto cmp64

	Label("cmp64")
	CMPQ(n, U8(64))                        // if n < 64:
	JB(LabelRef("cmp32"))                  //   goto cmp32
	VPTEST(vec.Compile(S256, 1)[0], maskY) // if (OR & maskY) != 0:
	JNZ(LabelRef("invalid"))               //   return false
	ADDQ(U8(64), p.Base)                   // p += 64
	SUBQ(U8(64), n)                        // n -= 64

	Label("cmp32")
	CMPQ(n, U8(32))          // if n < 32:
	JB(LabelRef("cmp16"))    //   goto cmp16
	VPTEST(p, maskY)         // if (p[0:32] & maskY) != 0:
	JNZ(LabelRef("invalid")) //   return false
	ADDQ(U8(32), p.Base)     // p += 32
	SUBQ(U8(32), n)          // n -= 32

	Label("cmp16")
	CMPQ(n, U8(16))           // if n <= 16:
	JLE(LabelRef("cmp_tail")) //   goto cmp_tail
	VPTEST(p, maskX)          // if (p[0:16] & maskX) != 0:
	JNZ(LabelRef("invalid"))  //   return false
	ADDQ(U8(16), p.Base)      // p += 16
	SUBQ(U8(16), n)           // n -= 16

	Label("cmp_tail")
	// At this point, we have <= 16 bytes to compare, but we know the total input
	// is >= 16 bytes. Move the pointer to the *last* 16 bytes of the input so we
	// can skip the fallback.
	SUBQ(Imm(16), n)      // n -= 16
	ADDQ(n, p.Base)       // p += n
	VPTEST(p, maskX)      // ZF = (p[0:16] & maskX) == 0
	JMP(LabelRef("done")) // return ZF

	Generate()
}
