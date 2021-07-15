// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/segmentio/asm/build/internal/x86"
)

func main() {
	TEXT("Despace", NOSPLIT, "func(data []byte)")
	Doc("remove spaces (in-place) from string bytes (UTF-8 or ASCII)")
	ptr := Mem{Base: Load(Param("data").Base(), GP64())}
	len := Load(Param("data").Len(), GP64())

	idx := GP64()
	XORQ(idx, idx)
	spaces := VecBroadcast(U8(' '), YMM())
	newline := VecBroadcast(U8('\n'), YMM())
	carriage := VecBroadcast(U8('\r'), YMM())

	Label("avx2_loop")
	next := GP64()
	MOVQ(idx, next)
	ahead := GP64()
	MOVQ(U64(256), ahead)
	ADDQ(ahead, next)
	CMPQ(next, len)
	JAE(LabelRef("x86_loop"))

	y := YMM()
	yspaces := YMM()
	ynewlines := YMM()
	ycarriages := YMM()
	VLDDQU(ptr, y)
	VPCMPEQB(y, spaces, yspaces)
	VPCMPEQB(y, newline, ynewlines)
	VPCMPEQB(y, carriage, ycarriages)
	results := YMM()
	VPOR(yspaces, ynewlines, results)
	VPOR(ycarriages, results, results)
	zero := YMM()
	VPOR(zero, zero, zero)
	VPTEST(results, zero)
	// Increment ptrs and loop.
	MOVQ(next, idx)
	JCC(LabelRef("avx2_loop"))
	//__m256i x = _mm256_loadu_si256((const __m256i *)(bytes + i));
	//// we do it the naive way, could be smarter?
	//__m256i xspaces = _mm256_cmpeq_epi8(x, spaces);
	//__m256i xnewline = _mm256_cmpeq_epi8(x, newline);
	//__m256i xcarriage = _mm256_cmpeq_epi8(x, carriage);
	//__m256i anywhite =
	//	_mm256_or_si256(_mm256_or_si256(xspaces, xnewline), xcarriage);
	//if (_mm256_testz_si256(anywhite, anywhite) == 1) { // no white space
	//	_mm256_storeu_si256((__m256i *)(bytes+pos), x);
	//	pos += 32;
	//}
	// Increment ptrs and loop.
	MOVQ(next, idx)
	JMP(LabelRef("avx2_loop"))

	Label("x86_loop")
	// Do something
	JAE(LabelRef("return"))

	Label("return")
	RET()
	Generate()

}
