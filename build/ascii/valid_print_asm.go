// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func main() {
	TEXT("ValidPrintString", NOSPLIT, "func(s string) bool")
	Doc("ValidPrintString returns true if s contains only printable ASCII characters.")

	p := Mem{Base: Load(Param("s").Base(), GP64())}
	n := Load(Param("s").Len(), GP64())
	ret, _ := ReturnIndex(0).Resolve()
	val := GP32()
	tmp := GP32()

	m1 := GP64()
	m2 := GP64()
	m3 := GP64()
	MOVQ(U64(0xDFDFDFDFDFDFDFE0), m1)
	MOVQ(U64(0x0101010101010101), m2)
	MOVQ(U64(0x8080808080808080), m3)

	Label("cmp8")
	CMPQ(n, U8(8))           // if n < 8:
	JB(LabelRef("cmp4"))     //   goto cmp4
	valid8(p, n, m1, m2, m3) // ZF = [compare 8 bytes]
	JNE(LabelRef("done"))    // return ZF if ZF == 0
	JMP(LabelRef("cmp8"))    // loop cmp8

	Label("cmp4")
	CMPQ(n, U8(4))        // if n < 4:
	JB(LabelRef("cmp3"))  //   goto cmp3
	valid4(p, n)          // ZF = [compare 4 bytes]
	JNE(LabelRef("done")) // return ZF if ZF == 0

	Label("cmp3")
	CMPQ(n, U8(3))            // if n < 3:
	JB(LabelRef("cmp2"))      //   goto cmp2
	MOVWLZX(p, tmp)           // tmp = a[i:i+2]
	MOVBLZX(p.Offset(2), val) // val = a[i+2:i+3]
	SHLL(U8(16), val)         // val <<= 16
	ORL(tmp, val)             // val = tmp | val
	ORL(U32(0x20000000), val) // val = 0x20000000 | val
	JMP(LabelRef("final"))

	Label("cmp2")
	CMPQ(n, U8(2))            // if n < 2:
	JB(LabelRef("cmp1"))      //   goto cmp1
	MOVWLZX(p, val)           // val = a[i:i+2]
	ORL(U32(0x20200000), val) // val = 0x20200000 | val
	JMP(LabelRef("final"))

	Label("cmp1")
	CMPQ(n, U8(0))            // if n == 0:
	JE(LabelRef("done"))      //   return true
	MOVBLZX(p, val)           // val = a[i:i+2]
	ORL(U32(0x20202000), val) // val = 0x20202000 | val

	Label("final")
	setup4(val)                 // [update val register]
	TESTL(U32(0x80808080), val) // ZF = (0x80808080 & val) == 0

	Label("done")
	SETEQ(ret.Addr) // return ZF
	RET()           // ...

	Generate()
}

func valid4(p Mem, n Register) {
	val := GP32()

	MOVL(p, val)                // val = a[i:i+8]
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

	MOVQ(p, val)                                    // val = a[i:i+8]
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
