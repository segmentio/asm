//go:build ignore
// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/asm"
	x86 "github.com/segmentio/asm/build/internal/x86"
)

func init() {
	ConstraintExpr("!pure go")
}

func main() {
	TEXT("Valid", NOSPLIT, "func(p []byte) bool")
	Doc("Valid reports whether p consists entirely of valid UTF-8-encoded runes.")

	ret, _ := ReturnIndex(0).Resolve()

	d := Load(Param("p").Base(), GP64())
	n := Load(Param("p").Len(), GP64())

	scratch := AllocLocal(32)
	scratchAddr := GP64()
	LEAQ(scratch, scratchAddr)

	Comment("Prepare the constant masks")

	incompleteMask := ConstArray64("incomplete_mask",
		0xFFFFFFFFFFFFFFFF,
		0xFFFFFFFFFFFFFFFF,
		0xFFFFFFFFFFFFFFFF,
		0xBFDFEFFFFFFFFFFF,
	)
	incompleteMaskY := YMM()
	VMOVDQU(incompleteMask, incompleteMaskY)

	continuation4Bytes := ConstArray64("cont4_vec",
		0xEFEFEFEFEFEFEFEF,
		0xEFEFEFEFEFEFEFEF,
		0xEFEFEFEFEFEFEFEF,
		0xEFEFEFEFEFEFEFEF,
	)

	continuation4BytesY := YMM()
	VMOVDQU(continuation4Bytes, continuation4BytesY)

	continuation3Bytes := ConstArray64("cont3_vec",
		0xDFDFDFDFDFDFDFDF,
		0xDFDFDFDFDFDFDFDF,
		0xDFDFDFDFDFDFDFDF,
		0xDFDFDFDFDFDFDFDF,
	)

	continuation3BytesY := YMM()
	VMOVDQU(continuation3Bytes, continuation3BytesY)

	Comment("High nibble of current byte")
	nibble1Errors := ConstArray32("nibble1_errors",
		0x02020202,
		0x02020202,
		0x80808080,
		0x49150121,
		0x02020202,
		0x02020202,
		0x80808080,
		0x49150121,
	)

	nibble1Y := YMM()
	VMOVDQU(nibble1Errors, nibble1Y)

	Comment("Low nibble of current byte")
	nibble2Errors := ConstArray32("nibble2_errors",
		0x8383A3E7,
		0xCBCBCB8B,
		0xCBCBCBCB,
		0xCBCBDBCB,
		0x8383A3E7,
		0xCBCBCB8B,
		0xCBCBCBCB,
		0xCBCBDBCB,
	)

	nibble2Y := YMM()
	VMOVDQU(nibble2Errors, nibble2Y)

	Comment("High nibble of the next byte")
	nibble3Errors := ConstArray32("nibble3_errors",
		0x01010101,
		0x01010101,
		0xBABAAEE6,
		0x01010101,
		0x01010101,
		0x01010101,
		0xBABAAEE6,
		0x01010101,
	)

	nibble3Y := YMM()
	VMOVDQU(nibble3Errors, nibble3Y)

	Comment("Nibble mask")
	nibbleMask := ConstArray64("nibble_mask",
		0x0F0F0F0F0F0F0F0F,
		0x0F0F0F0F0F0F0F0F,
		0x0F0F0F0F0F0F0F0F,
		0x0F0F0F0F0F0F0F0F,
	)

	nibbleMaskY := YMM()
	VMOVDQU(nibbleMask, nibbleMaskY)

	Comment("For the first pass, set the previous block as zero.")
	previousBlockY := YMM()
	zeroOutVector(previousBlockY)

	Comment("Zeroes the error vector.")
	errorY := YMM()
	zeroOutVector(errorY)

	Comment(`Zeroes the "previous block was incomplete" vector.`)
	incompletePreviousBlockY := YMM()
	zeroOutVector(incompletePreviousBlockY)

	Comment("Top of the loop.")
	Label("check_input")

	Comment("if bytes left >= 32")
	CMPQ(n, U8(32))
	Comment("go process the next block")
	JGE(LabelRef("process"))

	Comment("If < 32 bytes left")

	Comment("Fast exit if done")
	CMPQ(n, U8(0))
	JE(LabelRef("end"))

	Comment("If 0 < bytes left < 32.")
	Comment("At that point we know we need the scratch buffer.")
	Comment("Zero it") // TODO: is this necessary or is the runtime already doing it?
	tmpZeroY := zeroOutVector(YMM())
	VMOVDQU(tmpZeroY, Mem{Base: scratchAddr})

	Comment("Make a copy of the remaining bytes into the zeroed scratch space and make it the next block to read.")
	copySrcAddr := GP64()
	MOVQ(d, copySrcAddr)
	MOVQ(scratchAddr, d)

	copyN(d, copySrcAddr, n)
	MOVQ(U64(32), n)

	Comment("Process one 32B block of data")
	Label("process")
	currentBlockY := YMM()

	Comment("Load the next block of bytes")
	VMOVDQU(Mem{Base: d}, currentBlockY)
	SUBQ(U8(32), n)
	ADDQ(U8(32), d)

	Comment("Fast check to see if ASCII")
	tmp := GP32()
	VPMOVMSKB(currentBlockY, tmp)
	CMPL(tmp, Imm(0))
	JNZ(LabelRef("non_ascii"))

	Comment("If this all block is ASCII, there is nothing to do, and it is an error if any of the previous code point was incomplete.")
	VPOR(errorY, incompletePreviousBlockY, errorY)
	JMP(LabelRef("check_input"))

	Label("non_ascii")
	Comment("Check errors on the high nibble of the previous byte")
	previousY := pushLastByteFromAToFrontOfB(previousBlockY, currentBlockY)

	highPrev := highNibbles(previousY, nibbleMaskY)
	VPSHUFB(highPrev, nibble1Y, highPrev)

	Comment("Check errors on the low nibble of the previous byte")
	lowPrev := lowNibbles(previousY, nibbleMaskY)
	VPSHUFB(lowPrev, nibble2Y, lowPrev)
	VPAND(lowPrev, highPrev, highPrev)

	Comment("Check errors on the high nibble on the current byte")
	highCurr := highNibbles(currentBlockY, nibbleMaskY)
	VPSHUFB(highCurr, nibble3Y, highCurr)
	VPAND(highCurr, highPrev, highPrev)

	Comment("Find 2 bytes continuations")
	off2 := pushLast2BytesFromAToFrontOfB(previousBlockY, currentBlockY)
	VPSUBUSB(continuation3BytesY, off2, off2)

	Comment("Find 3 bytes continuations")
	off3 := pushLast3BytesFromAToFrontOfB(previousBlockY, currentBlockY)
	VPSUBUSB(continuation4BytesY, off3, off3)

	Comment("Combine them to have all continuations")
	continuationBitsY := YMM()
	VPOR(off2, off3, continuationBitsY)

	Comment("Perform a byte-sized signed comparison with zero to turn any non-zero bytes into 0xFF.")
	tmpY := zeroOutVector(YMM())
	VPCMPGTB(tmpY, continuationBitsY, continuationBitsY)

	Comment("Find bytes that are continuations by looking at their most significant bit.")
	msbMask := ConstArray64("msb_mask",
		0x8080808080808080,
		0x8080808080808080,
		0x8080808080808080,
		0x8080808080808080,
	)

	msbMaskY := YMM()
	VMOVDQU(msbMask, msbMaskY)
	VPAND(msbMaskY, continuationBitsY, continuationBitsY)

	Comment("Find mismatches between expected and actual continuation bytes")
	VPXOR(continuationBitsY, highPrev, continuationBitsY)

	Comment("Store result in sticky error")
	VPOR(errorY, continuationBitsY, errorY)

	Comment("Prepare for next iteration")
	VPSUBUSB(incompleteMaskY, currentBlockY, incompletePreviousBlockY)
	VMOVDQU(currentBlockY, previousBlockY)

	Comment("End of loop")
	JMP(LabelRef("check_input"))

	Label("end")

	Comment("If the previous block was incomplete, this is an error.")
	VPOR(incompletePreviousBlockY, errorY, errorY)

	Comment("Return whether any error bit was set")
	VPTEST(errorY, errorY)
	SETEQ(ret.Addr)
	RET()

	Generate()
}

