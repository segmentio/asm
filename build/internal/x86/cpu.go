package x86

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
	JumpIfFeatureABI(jmp, f, cpuAddr)
}

func JumpIfFeatureABI(jmp string, f cpu.X86Feature, abi Op) {
	jump(abi, LabelRef(jmp), f, false)
}

// JumpUnlessFeature constructs a jump sequence that tests for one or more feature flags.
// Unless all flags are matched, jump to the target label.
func JumpUnlessFeature(jmp string, f cpu.X86Feature) {
	JumpUnlessFeatureABI(jmp, f, cpuAddr)
}

func JumpUnlessFeatureABI(jmp string, f cpu.X86Feature, abi Op) {
	jump(abi, LabelRef(jmp), f, true)
}

// cpuAddr is a Mem operand containing the global symbolic reference to the
// X86 cpu feature flags.
var cpuAddr = NewDataAddr(Symbol{Name: "github·com∕segmentio∕asm∕cpu·X86"}, 0)

func jump(abi, jmp Op, f cpu.X86Feature, invert bool) {
	if bits.OnesCount64(uint64(f)) == 1 {
		// If the feature test is for a single flag, optimize the test using BTQ
		jumpSingleFlag(abi, jmp, f, invert)
	} else {
		jumpMultiFlag(abi, jmp, f, invert)
	}
}

func jumpSingleFlag(abi, jmp Op, f cpu.X86Feature, invert bool) {
	bit := U8(bits.TrailingZeros64(uint64(f)))

	// Likely only need lower 4 bytes
	//if bit < 32 {
	//	BTL(bit, abi)
	//} else {
	BTQ(bit, abi)
	//}

	if invert {
		JCC(jmp)
	} else {
		JCS(jmp)
	}
}

func jumpMultiFlag(abi, jmp Op, f cpu.X86Feature, invert bool) {
	r := GP64()
	MOVQ(abi, r)

	var op Op
	if f <= math.MaxUint32 {
		op = U32(f)
	} else {
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
