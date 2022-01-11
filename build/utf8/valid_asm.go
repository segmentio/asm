//go:build ignore
// +build ignore

package main

import (
	"bytes"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/asm"
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

	Comment("MSB mask")
	msbMask := ConstArray64("msb_mask",
		0x8080808080808080,
		0x8080808080808080,
		0x8080808080808080,
		0x8080808080808080,
	)

	msbMaskY := YMM()
	VMOVDQU(msbMask, msbMaskY)

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

	currentBlockY := YMM()

	Comment("if bytes left >= 32")
	CMPQ(n, U8(32))
	Comment("go process the next block")
	JGE(LabelRef("process"))

	Comment("If < 32 bytes left")

	Comment("Fast exit if done")
	CMPQ(n, U8(0))
	JE(LabelRef("end"))

	// At this point we know we need to load up to 32 bytes of input to
	// finish the validation and pad the rest of the input vector with
	// zeroes.
	//
	// This code assumes that the remainder of the input data ends right
	// before a page boundary. As a result, we need to take special care to
	// avoid a page fault.
	//
	// At a high level:
	//
	// 1. Move back the data pointer so that the 32 bytes load ends exactly
	// where the input does.
	//
	// 2. Shift right the loaded input so that the remaining input starts at
	// the beginning of the vector.
	//
	// 3. Pad the rest of the vector with zeroes.
	//
	// Because AVX2 32 bytes vectors are really two 16 bytes vector, we need
	// to jump through hoops to perform the shift operation accross
	// lates. This code has two versions, one for inputs of less than 16
	// bytes, and one for larger inputs. Though the latter as more steps,
	// they work using a shuffle mask to shift the bytes in the vector, and
	// a blend operation to stich together the various pieces of the
	// resulting vector.
	//
	// TODO: faster load code when not near a page boundary.

	Comment("If 0 < bytes left < 32")

	zeroes := YMM()
	VPXOR(zeroes, zeroes, zeroes)

	shuffleMaskBytes := make([]byte, 3*16)
	for i := byte(0); i < 16; i++ {
		shuffleMaskBytes[i] = i
		shuffleMaskBytes[i+16] = i
		shuffleMaskBytes[i+32] = i
	}
	shuffleMask := ConstBytes("shuffle_mask", shuffleMaskBytes)

	shuffleClearMaskBytes := make([]byte, 3*16)
	for i := byte(0); i < 16; i++ {
		shuffleClearMaskBytes[i] = i
		shuffleClearMaskBytes[i+16] = 0xFF
		shuffleClearMaskBytes[i+32] = 0xFF
	}
	shuffleClearMask := ConstBytes("shuffle_clear_mask", shuffleClearMaskBytes)

	offset := GP64()
	shuffleMaskPtr := GP64()
	shuffle := YMM()
	tmp1 := YMM()

	MOVQ(U64(32), offset)
	SUBQ(n, offset)

	SUBQ(offset, d)

	VMOVDQU(Mem{Base: d}, currentBlockY)

	CMPQ(n, U8(16))
	JA(LabelRef("tail_load_large"))

	Comment("Shift right that works if remaining bytes <= 16, safe next to a page boundary")

	VPERM2I128(Imm(3), currentBlockY, zeroes, currentBlockY)

	LEAQ(shuffleClearMask.Offset(16), shuffleMaskPtr)
	ADDQ(n, offset)
	ADDQ(n, offset)
	SUBQ(Imm(32), offset)
	SUBQ(offset, shuffleMaskPtr)
	VMOVDQU(Mem{Base: shuffleMaskPtr}, shuffle)

	VPSHUFB(shuffle, currentBlockY, currentBlockY)

	XORQ(n, n)
	JMP(LabelRef("loaded"))

	Comment("Shift right that works if remaining bytes >= 16, safe next to a page boundary")
	Label("tail_load_large")

	ADDQ(n, offset)
	ADDQ(n, offset)
	SUBQ(Imm(48), offset)

	LEAQ(shuffleMask.Offset(16), shuffleMaskPtr)
	SUBQ(offset, shuffleMaskPtr)
	VMOVDQU(Mem{Base: shuffleMaskPtr}, shuffle)

	VPSHUFB(shuffle, currentBlockY, tmp1)

	tmp2 := YMM()
	VPERM2I128(Imm(3), currentBlockY, zeroes, tmp2)

	VPSHUFB(shuffle, tmp2, tmp2)

	blendMaskBytes := make([]byte, 3*16)
	for i := byte(0); i < 16; i++ {
		blendMaskBytes[i] = 0xFF
		blendMaskBytes[i+16] = 0x00
		blendMaskBytes[i+32] = 0xFF
	}
	blendMask := ConstBytes("blend_mask", blendMaskBytes)

	blendMaskStartPtr := GP64()
	LEAQ(blendMask.Offset(16), blendMaskStartPtr)
	SUBQ(offset, blendMaskStartPtr)

	blend := YMM()
	VBROADCASTF128(Mem{Base: blendMaskStartPtr}, blend)
	VPBLENDVB(blend, tmp1, tmp2, currentBlockY)

	XORQ(n, n)
	JMP(LabelRef("loaded"))

	Comment("Process one 32B block of data")
	Label("process")

	Comment("Load the next block of bytes")
	VMOVDQU(Mem{Base: d}, currentBlockY)
	SUBQ(U8(32), n)
	ADDQ(U8(32), d)

	Label("loaded")

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

	Comment("Prepare intermediate vector for push operations")
	vp := YMM()
	VPERM2I128(Imm(3), previousBlockY, currentBlockY, vp)

	Comment("Check errors on the high nibble of the previous byte")
	previousY := YMM()
	VPALIGNR(Imm(15), vp, currentBlockY, previousY)

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
	off2 := YMM()
	VPALIGNR(Imm(14), vp, currentBlockY, off2)
	VPSUBUSB(continuation3BytesY, off2, off2)

	Comment("Find 4 bytes continuations")
	off3 := YMM()
	VPALIGNR(Imm(13), vp, currentBlockY, off3)

	VPSUBUSB(continuation4BytesY, off3, off3)

	Comment("Combine them to have all continuations")
	continuationBitsY := YMM()
	VPOR(off2, off3, continuationBitsY)

	Comment("Perform a byte-sized signed comparison with zero to turn any non-zero bytes into 0xFF.")
	tmpY := zeroOutVector(YMM())
	VPCMPGTB(tmpY, continuationBitsY, continuationBitsY)

	Comment("Find bytes that are continuations by looking at their most significant bit.")
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
