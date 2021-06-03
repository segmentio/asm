// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/x86"
)

func main() {
	TEXT("ContainsByte", NOSPLIT, "func(haystack []byte, needle byte) bool")

	haystack := Load(Param("haystack").Base(), GP64())
	length := Load(Param("haystack").Len(), GP64())

	// Broadcast the needle byte to each 8 bytes in a GP64.
	needle := GP64()
	XORQ(needle, needle)
	Load(Param("needle"), needle.As8())
	tmp := GP64()
	for i := 3; i <= 5; i++ {
		MOVQ(needle, tmp)
		SHLQ(U8(1<<i), tmp)
		ORQ(tmp, needle)
	}

	// Prepare masks: one with LSB set in each byte, another
	// with MSB set in each byte, and another zeroed YMM register.
	lsb := GP64()
	msb := GP64()
	MOVQ(U64(0x0101010101010101), lsb)
	MOVQ(U64(0x8080808080808080), msb)

	ret, _ := ReturnIndex(0).Resolve()
	MOVB(U8(0), ret.Addr)
	JMP(LabelRef("start"))

	needleVec := YMM()
	zero := YMM()

	Label("found")
	MOVB(U8(1), ret.Addr)
	JMP(LabelRef("done"))

	Label("avx2_found")
	MOVB(U8(1), ret.Addr)
	JMP(LabelRef("avx2_done"))

	VariableLengthBytes([]Register{haystack}, length, VariableLengthBytesImpl{
		SetupXMM: func() {
			PXOR(zero.AsX(), zero.AsX())
			PINSRQ(Imm(0), needle, needleVec.AsX())
			PINSRQ(Imm(1), needle, needleVec.AsX())
		},
		SetupYMM: func() {
			VZEROUPPER()
			VPBROADCASTQ(needleVec.AsX(), needleVec)
		},
		Generate: func(inputs []Register, memory ...Memory) {
			haystack := inputs[0]

			regs := make([]Op, len(memory))
			for i, m := range memory {
				regs[i] = m.Load(haystack)
			}

			switch memory[0].Size {
			case 1:
				for i := range memory {
					CMPB(regs[i], needle.As8())
					JE(LabelRef("found"))
				}
			case 2:
				results := make([]Op, len(memory))
				for i := range memory {
					XORW(needle.As16(), regs[i])
					results[i] = GP16()
					MOVW(regs[i], results[i])
				}
				for i := range memory {
					SUBW(lsb.As16(), results[i])
					NOTW(regs[i])
					ANDW(regs[i], results[i])
				}
				result := reduce(results, binary(ORW))
				ANDW(msb.As16(), result)
				JNZ(LabelRef("found"))
			case 4:
				results := make([]Op, len(memory))
				for i := range memory {
					XORL(needle.As32(), regs[i])
					results[i] = GP32()
					MOVL(regs[i], results[i])
				}
				for i := range memory {
					SUBL(lsb.As32(), results[i])
					ANDNL(results[i], regs[i], results[i])
				}
				result := reduce(results, binary(ORL))
				ANDL(msb.As32(), result)
				JNZ(LabelRef("found"))
			case 8:
				results := make([]Op, len(memory))
				for i := range memory {
					XORQ(needle, regs[i])
					results[i] = GP64()
					MOVQ(regs[i], results[i])
				}
				for i := range memory {
					SUBQ(lsb, results[i])
					ANDNQ(results[i], regs[i], results[i])
				}
				result := reduce(results, binary(ORQ))
				ANDQ(msb, result)
				JNZ(LabelRef("found"))
			case 16:
				for i := range memory {
					PCMPEQB(needleVec.AsX(), regs[i])
				}
				result := reduce(regs, binary(POR))
				PTEST(result, zero.AsX())
				JCC(LabelRef("found"))
			case 32:
				for i := range memory {
					VPCMPEQB(needleVec, regs[i], regs[i])
				}
				result := reduce(regs, vex(VPOR))
				VPTEST(result, zero)
				JCC(LabelRef("avx2_found"))
			}
		},
	})
}

func reduce(ops []Op, op func(Op, Op) Op) Op {
	for len(ops) > 1 {
		ops = append(ops[2:], op(ops[0], ops[1]))
	}
	return ops[0]
}

func binary(ins func(Op, Op)) func(Op, Op) Op {
	return func(src Op, dst Op) Op {
		ins(src, dst)
		return dst
	}
}

func vex(ins func(Op, Op, Op)) func(Op, Op) Op {
	return func(src Op, dst Op) Op {
		ins(src, dst, dst)
		return dst
	}
}
