//go:build ignore
// +build ignore

package main

import (
	"bytes"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/asm"
	. "github.com/segmentio/asm/build/internal/x86"
	"github.com/segmentio/asm/cpu"
)

func init() {
	ConstraintExpr("!purego")
}

func incompleteMaskData() []byte {
	// The incomplete mask is used on every block to flag the bytes that are
	// incomplete if this is the last block (for example a byte that starts
	// a 4 byte character only 3 bytes before the end).
	any := byte(0xFF)
	needs4 := byte(0b11110000) - 1
	needs3 := byte(0b11100000) - 1
	needs2 := byte(0b11000000) - 1
	b := [32]byte{
		any, any, any, any, any, any, any, any,
		any, any, any, any, any, any, any, any,
		any, any, any, any, any, any, any, any,
		any, any, any, any, any, needs4, needs3, needs2,
	}
	return b[:]
}

func continuationMaskData(pattern byte) []byte {
	// Pattern is something like 0b11100000 to accept all bytes of the form
	// 111xxxxx.
	v := pattern - 1
	return bytes.Repeat([]byte{v}, 32)
}

func nibbleMasksData() (nib1, nib2, nib3 []byte) {
	const (
		TooShort     = 1 << 0
		TooLong      = 1 << 1
		Overlong3    = 1 << 2
		Surrogate    = 1 << 4
		Overlong2    = 1 << 5
		TwoConts     = 1 << 7
		TooLarge     = 1 << 3
		TooLarge1000 = 1 << 6
		Overlong4    = 1 << 6
		Carry        = TooShort | TooLong | TwoConts
	)

	fullMask := func(b [16]byte) []byte {
		m := make([]byte, 32)
		copy(m, b[:])
		copy(m[16:], b[:])
		return m
	}

	nib1 = fullMask([16]byte{
		// 0_______ ________ <ASCII in byte 1>
		TooLong, TooLong, TooLong, TooLong,
		TooLong, TooLong, TooLong, TooLong,
		// 10______ ________ <continuation in byte 1>
		TwoConts, TwoConts, TwoConts, TwoConts,
		// 1100____ ________ <two byte lead in byte 1>
		TooShort | Overlong2,
		// 1101____ ________ <two byte lead in byte 1>
		TooShort,
		// 1110____ ________ <three byte lead in byte 1>
		TooShort | Overlong3 | Surrogate,
		// 1111____ ________ <four+ byte lead in byte 1>
		TooShort | TooLarge | TooLarge1000 | Overlong4,
	})

	nib2 = fullMask([16]byte{
		// ____0000 ________
		Carry | Overlong3 | Overlong2 | Overlong4,
		// ____0001 ________
		Carry | Overlong2,
		// ____001_ ________
		Carry,
		Carry,

		// ____0100 ________
		Carry | TooLarge,
		// ____0101 ________
		Carry | TooLarge | TooLarge1000,
		// ____011_ ________
		Carry | TooLarge | TooLarge1000,
		Carry | TooLarge | TooLarge1000,

		// ____1___ ________
		Carry | TooLarge | TooLarge1000,
		Carry | TooLarge | TooLarge1000,
		Carry | TooLarge | TooLarge1000,
		Carry | TooLarge | TooLarge1000,
		Carry | TooLarge | TooLarge1000,
		// ____1101 ________
		Carry | TooLarge | TooLarge1000 | Surrogate,
		Carry | TooLarge | TooLarge1000,
		Carry | TooLarge | TooLarge1000,
	})

	nib3 = fullMask([16]byte{
		// ________ 0_______ <ASCII in byte 2>
		TooShort, TooShort, TooShort, TooShort,
		TooShort, TooShort, TooShort, TooShort,

		// ________ 1000____
		TooLong | Overlong2 | TwoConts | Overlong3 | TooLarge1000 | Overlong4,
		// ________ 1001____
		TooLong | Overlong2 | TwoConts | Overlong3 | TooLarge,
		// ________ 101_____
		TooLong | Overlong2 | TwoConts | Surrogate | TooLarge,
		TooLong | Overlong2 | TwoConts | Surrogate | TooLarge,

		// ________ 11______
		TooShort, TooShort, TooShort, TooShort,
	})

	return
}

