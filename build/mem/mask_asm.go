// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"

	"github.com/segmentio/asm/build/internal/x86"
)

func main() {
	TEXT("Mask", NOSPLIT, "func(dst, src []byte) int")
	Doc("Mask set bits of dst to zero and copies the one-bits of src to dst, returning the number of bytes written.")

	dst := Load(Param("dst").Base(), GP64())
	src := Load(Param("src").Base(), GP64())

	n := Load(Param("dst").Len(), GP64())
	x := Load(Param("src").Len(), GP64())

	CMPQ(x, n)
	CMOVQGT(x, n)
	Store(n, ReturnIndex(0))

	x86.VariableLengthBytes([]Register{src, dst}, n, func(regs []Register, memory ...x86.Memory) {
		src, dst := regs[0], regs[1]

		count := len(memory)
		operands := make([]Op, count*2)

		for i, m := range memory {
			operands[i] = m.Load(src)
			if m.Size == 32 {
				operands[i+count] = m.Resolve(dst)
			} else {
				operands[i+count] = m.Load(dst)
			}
		}

		for i, m := range memory {
			if m.Size == 32 {
				VPAND(operands[i+count], operands[i], operands[i])
			} else {
				x86.BinaryOpTable(ANDB, ANDW, ANDL, ANDQ, PAND)[m.Size](operands[i+count], operands[i])
			}
		}

		for i, m := range memory {
			m.Store(operands[i].(Register), dst)
		}
	})
}
