/// +build ignore

package main

import (
	"fmt"
	"math"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/asm"
)

const unroll = 4

const pageSize = 4096

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
	keyCap := Load(Param("key").Cap(), GP64())

	// None of the keys we're searching through have a length greater than 16,
	// so bail early if the input is more than 16 bytes long.
	CMPQ(keyLen.As64(), Imm(16))
	JA(LabelRef("notfound"))

	// Load the remaining inputs.
	buffer := Load(Param("buffer"), GP64())
	lengths := Load(Param("lengths").Base(), GP64())
	count := Load(Param("lengths").Len(), GP64())

	// Load the input key. We're going to be unconditionally loading 16 bytes,
	// so first check if it's safe to do so (cap(k) >= 16). If not, and we're
	// near a page boundary, we must load+shuffle to avoid a fault.
	CMPQ(keyCap, Imm(16))
	JB(LabelRef("check_input"))
	Label("load")
	key := XMM()
	VMOVUPS(Mem{Base: keyPtr}, key)
	Label("start")

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
		JNE(LabelRef(fmt.Sprintf("try%d", n+1)))
		VPCMPEQB(Mem{Base: buffer, Disp: 16 * n}, key, x[n])
		VPMOVMSKB(x[n], g[n])
		ANDL(match, g[n])
		CMPL(match, g[n])
		JNE(LabelRef(fmt.Sprintf("try%d", n+1)))
		if n > 0 {
			// Correct the loop increment before returning.
			ADDQ(Imm(uint64(n)), i)
		}
		JMP(LabelRef("done"))
	}

	// Advance and loop again.
	Label(fmt.Sprintf("try%d", unroll))
	ADDQ(Imm(unroll), i)
	ADDQ(Imm(16*unroll), buffer)
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

	Label("check_input")
	pageOffset := GP64()
	MOVQ(keyPtr, pageOffset)
	ANDQ(U32(pageSize-1), pageOffset)
	CMPQ(pageOffset, U32(pageSize-16))
	JBE(LabelRef("load"))

	// If the input key is near a page boundary, we instead want to load the
	// 16 bytes up to and including the key, then shuffle the key forward in the
	// register. E.g. for key "foo" we would load the 13 bytes prior to the key
	// along with "foo" and then move the last 3 bytes forward so the first 3
	// bytes are equal to "foo".
	Label("tail_load")
	offset := GP64()
	MOVQ(^U64(0)-16+1, offset)
	ADDQ(keyLen.As64(), offset)
	VMOVUPS(Mem{Base: keyPtr, Index: offset, Scale: 1}, key)

	var shuffleBytes [16 * 2]byte
	for j := 0; j < 16; j++ {
		shuffleBytes[j] = byte(j)
		shuffleBytes[j+16] = byte(j)
	}
	shuffleMasks := ConstBytes("shuffle_masks", shuffleBytes[:])
	shuffleMasksPtr := GP64()
	LEAQ(shuffleMasks.Offset(16), shuffleMasksPtr)
	SUBQ(keyLen.As64(), shuffleMasksPtr)
	shuffle := XMM()
	VMOVUPS(Mem{Base: shuffleMasksPtr}, shuffle)
	VPSHUFB(shuffle, key, key)

	JMP(LabelRef("start"))
}
