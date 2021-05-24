// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/x86"

	"github.com/segmentio/asm/cpu"
)

const x = false

func main() {
	TEXT("ValidPrintString", NOSPLIT, "func(s string) bool")
	Doc("ValidPrintString returns true if s contains only printable ASCII characters.")

	p := Mem{Base: Load(Param("s").Base(), GP64())}
	n := Load(Param("s").Len(), GP64())
	ret, _ := ReturnIndex(0).Resolve()

	m1 := GP64()
	m2 := GP64()
	m3 := GP64()
	val := GP32()
	tmp := GP32()

	CMPQ(n, U8(16))                // if n < 16:
	JB(LabelRef("init"))           //   goto init
	JumpIfFeature("avx", cpu.AVX2) // goto avx if supported

	Label("init")
	CMPQ(n, U8(8))       // if n < 8:
	JB(LabelRef("cmp4")) //   goto cmp4
	MOVQ(U64(0xDFDFDFDFDFDFDFE0), m1)
	MOVQ(U64(0x0101010101010101), m2)
	MOVQ(U64(0x8080808080808080), m3)

	Label("cmp8")
	valid8(p, n, m1, m2, m3) // ZF = [compare 8 bytes]
	JNE(LabelRef("done"))    // return ZF if ZF == 0
	CMPQ(n, U8(8))           // if n < 8:
	JB(LabelRef("cmp4"))     //   goto cmp4
	JMP(LabelRef("cmp8"))    // loop cmp8

	Label("cmp4")
	CMPQ(n, U8(4))        // if n < 4:
	JB(LabelRef("cmp3"))  //   goto cmp3
	valid4(p, n)          // ZF = [compare 4 bytes]
	JNE(LabelRef("done")) // return ZF if ZF == 0

	Label("cmp3")
	CMPQ(n, U8(3))            // if n < 3:
	JB(LabelRef("cmp2"))      //   goto cmp2
	MOVWLZX(p, tmp)           // tmp = p[0:2]
	MOVBLZX(p.Offset(2), val) // val = p[2:3]
	SHLL(U8(16), val)         // val <<= 16
	ORL(tmp, val)             // val = tmp | val
	ORL(U32(0x20000000), val) // val = 0x20000000 | val
	JMP(LabelRef("final"))

	Label("cmp2")
	CMPQ(n, U8(2))            // if n < 2:
	JB(LabelRef("cmp1"))      //   goto cmp1
	MOVWLZX(p, val)           // val = p[0:2]
	ORL(U32(0x20200000), val) // val = 0x20200000 | val
	JMP(LabelRef("final"))

	Label("cmp1")
	CMPQ(n, U8(0))            // if n == 0:
	JE(LabelRef("done"))      //   return true
	MOVBLZX(p, val)           // val = p[0:1]
	ORL(U32(0x20202000), val) // val = 0x20202000 | val

	Label("final")
	setup4(val)                 // [update val register]
	TESTL(U32(0x80808080), val) // ZF = (0x80808080 & val) == 0

	Label("done")
	SETEQ(ret.Addr) // return ZF
	RET()           // ...

	Label("avx")

	minY := VecBroadcast(U8(0x1F), YMM())
	maxY := VecBroadcast(U8(0x7E), YMM())
	minX := minY.(Vec).AsX()
	maxX := maxY.(Vec).AsX()

	Label("cmp128")
	CMPQ(n, U8(128))                    // if n < 128:
	JB(LabelRef("cmp64"))               //   goto cmp64
	validAVX(p, n, minY, maxY, 4, S256) // ZF = [compare 128 bytes]
	JNE(LabelRef("done"))               // return if ZF == 0
	JMP(LabelRef("cmp128"))             // loop cmp128

	Label("cmp64")
	CMPQ(n, U8(64))                     // if n < 64:
	JB(LabelRef("cmp32"))               //   goto cmp32
	validAVX(p, n, minY, maxY, 2, S256) // ZF = [compare 64 bytes]
	JNE(LabelRef("done"))               // return ZF if ZF == 0

	Label("cmp32")
	CMPQ(n, U8(32))                     // if n < 32:
	JB(LabelRef("cmp16"))               //   goto cmp16
	validAVX(p, n, minY, maxY, 1, S256) // ZF = [compare 32 bytes]
	JNE(LabelRef("done"))               // return ZF if ZF == 0

	Label("cmp16")
	CMPQ(n, U8(16))                     // if n < 16:
	JB(LabelRef("init"))                //   goto init
	validAVX(p, n, minX, maxX, 1, S128) // ZF = [compare 16 bytes]
	JNE(LabelRef("done"))               // return ZF if ZF == 0

	CMPQ(n, U8(0))        // if n == 0:
	JE(LabelRef("done"))  //   goto done
	JMP(LabelRef("init")) // gpto init

	Generate()
}

