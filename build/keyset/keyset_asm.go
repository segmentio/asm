/// +build ignore

package main

import (
	"fmt"
	"math"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

const unroll = 4

func init() {
	ConstraintExpr("!purego")
}

func main() {
	searchAVX()

	Generate()
}

// searchAVX searches for a key in a set of keys.
//
// Each key in the set of keys should be padded to 16 bytes and concatenated
// into a single buffer. The length of each key should be available in the
// slice of lengths. len(buffer) should equal len(lengths)*16. The routine
// searches for the input key in the set of keys and returns its index if found.
// If not found, the routine returns the number of keys (len(lengths)).
func searchAVX() {
	TEXT("searchAVX", NOSPLIT, "func(buffer *byte, lengths []uint32, key []byte) int")

	// Load the input key and length. Put the length in CX so that we can use
	// it in a variable shift below (SHL only accepts CL for variable shifts).
	keyPtr := Load(Param("key").Base(), GP64())
	keyLen := CX
	Load(Param("key").Len(), keyLen.As64())

	// None of the keys we're searching through have a length greater than 16,
	// so bail early if the input is more than 16 bytes long.
	CMPQ(keyLen.As64(), Imm(16))
	JA(LabelRef("notfound"))

	// Load the remaining inputs.
	buffer := Load(Param("buffer"), GP64())
	lengths := Load(Param("lengths").Base(), GP64())
	count := Load(Param("lengths").Len(), GP64())

	// Load the input key.
	// FIXME: check if near a page boundary and load+shuffle so it doesn't fault
	key := XMM()
	VMOVUPS(Mem{Base: keyPtr}, key)

	// Build a mask with popcount(mask) = keyLen, e.g. for keyLen=4 the mask
	// is 15 (0b1111).
	// TODO: SHL will only accept CL for variable shifts. CL = CX = keyLen. Why
	//  doesn't keyLen.As8L() work?
	match := GP32()
	MOVL(U32(1), match)
	SHLL(CL, match)
	DECL(match)

	// Zero out i so we can use it as the loop increment.
	i := GP64()
	XORQ(i, i)

	// Round the key count down to the nearest multiple of unroll to determine
	// how many iterations of the big loop we'll need.
	truncatedCount := GP64()
	MOVQ(count, truncatedCount)
	shift := uint64(math.Log2(float64(unroll)))
	SHRQ(Imm(shift), truncatedCount)
	SHLQ(Imm(shift), truncatedCount)

	// Loop over multiple keys.
	Label("bigloop")
	CMPQ(i, truncatedCount)
	JE(LabelRef("loop"))

	x := []VecPhysical{X8, X9, X10, X11, X12, X13, X14, X15}
	g := []Physical{R8L, R9L, R10L, R11L, R12L, R13L, R14L, R15L}

	for n := 0; n < unroll; n++ {
		// Try to match against the input key.
		// Check lengths first, then if length matches check bytes match.
		Label(fmt.Sprintf("try%d", n))
		CMPL(keyLen.As32(), Mem{Base: lengths, Index: i, Disp: 4 * n, Scale: 4})
		JNE(LabelRef(fmt.Sprintf("try%d", n + 1)))
		VPCMPEQB(Mem{Base: buffer, Disp: 16 * n}, key, x[n])
		VPMOVMSKB(x[n], g[n])
		ANDL(match, g[n])
		CMPL(match, g[n])
		JNE(LabelRef(fmt.Sprintf("try%d", n + 1)))
		if n > 0 {
			// Correct the loop increment before returning.
			ADDQ(Imm(uint64(n)), i)
		}
		JMP(LabelRef("done"))
	}

	// Advance and loop again.
	Label(fmt.Sprintf("try%d", unroll))
	ADDQ(Imm(unroll), i)
	ADDQ(Imm(16 * unroll), buffer)
	JMP(LabelRef("bigloop"))

	// Loop over the remaining keys.
	Label("loop")
	CMPQ(i, count)
	JE(LabelRef("done"))

	// Try to match against the input key.
	// Check lengths first, then if length matches check bytes match.
	CMPL(keyLen.As32(), Mem{Base: lengths, Index: i, Scale: 4})
	JNE(LabelRef("next"))
	maskX := XMM()
	mask := GP32()
	VPCMPEQB(Mem{Base: buffer}, key, maskX)
	VPMOVMSKB(maskX, mask)
	ANDL(match, mask)
	CMPL(match, mask)
	JE(LabelRef("done"))

	// Advance and loop again.
	Label("next")
	INCQ(i)
	ADDQ(Imm(16), buffer)
	JMP(LabelRef("loop"))

	// Return the loop increment, or the count if the key wasn't found.
	Label("done")
	Store(i, ReturnIndex(0))
	RET()
	Label("notfound")
	Store(count, ReturnIndex(0))
	RET()
}
