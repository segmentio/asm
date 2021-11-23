//go:build ignore
// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/gotypes"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/asm"
	"github.com/segmentio/asm/build/internal/x86"
)

func init() {
	ConstraintExpr("!pure go")
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

func stdlib(d Register, n Register, ret *Basic) {
	Comment("Non-vectorized implementation from the stdlib. Used for small inputs.")
	mask := GP64()
	MOVQ(U64(0x8080808080808080), mask)

	first := Mem{Base: GP64(), Scale: 1}
	LEAQ(ConstBytes("first", firstData[:]), first.Base)

	acceptRanges := Mem{Base: GP64(), Scale: 2}
	LEAQ(ConstBytes("accept_ranges", acceptRangesData[:]), acceptRanges.Base)

	Comment("Fast ascii-check loop")
	Label("start_loop")
	CMPQ(n, U8(32))
	JL(LabelRef("end_loop"))
	TESTQ(mask, d)
	JNZ(LabelRef("end_loop"))
	SUBQ(U8(8), n)
	ADDQ(U8(8), d)
	JMP(LabelRef("start_loop"))
	Label("end_loop")

	Comment("UTF-8 check byte-by-byte")
	i := GP64()
	XORQ(i, i)

	b := Mem{Base: d, Index: i, Scale: 1}
	pi := GP32()

	Label("start_utf8_loop")         // for
	CMPQ(i, n)                       // i < n
	JGE(LabelRef("stdlib_ret_true")) //   end of loop, return true

	MOVBLZX(b, pi) // pi = b[i]

	CMPB(pi.As8(), Imm(runeSelf))    // if pi >= runeSelf
	JAE(LabelRef("test_first"))      //   more testing to do
	ADDQ(Imm(1), i)                  // else: i++
	JMP(LabelRef("start_utf8_loop")) //   continue

	Label("test_first")
	x := GP8()
	MOVB(first.Idx(pi, 1), x)         // x = first[pi]
	CMPB(x, Imm(xx))                  // if x == xx
	JEQ(LabelRef("stdlib_ret_false")) //   return false (illegal started byte)

	size := GP64()
	MOVBQZX(x, size)     // size = x
	ANDQ(Imm(0x7), size) // size &= 7
	i2 := GP64()
	MOVQ(i, i2)                      // i2 = i
	ADDQ(size, i2)                   // i2 += size
	CMPQ(i2, n)                      // if i2 > n
	JA(LabelRef("stdlib_ret_false")) //  return false (short or invalid)

	accept := GP16()
	SHRB(Imm(4), x)                      // x = x >> 4
	MOVW(acceptRanges.Idx(x, 2), accept) // accept = acceptRanges[x]
	acceptLo := GP8()
	MOVB(accept.As8(), acceptLo) // acceptLo = accept.lo
	SHRW(Imm(8), accept)         // accept = accept.hi
	// TODO: ^ this method to grab the fields seems odd

	c := GP8()
	MOVB(b.Offset(1), c)             // c = b[i+1]
	CMPB(c, acceptLo)                // if c < accept.lo
	JB(LabelRef("stdlib_ret_false")) //   return false
	CMPB(accept.As8(), c)            // if accept.hi < c
	JB(LabelRef("stdlib_ret_false")) //   return false

	CMPQ(size, Imm(2))        // if size == 2
	JEQ(LabelRef("inc_size")) //   -> inc_size

	MOVB(b.Offset(2), c)             // c = b[i+2]
	CMPB(c, U8(locb))                // if c < locb
	JB(LabelRef("stdlib_ret_false")) //   return false
	CMPB(c, U8(hicb))                // if hicb < c
	JA(LabelRef("stdlib_ret_false")) //   return false

	CMPQ(size, Imm(3))        // if size == 3
	JEQ(LabelRef("inc_size")) //   -> inc_size

	MOVB(b.Offset(3), c)             // c = b[i+3]
	CMPB(c, Imm(locb))               // if c < locb
	JB(LabelRef("stdlib_ret_false")) //   return false
	CMPB(c, Imm(hicb))               // if hicb < c
	JA(LabelRef("stdlib_ret_false")) //   return false

	Label("inc_size")
	ADDQ(size, i) // i += size

	JMP(LabelRef("start_utf8_loop"))

	Label("stdlib_ret_true")
	MOVB(Imm(1), ret.Addr)
	RET()
	Label("stdlib_ret_false")
	MOVB(Imm(0), ret.Addr)
	RET()

	Comment("End of stdlib implementation")
}

func main() {
	TEXT("Valid", NOSPLIT, "func(p []byte) bool")
	Doc("Valid reports whether p consists entirely of valid UTF-8-encoded runes.")

	ret, _ := ReturnIndex(0).Resolve()

	d := Load(Param("p").Base(), GP64())
	n := Load(Param("p").Len(), GP64())

	//	x86.JumpUnlessFeature("stdlib", cpu.AVX2) // TODO
	JMP(LabelRef("stdlib")) // TODO:REMOVE ME

	Comment("if input < 32 bytes")
	CMPQ(n, U8(32))
	JG(LabelRef("init_avx"))

	Label("stdlib")
	stdlib(d, n, ret)

	Label("init_avx")

	scratch := AllocLocal(32)
	scratchAddr := GP64()
	LEAQ(scratch, scratchAddr)

	//stdlib(d, n, ret)

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