func valid4(p Mem, n Register) {
	val := GP32()

	MOVL(p, val)                // val = p[0:4]
	setup4(val)                 // [update val register]
	ADDQ(U8(4), p.Base)         // p += 4
	SUBQ(U8(4), n)              // n -= 4
	TESTL(U32(0x80808080), val) // ZF = (0x80808080 & val) == 0
}

func setup4(val Register) {
	nval := GP32()
	tmp1 := GP32()
	tmp2 := GP32()

	MOVL(val, nval)                              // nval = val
	LEAL(Mem{Disp: 0xDFDFDFE0, Base: val}, tmp1) // tmp1 = val + 0xDFDFDFE0
	NOTL(nval)                                   // nval = ^nval
	ANDL(nval, tmp1)                             // tmp1 = nval & tmp1
	LEAL(Mem{Disp: 0x01010101, Base: val}, tmp2) // tmp2 = val + 0x01010101
	ORL(tmp2, val)                               // val = val | tmp2
	ORL(tmp1, val)                               // val = val | tmp1
}

func valid8(p Mem, n, m1, m2, m3 Register) {
	val := GP64()
	nval := GP64()
	tmp1 := GP64()
	tmp2 := GP64()

	MOVQ(p, val)                                    // val = p[0:8]
	MOVQ(val, nval)                                 // nval = val
	LEAQ(Mem{Base: val, Index: m1, Scale: 1}, tmp1) // tmp1 = val + m1
	NOTQ(nval)                                      // nval = ^nval
	ANDQ(nval, tmp1)                                // tmp1 = nval & tmp1
	LEAQ(Mem{Base: val, Index: m2, Scale: 1}, tmp2) // tmp2 = val + m2
	ORQ(tmp2, val)                                  // val = val | tmp2
	ORQ(tmp1, val)                                  // val = val | tmp1
	ADDQ(U8(8), p.Base)                             // p += 8
	SUBQ(U8(8), n)                                  // n -= 8
	TESTQ(m3, val)                                  // ZF = (m3 & val) == 0
}

func validAVX(p Mem, n, min, max Register, lanes int, s Spec) {
	msk := GP32()
	out := make([]VecPhysical, 0)
	vec := VecList(s, 14)
	sz := int(s.Size())

	for i := 0; i < lanes; i++ {
		v0 := vec[2*i]

		VMOVDQU(p.Offset(sz*i), v0) // v0 = p[i*sz:i*sz+sz]
	}

	for i := 0; i < lanes; i++ {
		v0 := vec[2*i]
		v1 := vec[2*i+1]
		out = append(out, v0)

		VPCMPGTB(min, v0, v1) // v1 = bytes that are greater than the min-1 (i.e. valid at lower end)
		VPCMPGTB(max, v0, v0) // v0 = bytes that are greater than the max (i.e. invalid at upper end)
		VPANDN(v1, v0, v0)    // y2 & ~y3 mask should be full unless there's an invalid byte
	}

	for len(out) > 1 {
		v0 := out[0]
		v1 := out[1]
		out = append(out[2:], v0)

		VPAND(v1, v0, v0)
	}

	ADDQ(Imm(uint64(sz*lanes)), p.Base) // p += sz*lanes
	SUBQ(Imm(uint64(sz*lanes)), n)      // n -= sz*lanes
	VPMOVMSKB(out[0], msk)              // msk[0,1,2,...] = v0[0,8,16,...]
	XORL(U32(^uint32(0)>>(32-sz)), msk) // ZF = (msk == 0xFFFFFFFF)
}
