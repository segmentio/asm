// / +build ignore

package main

import (
	"fmt"
	"math"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/asm"
)

const (
	unroll    = 4
	pageSize  = 4096
	maxLength = 16
)

func init() {
	ConstraintExpr("!purego")
}

func main() {
	Lookup()

	Generate()
}

// Lookup searches for a key in a set of keys.
//
// Each key in the set of keys should be padded to 16 bytes and concatenated
// into a single buffer. The routine searches for the input key in the set of
// keys and returns its index if found. If not found, the routine returns the
// number of keys (len(keyset)/16).
func Lookup() {
	TEXT("Lookup", NOSPLIT, "func(keyset []byte, key []byte) int")
	Doc("Lookup searches for a key in a set of keys, returning its index if ",
		"found. If the key cannot be found, the number of keys is returned.")

	// Load inputs.
	keyset := Load(Param("keyset").Base(), GP64())
	count := Load(Param("keyset").Len(), GP64())
	SHRQ(Imm(4), count)
	keyPtr := Load(Param("key").Base(), GP64())
	keyLen := Load(Param("key").Len(), GP64())
	keyCap := Load(Param("key").Cap(), GP64())

	// None of the keys are larger than maxLength.
	CMPQ(keyLen, Imm(maxLength))
	JA(LabelRef("not_found"))

	// We're going to be unconditionally loading 16 bytes from the input key
	// so first check if it's safe to do so (cap >= 16). If not, defer to
	// safe_load for additional checks.
	CMPQ(keyCap, Imm(maxLength))
	JB(LabelRef("safe_load"))

	// Load the input key and pad with zeroes to 16 bytes.
	Label("load")
	key := XMM()
	VMOVUPS(Mem{Base: keyPtr}, key)
	Label("prepare")
	zeroes := XMM()
	VPXOR(zeroes, zeroes, zeroes)
	ones := XMM()
	VPCMPEQB(ones, ones, ones)
	var blendBytes [maxLength * 2]byte
	for j := 0; j < maxLength; j++ {
		blendBytes[j] = 0xFF
	}
	blendMasks := ConstBytes("blend_masks", blendBytes[:])
	blendMasksPtr := GP64()
	LEAQ(blendMasks.Offset(maxLength), blendMasksPtr)
	SUBQ(keyLen, blendMasksPtr)
	blend := XMM()
	VMOVUPS(Mem{Base: blendMasksPtr}, blend)
	VPBLENDVB(blend, key, zeroes, key)

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

	// Loop over multiple keys in the big loop.
	Label("bigloop")
	CMPQ(i, truncatedCount)
	JE(LabelRef("loop"))

	x := []VecPhysical{X8, X9, X10, X11, X12, X13, X14, X15}
	for n := 0; n < unroll; n++ {
		VPCMPEQB(Mem{Base: keyset, Disp: maxLength * n}, key, x[n])
		VPTEST(ones, x[n])
		var target string
		if n == 0 {
			target = "done"
		} else {
			target = fmt.Sprintf("found%d", n)
		}
		JCS(LabelRef(target))
	}

	// Advance and loop again.
	ADDQ(Imm(unroll), i)
	ADDQ(Imm(unroll*maxLength), keyset)
	JMP(LabelRef("bigloop"))

	// Loop over the remaining keys.
	Label("loop")
	CMPQ(i, count)
	JE(LabelRef("done"))

	// Try to match against the input key.
	match := XMM()
	VPCMPEQB(Mem{Base: keyset}, key, match)
	VPTEST(ones, match)
	JCS(LabelRef("done"))

	// Advance and loop again.
	Label("next")
	INCQ(i)
	ADDQ(Imm(maxLength), keyset)
	JMP(LabelRef("loop"))
	JMP(LabelRef("done"))

	// Return the loop increment, or the count if the key wasn't found. If we're
	// here from a jump within the big loop, the loop increment needs
	// correcting first.
	for j := unroll - 1; j > 0; j-- {
		Label(fmt.Sprintf("found%d", j))
		INCQ(i)
	}
	Label("done")
	Store(i, ReturnIndex(0))
	RET()
	Label("not_found")
	Store(count, ReturnIndex(0))
	RET()

	// If the input key is near a page boundary, we must change the way we load
	// it to avoid a fault. We instead want to load the 16 bytes up to and
	// including the key, then shuffle the key forward in the register. E.g. for
	// key "foo" we would load the 13 bytes prior to the key along with "foo"
	// and then move the last 3 bytes forward so the first 3 bytes are equal
	// to "foo".
	Label("safe_load")
	pageOffset := GP64()
	MOVQ(keyPtr, pageOffset)
	ANDQ(U32(pageSize-1), pageOffset)
	CMPQ(pageOffset, U32(pageSize-maxLength))
	JBE(LabelRef("load")) // Not near a page boundary.
	offset := GP64()
	MOVQ(^U64(0)-maxLength+1, offset)
	ADDQ(keyLen, offset)
	VMOVUPS(Mem{Base: keyPtr, Index: offset, Scale: 1}, key)
	var shuffleBytes [maxLength * 2]byte
	for j := 0; j < maxLength; j++ {
		shuffleBytes[j] = byte(j)
		shuffleBytes[j+maxLength] = byte(j)
	}
	shuffleMasks := ConstBytes("shuffle_masks", shuffleBytes[:])
	shuffleMasksPtr := GP64()
	LEAQ(shuffleMasks.Offset(maxLength), shuffleMasksPtr)
	SUBQ(keyLen, shuffleMasksPtr)
	shuffle := XMM()
	VMOVUPS(Mem{Base: shuffleMasksPtr}, shuffle)
	VPSHUFB(shuffle, key, key)
	JMP(LabelRef("prepare"))
}
