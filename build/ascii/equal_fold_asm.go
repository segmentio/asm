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
	XORQ(i, i)            // i = 0

	CMPQ(n, U8(16))                // if n < 16:
	JB(LabelRef("cmp8"))           //   goto cmp8
	JumpIfFeature("avx", cpu.AVX2) // goto avx if supported

	Label("cmp8")
	// TODO: add fallback
	t := GP64()
	XORQ(t, t)

	Label("done")
	SETEQ(ret.Addr) // return ZF
	RET()           // ...

	Label("avx")

	maskY := [...]Register{
		VecBroadcast(U8(0x20), YMM()), // "case" bit
		VecBroadcast(U8(0x1F), YMM()), // 0b10000000 - 'a'
		VecBroadcast(U8(0x9A), YMM()), // 'z' - 'a' + 1 - 0x80 (overflowed 8-bits)
		VecBroadcast(U8(0x01), YMM()), // 1-bit for testing
	}

	maskX := [...]Register{
		maskY[0].(Vec).AsX(),
		maskY[1].(Vec).AsX(),
		maskY[2].(Vec).AsX(),
		maskY[3].(Vec).AsX(),
	}

	Label("cmp128")
	CMPQ(n, U8(128))                  // if n < 128:
	JB(LabelRef("cmp64"))             //   goto cmp64
	equalAVX(a, b, n, maskY, 4, S256) // ZF = [compare 128 bytes]
	JNE(LabelRef("done"))             // return if ZF == 0
	JMP(LabelRef("cmp128"))           // loop cmp128

	Label("cmp64")
	CMPQ(n, U8(64))                   // if n < 64:
	JB(LabelRef("cmp32"))             //   goto cmp32
	equalAVX(a, b, n, maskY, 2, S256) // ZF = [compare 64 bytes]
	JNE(LabelRef("done"))             // return ZF if ZF == 0

	Label("cmp32")
	CMPQ(n, U8(32))                   // if n < 32:
	JB(LabelRef("cmp16"))             //   goto cmp16
	equalAVX(a, b, n, maskY, 1, S256) // ZF = [compare 32 bytes]
	JNE(LabelRef("done"))             // return ZF if ZF == 0

	Label("cmp16")
	CMPQ(n, U8(16))                   // if n < 16:
	JB(LabelRef("cmp8"))              //   goto cmp8
	equalAVX(a, b, n, maskX, 1, S128) // ZF = [compare 16 bytes]
	JNE(LabelRef("done"))             // return ZF if ZF == 0

	CMPQ(n, U8(0))        // if n == 0:
	JE(LabelRef("done"))  //   goto done
	JMP(LabelRef("cmp8")) // gpto cmp8

	Generate()
}

func equalAVX(a, b Mem, n Register, mask [4]Register, lanes int, s Spec) {
	msk := GP32()
	out := make([]VecPhysical, 0)
	vec := VecList(s, 12)
	sz := int(s.Size())

	for i := 0; i < lanes; i++ {
		v0 := vec[3*i]
		v1 := vec[3*i+1]
		v2 := vec[3*i+2]
		out = append(out, v0)

		VLDDQU(a.Offset(sz*i), v0) // load 32 bytes from a
		VLDDQU(b.Offset(sz*i), v1) // load 32 bytes from b
		VXORPD(v0, v1, v1)         // calculate difference between a and b
		VPCMPEQB(mask[0], v1, v2)  // check if above difference is the 6th bit
		VORPD(mask[0], v0, v0)     // set the 6th bit for a
		VPADDB(mask[1], v0, v0)    // add 0x1f to each byte to set top bit for letters
		VPCMPGTB(v0, mask[2], v0)  // compare if not letter: v - 'a' < 'z' - 'a' + 1
		VPAND(v2, v0, v0)          // combine 6th-bit difference with letter range
		VPAND(mask[3], v0, v0)     // merge test mask
		VPSLLW(Imm(5), v0, v0)     // shift into case bit position
		VPCMPEQB(v1, v0, v0)       // compare original difference with case-only difference
	}

	for len(out) > 1 {
		v0 := out[0]
		v1 := out[1]
		out = append(out[2:], v0)

		VPAND(v1, v0, v0) // merge all comparisons together
	}

	ADDQ(Imm(uint64(sz*lanes)), a.Index) // i += sz*lanes
	SUBQ(Imm(uint64(sz*lanes)), n)       // n -= sz*lanes
	VPMOVMSKB(out[0], msk)               // msk[0,1,2,...] = y0[0,8,16,...]
	XORL(U32(^uint32(0)>>(32-sz)), msk)  // ZF = (msk == 0xFFFFFFFF)
}
