//go:build ignore
// +build ignore

package main

import (
	"bytes"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/gotypes"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/asm"
	. "github.com/segmentio/asm/build/internal/x86"
	"github.com/segmentio/asm/cpu"
)

func init() {
	ConstraintExpr("!purego")
}

const (
	runeSelf = 0x80

	// The default lowest and highest continuation byte.
	locb = 0b10000000 // 128 0x80
	hicb = 0b10111111 // 191 0xBF

	// These names of these constants are chosen to give nice alignment in the
	// table below. The first nibble is an index into acceptRanges or F for
	// special one-byte cases. The second nibble is the Rune length or the
	// Status for the special one-byte case.
	xx = 0xF1 // invalid: size 1
	as = 0xF0 // ASCII: size 1
	s1 = 0x02 // accept 0, size 2
	s2 = 0x13 // accept 1, size 3
	s3 = 0x03 // accept 0, size 3
	s4 = 0x23 // accept 2, size 3
	s5 = 0x34 // accept 3, size 4
	s6 = 0x04 // accept 0, size 4
	s7 = 0x44 // accept 4, size 4
)

// TODO: find the address of this from unicode/utf8.
// first is information about the first byte in a UTF-8 sequence.
var firstData = [256]uint8{
	//   1   2   3   4   5   6   7   8   9   A   B   C   D   E   F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x00-0x0F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x10-0x1F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x20-0x2F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x30-0x3F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x40-0x4F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x50-0x5F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x60-0x6F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x70-0x7F
	//   1   2   3   4   5   6   7   8   9   A   B   C   D   E   F
	xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0x80-0x8F
	xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0x90-0x9F
	xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0xA0-0xAF
	xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0xB0-0xBF
	xx, xx, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, // 0xC0-0xCF
	s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, // 0xD0-0xDF
	s2, s3, s3, s3, s3, s3, s3, s3, s3, s3, s3, s3, s3, s4, s3, s3, // 0xE0-0xEF
	s5, s6, s6, s6, s7, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0xF0-0xFF
}

// acceptRange gives the range of valid values for the second byte in a UTF-8
// sequence.
type acceptRange struct {
	lo uint8 // lowest value for second byte.
	hi uint8 // highest value for second byte.
}

var acceptRangesData = [32]uint8{
	locb, hicb,
	0xA0, hicb,
	locb, 0x9F,
	0x90, hicb,
	locb, 0x8F,
}