func pushLast2BytesFromAToFrontOfB(a, b VecVirtual) VecVirtual {
	out := YMM()
	VPERM2I128(Imm(3), a, b, out)
	VPALIGNR(Imm(14), out, b, out)
	return out
}

func pushLast3BytesFromAToFrontOfB(a, b VecVirtual) VecVirtual {
	out := YMM()
	VPERM2I128(Imm(3), a, b, out)
	VPALIGNR(Imm(13), out, b, out)
	return out
}

func pushLastByteFromAToFrontOfB(a, b VecVirtual) VecVirtual {
	out := YMM()
	VPERM2I128(Imm(3), a, b, out)
	VPALIGNR(Imm(15), out, b, out)
	return out
}

func lowNibbles(a VecVirtual, nibbleMask VecVirtual) VecVirtual {
	out := YMM()
	VPAND(a, nibbleMask, out)
	return out
}

func highNibbles(a VecVirtual, nibbleMask VecVirtual) VecVirtual {
	out := YMM()
	VPSRLW(Imm(4), a, out)
	VPAND(out, nibbleMask, out)
	return out
}

func zeroOutVector(y VecVirtual) VecVirtual {
	VXORPS(y, y, y)
	return y
}

// Assumptions:
// - len(dst) == 32 bytes
// - 0 < len(src) < 32
// TODO: try to use x86.bytes
func copyN(dst Register, src Register, n Register) {
	v := x86.VariableLengthBytes{
		Unroll: 128,
		Process: func(regs []Register, memory ...x86.Memory) {
			src, dst := regs[0], regs[1]

			count := len(memory)
			operands := make([]Op, count*2)

			for i, m := range memory {
				operands[i] = m.Load(src)
			}

			for i, m := range memory {
				m.Store(operands[i].(Register), dst)
			}
		},
	}

	inputs := []Register{src, dst}

	CMPQ(n, Imm(1))
	JE(LabelRef("handle1"))

	CMPQ(n, Imm(3))
	JBE(LabelRef("handle2to3"))

	CMPQ(n, Imm(4))
	JE(LabelRef("handle4"))

	CMPQ(n, Imm(8))
	JB(LabelRef("handle5to7"))
	JE(LabelRef("handle8"))

	CMPQ(n, Imm(16))
	JBE(LabelRef("handle9to16"))

	CMPQ(n, Imm(32))
	JBE(LabelRef("handle17to32"))

	Label("handle1")
	v.Process(inputs, x86.Memory{Size: 1})
	JMP(LabelRef("after_copy"))

	Label("handle2to3")
	v.Process(inputs,
		x86.Memory{Size: 2},
		x86.Memory{Size: 2, Index: n, Offset: -2})
	JMP(LabelRef("after_copy"))

	Label("handle4")
	v.Process(inputs, x86.Memory{Size: 4})
	JMP(LabelRef("after_copy"))

	Label("handle5to7")
	v.Process(inputs,
		x86.Memory{Size: 4},
		x86.Memory{Size: 4, Index: n, Offset: -4})
	JMP(LabelRef("after_copy"))

	Label("handle8")
	v.Process(inputs, x86.Memory{Size: 8})
	JMP(LabelRef("after_copy"))

	Label("handle9to16")
	v.Process(inputs,
		x86.Memory{Size: 8},
		x86.Memory{Size: 8, Index: n, Offset: -8})
	JMP(LabelRef("after_copy"))

	Label("handle17to32")
	v.Process(inputs,
		x86.Memory{Size: 16},
		x86.Memory{Size: 16, Index: n, Offset: -16})

	Label("after_copy")
}
