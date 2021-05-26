// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/x86"

	"fmt"

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
	tmp := GP32()

	CMPQ(n, bn.Addr)      // if len(a) != len(b):
	JNE(LabelRef("done")) //   return false
	XORQ(i, i)            // i = 0

	CMPQ(n, U8(16))                     // if n < 16:
	JB(LabelRef("init_x86"))            //   goto init_x86
	JumpIfFeature("init_avx", cpu.AVX2) // goto init_avx if supported

	Label("init_x86")

	cmp := GP32()
	av := GP32()
	bv := GP32()

	// Map to convert ASCII upper characters to lower case.
	lowerCase := Mem{Base: GP64(), Scale: 1}
	LEAQ(NewDataAddr(Symbol{Name: "github·com∕segmentio∕asm∕ascii·lowerCase"}, 0), lowerCase.Base)
	XORL(cmp, cmp)

	Label("cmp8")
	CMPQ(n, U8(8))       // if n < 0:
	JB(LabelRef("cmp7")) //   goto cmp7
	for i := 0; i < 8; i++ {
		MOVBLZX(a.Offset(i), av)             // av = a[i]
		MOVBLZX(b.Offset(i), bv)             // bv = b[i]
		MOVB(lowerCase.Idx(av, 1), av.As8()) // av = lowerCase[av]
		XORB(lowerCase.Idx(bv, 1), av.As8()) // av = lowerCase[bv] ^ av
		ORB(av.As8(), cmp.As8())             // cmp |= av
	}
	JNE(LabelRef("done")) // return false if ZF == 0
	ADDQ(Imm(8), a.Index) // i += 8
	SUBQ(Imm(8), n)       // n -= 8
	JMP(LabelRef("cmp8"))

	for i := 6; i >= 0; i-- {
		Label(fmt.Sprintf("cmp%d", i+1))
		next := "success"
		if i > 0 {
			next = fmt.Sprintf("cmp%d", i)
		}

		CMPQ(n, U8(i+1))                     // if n < i:
		JB(LabelRef(next))                   //   goto cmp${i-1}
		MOVBLZX(a.Offset(i), av)             // av = a[i]
		MOVBLZX(b.Offset(i), bv)             // bv = b[i]
		MOVB(lowerCase.Idx(av, 1), av.As8()) // av = lowerCase[av]
		XORB(lowerCase.Idx(bv, 1), av.As8()) // av = lowerCase[bv] ^ av
		ORB(av.As8(), cmp.As8())             // cmp |= av
	}

	Label("done")
	SETEQ(ret.Addr) // return ZF
	RET()           // ...

	Label("success")
	MOVB(U8(1), ret.Addr) // return true
	RET()                 // ...

	Label("init_avx")

	bit := VecBroadcast(U8(0x20), YMM()) // "case" bit
	msk := VecBroadcast(U8(0x1F), YMM()) // 0b10000000 - 'a'
	rng := VecBroadcast(U8(0x9A), YMM()) // 'z' - 'a' + 1 - 0x80 (overflowed 8-bits)
	one := VecBroadcast(U8(0x01), YMM()) // 1-bit for ANDing with comparison

	vec := NewVectorizer(12, func(l VectorLane) Register {
		v0 := l.Read(a)
		v1 := l.Read(b)
		v2 := l.Alloc()

		VXORPD(v0, v1, v1)     // calculate difference between a and b
		VPCMPEQB(bit, v1, v2)  // check if above difference is the 6th bit
		VORPD(bit, v0, v0)     // set the 6th bit for a
		VPADDB(msk, v0, v0)    // add 0x1f to each byte to set top bit for letters
		VPCMPGTB(v0, rng, v0)  // compare if not letter: v - 'a' < 'z' - 'a' + 1
		VPAND(v2, v0, v0)      // combine 6th-bit difference with letter range
		VPAND(one, v0, v0)     // merge test mask
		VPSLLW(Imm(5), v0, v0) // shift into case bit position
		VPCMPEQB(v1, v0, v0)   // compare original difference with case-only difference

		return v0
	}).Reduce(ReduceAnd) // merge all comparisons together

	cmpAVX := func(spec Spec, lanes int, incr bool) {
		sz := int(spec.Size())
		out := vec.Compile(spec, lanes)[0] // [compare sz*lanes bytes]
		if incr {
			ADDQ(U8(sz*lanes), a.Index) // i += sz*lanes
			SUBQ(U8(sz*lanes), n)       // n -= sz*lanes
		}
		VPMOVMSKB(out, tmp)                 // tmp[0,1,2,...] = y0[0,8,16,...]
		XORL(U32(^uint32(0)>>(32-sz)), tmp) // ZF = (tmp == 0xFFFFFFFF)
	}

	Label("cmp128")
	CMPQ(n, U8(128))        // if n < 128:
	JB(LabelRef("cmp64"))   //   goto cmp64
	cmpAVX(S256, 4, true)   // ZF = [compare 128 bytes]
	JNE(LabelRef("done"))   // return if ZF == 0
	JMP(LabelRef("cmp128")) // loop cmp128

	Label("cmp64")
	CMPQ(n, U8(64))       // if n < 64:
	JB(LabelRef("cmp32")) //   goto cmp32
	cmpAVX(S256, 2, true) // ZF = [compare 64 bytes]
	JNE(LabelRef("done")) // return if ZF == 0

	Label("cmp32")
	CMPQ(n, U8(32))       // if n < 32:
	JB(LabelRef("cmp16")) //   goto cmp16
	cmpAVX(S256, 1, true) // ZF = [compare 32 bytes]
	JNE(LabelRef("done")) // return if ZF == 0

	Label("cmp16")

	// Convert YMM masks to XMM
	bit = bit.(Vec).AsX()
	msk = msk.(Vec).AsX()
	rng = rng.(Vec).AsX()
	one = one.(Vec).AsX()

	CMPQ(n, U8(16))           // if n <= 16:
	JLE(LabelRef("cmp_tail")) //   goto cmp_tail
	cmpAVX(S128, 1, true)     // ZF = [compare 16 bytes]
	JNE(LabelRef("done"))     // return if ZF == 0

	Label("cmp_tail")
	// At this point, we have <= 16 bytes to compare, but we know the total input
	// is >= 16 bytes. Move the pointer to the *last* 16 bytes of the input so we
	// can skip the fallback.
	SUBQ(Imm(16), n)       // n -= 16
	ADDQ(n, a.Index)       // i += n
	cmpAVX(S128, 1, false) // ZF = [compare 16 bytes]
	JMP(LabelRef("done"))  // return ZF

	Generate()
}
