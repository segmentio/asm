package amd64

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"

	"math"
	"math/bits"

	"github.com/segmentio/asm/cpu"
)

// JumpIfFeature constructs a jump sequence that tests for one or more feature flags.
// If all flags are matched, jump to the target label.
func JumpIfFeature(jmp string, f cpu.X86Feature) {
	jump(LabelRef(jmp), f, false)
}

// JumpUnlessFeature constructs a jump sequence that tests for one or more feature flags.
// Unless all flags are matched, jump to the target label.
func JumpUnlessFeature(jmp string, f cpu.X86Feature) {
	jump(LabelRef(jmp), f, true)
}

// cpuAddr is a Mem operand containing the global symbolic reference to the
// X86 cpu feature flags.
var cpuAddr = NewDataAddr(Symbol{Name: "github·com∕segmentio∕asm∕cpu·X86"}, 0)

func jump(jmp Op, f cpu.X86Feature, invert bool) {
	if bits.OnesCount64(uint64(f)) == 1 {
		// If the feature test is for a single flag, optimize the test using BTQ
		jumpSingleFlag(jmp, f, invert)
	} else {
		jumpMultiFlag(jmp, f, invert)
	}
}

func jumpSingleFlag(jmp Op, f cpu.X86Feature, invert bool) {
	bit := U8(bits.TrailingZeros64(uint64(f)))

	// Likely only need lower 4 bytes
	if bit < 32 {
		BTL(bit, cpuAddr)
	} else {
		BTQ(bit, cpuAddr)
	}

	if invert {
		JCC(jmp)
	} else {
		JCS(jmp)
	}
}

func jumpMultiFlag(jmp Op, f cpu.X86Feature, invert bool) {
	r := GP64()
	MOVQ(cpuAddr, r)

	var op Op
	switch {
	case f <= math.MaxUint8:
		op = U8(f)
	case f <= math.MaxUint32:
		op = U32(f)
	default:
		op = GP64()
		MOVQ(U64(f), op)
	}

	ANDQ(op, r)
	CMPQ(r, op)

	if invert {
		JNE(jmp)
	} else {
		JEQ(jmp)
	}
}
