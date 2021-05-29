// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/x86"
)

func main() {
	TEXT("containsByteAVX2", NOSPLIT, "func(haystack []byte, needle byte) bool")

	haystack := Load(Param("haystack").Base(), GP64())
	length := Load(Param("haystack").Len(), GP64())
	end := GP64()
	LEAQ(Mem{Base: haystack, Index: length, Scale: 1}, end)

	needle := Load(Param("needle"), GP8())
	needleVec := VecBroadcast(needle, YMM())

	ret, _ := ReturnIndex(0).Resolve()
	MOVB(U8(0), ret.Addr)

	zero := YMM()
	VPXOR(zero, zero, zero)

	vec := NewVectorizer(15, func(l VectorLane) Register {
		r := l.Alloc()
		VPCMPEQB(l.Offset(Mem{Base: haystack}), needleVec, r)
		return r
	}).Reduce(ReduceOr)

	next := GP64()
	MOVQ(haystack, next)

	Label("avx2_loop")
	const unroll = 8
	ADDQ(U32(32 * unroll), next)
	CMPQ(next, end)
	JA(LabelRef("tail_loop"))
	VPTEST(vec.Compile(S256, unroll)[0], zero)
	JCC(LabelRef("found"))
	MOVQ(next, haystack)
	JMP(LabelRef("avx2_loop"))

	// Slow/tail loop.
	// FIXME: unroll, or maybe have blocks for specific sizes of remaining input.
	Label("tail_loop")
	CMPQ(haystack, end)
	JE(LabelRef("done"))
	value := GP8()
	MOVB(Mem{Base: haystack}, value)
	CMPB(needle, value)
	JE(LabelRef("found"))
	ADDQ(Imm(1), haystack)
	JMP(LabelRef("tail_loop"))

	Label("found")
	MOVB(U8(1), ret.Addr)
	Label("done")
	RET()
	Generate()
}