func stdlib(d Register, n Register, validAsciiReg Register, retUtf8 *Basic, retAscii *Basic) {
	Comment("Non-vectorized implementation from the stdlib. Used for small inputs.")
	mask := GP64()
	MOVQ(U64(0x8080808080808080), mask)

	Comment("Fast ascii-check loop")
	Label("start_loop")
	CMPQ(n, U8(8))
	JL(LabelRef("end_loop"))
	tmp := GP64()
	MOVQ(Mem{Base: d}, tmp)
	TESTQ(mask, tmp)
	JNZ(LabelRef("fail_loop"))
	SUBQ(U8(8), n)
	ADDQ(U8(8), d)
	JMP(LabelRef("start_loop"))
	Label("fail_loop")
	XORB(validAsciiReg, validAsciiReg)
	Label("end_loop")

	Comment("UTF-8 check byte-by-byte")

	end := GP64()
	LEAQ(Mem{Base: d, Index: n, Scale: 1}, end)

	first := Mem{Base: GP64(), Scale: 1}
	LEAQ(ConstBytes("first", firstData[:]), first.Base)

	acceptRanges := Mem{Base: GP64(), Scale: 2}
	LEAQ(ConstBytes("accept_ranges", acceptRangesData[:]), acceptRanges.Base)
	JMP(LabelRef("start_utf8_loop_set"))

	Label("start_utf8_loop") // for
	nextD := GP64()
	MOVQ(nextD, d)
	Label("start_utf8_loop_set")
	CMPQ(d, end)                     // i < n
	JGE(LabelRef("stdlib_ret_true")) //   end of loop, return true

	pi := GP32()
	MOVBLZX(Mem{Base: d}, pi) // pi = b[i]

	CMPB(pi.As8(), Imm(runeSelf))        // if pi >= runeSelf
	JAE(LabelRef("test_first"))          //   more testing to do
	LEAQ(Mem{Base: d}.Offset(1), d)      // else: i++
	JMP(LabelRef("start_utf8_loop_set")) //   continue

	Label("test_first")
	XORB(validAsciiReg, validAsciiReg)
	x := GP32()
	XORL(x, x)
	MOVB(first.Idx(pi, 1), x.As8())   // x = first[pi]
	CMPB(x.As8(), Imm(xx))            // if x == xx
	JEQ(LabelRef("stdlib_ret_false")) //   return false (illegal started byte)

	size := GP32()
	MOVBLZX(x.As8(), size) // size = x
	ANDL(Imm(0x7), size)   // size &= 7
	LEAQ(Mem{Base: d}.Idx(size, 1), nextD)
	CMPQ(nextD, end)                 // if i2 > n
	JA(LabelRef("stdlib_ret_false")) //  return false (short or invalid)

	SHRB(Imm(4), x.As8()) // x = x >> 4

	acceptLo := GP8()
	MOVBLZX(acceptRanges.Idx(x, 2).Offset(0), acceptLo.As32())
	acceptHi := GP8()
	MOVBLZX(acceptRanges.Idx(x, 2).Offset(1), acceptHi.As32())

	c1 := GP8()
	MOVB(Mem{Base: d}.Offset(1), c1) // c = b[i+1]
	CMPB(c1, acceptLo)               // if c < accept.lo
	JB(LabelRef("stdlib_ret_false")) //   return false
	CMPB(acceptHi, c1)               // if accept.hi < c
	JB(LabelRef("stdlib_ret_false")) //   return false

	CMPL(size, Imm(2))               // if size == 2
	JEQ(LabelRef("start_utf8_loop")) //   -> inc_size

	c2 := GP32()
	MOVBLZX(Mem{Base: d}.Offset(2), c2) // c = b[i+2]
	SUBL(Imm(locb), c2)
	CMPB(c2.As8(), Imm(hicb-locb))
	JHI(LabelRef("stdlib_ret_false"))

	CMPL(size, Imm(3))               // if size == 3
	JEQ(LabelRef("start_utf8_loop")) //   -> inc_size

	c3 := GP32()
	MOVBLZX(Mem{Base: d}.Offset(3), c3) // c = b[i+3]
	SUBL(Imm(locb), c3)
	CMPB(c3.As8(), Imm(hicb-locb))
	JLS(LabelRef("start_utf8_loop"))

	Label("stdlib_ret_false")
	MOVB(Imm(0), retUtf8.Addr)
	MOVB(validAsciiReg, retAscii.Addr)
	RET()

	Label("stdlib_ret_true")
	MOVB(Imm(1), retUtf8.Addr)
	MOVB(validAsciiReg, retAscii.Addr)
	RET()

	Comment("End of stdlib implementation")
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
	TEXT("Validate", NOSPLIT, "func(p []byte) (bool, bool)")
	Doc("Validate is a more precise version of Valid that also indicates whether the input was valid ASCII.")

	retUtf8, _ := ReturnIndex(0).Resolve()
	retAscii, _ := ReturnIndex(1).Resolve()

	d := Load(Param("p").Base(), GP64())
	n := Load(Param("p").Len(), GP64())

	validAsciiReg := GP8()
	MOVB(Imm(1), validAsciiReg)

	JumpUnlessFeature("stdlib", cpu.AVX2)

	// 32 has been found empirically on an Intel i7-8559U machine. After
	// that size, the AVX2 implementation is faster than the stdlib one.
	Comment("if input < 32 bytes")
	CMPQ(n, U8(32))
	JGE(LabelRef("init_avx"))

	Label("stdlib")
	stdlib(d, n, validAsciiReg, retUtf8, retAscii)

	Label("init_avx")

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

	hasPreviousBlock := GP8()
	XORB(hasPreviousBlock, hasPreviousBlock)

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

	// If the AVX code never ran, we can proceed with using the stdlib
	// implementation.
	CMPB(hasPreviousBlock, Imm(1))
	JNE(LabelRef("stdlib"))

	// Fast exit if the error vector is not empty.
	VPTEST(errorY, errorY)
	JNZ(LabelRef("exit"))

	// If the AVX code has ran at least once, we need to walk back up to 4
	// bytes to take into account continuations. This is done by
	// substracting the current input offset with the number of bytes
	// between the first non-zero byte of incompletePreviousBlockY and the
	// end of the vector. That way, stdlib starts at the first known
	// incomplete byte.

	zeroes := zeroOutVector(YMM())
	VPCMPEQB(incompletePreviousBlockY, zeroes, zeroes)
	// 'zeroes' now contains 1111 for all bytes that were zero, 0000
	// otherwise.

	bitset := GP64()
	VPMOVMSKB(zeroes, bitset.As32())
	// bitset now contains a 1 at the position of each zero byte of
	// incompletePreviousBlockY, 0 otherwise.
	NOTL(bitset.As32())
	// Now bitset has 0 bits for each zero byte of incompltePreviousBlockY.
	LZCNTL(bitset.As32(), bitset.As32())
	// bitset is now an unsigned int in [0,32], corresponding to the number
	// of leading zero bytes in incompletePreviousBlockY.
	SUBQ(Imm(32), d)
	ADDQ(bitset, d)
	ADDQ(Imm(32), n)
	SUBQ(bitset, n)
	JMP(LabelRef("stdlib"))

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
	XORB(validAsciiReg, validAsciiReg)
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
	MOVB(Imm(1), hasPreviousBlock)

	Comment("End of loop")
	JMP(LabelRef("check_input"))

	Label("end")

	Comment("If the previous block was incomplete, this is an error.")
	VPOR(incompletePreviousBlockY, errorY, errorY)

	Comment("Return whether any error bit was set")
	VPTEST(errorY, errorY)
	Label("exit")
	SETEQ(retUtf8.Addr)
	MOVB(validAsciiReg, retAscii.Addr)
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