func main() {
	TEXT("validateAvx", NOSPLIT, "func(p []byte) byte")
	Doc("Optimized version of Validate for inputs of more than 32B.")

	ret, err := ReturnIndex(0).Resolve()
	if err != nil {
		panic(err)
	}

	d := Load(Param("p").Base(), GP64())
	n := Load(Param("p").Len(), GP64())

	JumpIfFeature("init_avx", cpu.AVX2)

	// TODO: call stdlib

	Label("init_avx")

	isAscii := GP8()
	MOVB(Imm(1), isAscii)

	Comment("Prepare the constant masks")

	incompleteMask := ConstBytes("incomplete_mask", incompleteMaskData())
	incompleteMaskY := YMM()
	VMOVDQU(incompleteMask, incompleteMaskY)

	continuation4Bytes := ConstBytes("cont4_vec", continuationMaskData(0b11110000))
	continuation4BytesY := YMM()
	VMOVDQU(continuation4Bytes, continuation4BytesY)

	continuation3Bytes := ConstBytes("cont3_vec", continuationMaskData(0b11100000))
	continuation3BytesY := YMM()
	VMOVDQU(continuation3Bytes, continuation3BytesY)

	nib1Data, nib2Data, nib3Data := nibbleMasksData()

	Comment("High nibble of current byte")
	nibble1Errors := ConstBytes("nibble1_errors", nib1Data)
	nibble1Y := YMM()
	VMOVDQU(nibble1Errors, nibble1Y)

	Comment("Low nibble of current byte")
	nibble2Errors := ConstBytes("nibble2_errors", nib2Data)
	nibble2Y := YMM()
	VMOVDQU(nibble2Errors, nibble2Y)

	Comment("High nibble of the next byte")
	nibble3Errors := ConstBytes("nibble3_errors", nib3Data)
	nibble3Y := YMM()
	VMOVDQU(nibble3Errors, nibble3Y)

	Comment("Nibble mask")
	lowerNibbleMask := ConstArray64("nibble_mask",
		0x0F0F0F0F0F0F0F0F,
		0x0F0F0F0F0F0F0F0F,
		0x0F0F0F0F0F0F0F0F,
		0x0F0F0F0F0F0F0F0F,
	)

	nibbleMaskY := YMM()
	VMOVDQU(lowerNibbleMask, nibbleMaskY)

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

	Comment("Prepare scratch space")
	scratch := AllocLocal(32)
	scratchAddr := GP64()
	LEAQ(scratch, scratchAddr)
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

	Comment("If this whole block is ASCII, there is nothing to do, and it is an error if any of the previous code point was incomplete.")
	VPOR(errorY, incompletePreviousBlockY, errorY)
	JMP(LabelRef("check_input"))

	Label("non_ascii")
	XORB(isAscii, isAscii)

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

	Comment("Find 3 bytes continuations")
	off2 := pushLast2BytesFromAToFrontOfB(previousBlockY, currentBlockY)
	VPSUBUSB(continuation3BytesY, off2, off2)

	Comment("Find 4 bytes continuations")
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
	out := GP8()
	SETEQ(out)

	Comment("Bit 0 tells if the input is valid utf8, bit 1 tells if it's valid ascii")
	ANDB(out, isAscii)
	SHLB(Imm(1), isAscii)
	ORB(isAscii, out)

	MOVB(out, ret.Addr)
	VZEROUPPER()
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
	v := VariableLengthBytes{
		Process: func(regs []Register, memory ...Memory) {
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
	v.Process(inputs, Memory{Size: 1})
	JMP(LabelRef("after_copy"))

	Label("handle2to3")
	v.Process(inputs,
		Memory{Size: 2},
		Memory{Size: 2, Index: n, Offset: -2})
	JMP(LabelRef("after_copy"))

	Label("handle4")
	v.Process(inputs, Memory{Size: 4})
	JMP(LabelRef("after_copy"))

	Label("handle5to7")
	v.Process(inputs,
		Memory{Size: 4},
		Memory{Size: 4, Index: n, Offset: -4})
	JMP(LabelRef("after_copy"))

	Label("handle8")
	v.Process(inputs, Memory{Size: 8})
	JMP(LabelRef("after_copy"))

	Label("handle9to16")
	v.Process(inputs,
		Memory{Size: 8},
		Memory{Size: 8, Index: n, Offset: -8})
	JMP(LabelRef("after_copy"))

	Label("handle17to32")
	v.Process(inputs,
		Memory{Size: 16},
		Memory{Size: 16, Index: n, Offset: -16})

	Label("after_copy")
}
