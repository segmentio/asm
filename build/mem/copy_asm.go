// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/reg"

	"github.com/segmentio/asm/build/internal/x86"
)

func main() {
	TEXT("Copy", NOSPLIT, "func(dst, src []byte) int")
	Doc("Copy copies src to dst, returning the number of bytes written.")

	dst := Load(Param("dst").Base(), GP64())
	src := Load(Param("src").Base(), GP64())

	n := Load(Param("dst").Len(), GP64())
	x := Load(Param("src").Len(), GP64())

	CMPQ(x, n)
	CMOVQGT(x, n)
	Store(n, ReturnIndex(0))

	x86.VariableLengthBytes([]Register{src, dst}, n, func (length int, regs []Register, memory ...x86.Memory) {
		src, dst := regs[0], regs[1]

		reg := make([]Register, len(memory))

		switch length {
		case 1:
			for i, m := range memory {
				reg[i] = GP8()
				MOVB(m.Get(src), reg[i])
			}
			for i, m := range memory {
				MOVB(reg[i], m.Get(dst))
			}
		case 2:
			for i, m := range memory {
				reg[i] = GP16()
				MOVW(m.Get(src), reg[i])
			}
			for i, m := range memory {
				MOVW(reg[i], m.Get(dst))
			}
		case 4:
			for i, m := range memory {
				reg[i] = GP32()
				MOVL(m.Get(src), reg[i])
			}
			for i, m := range memory {
				MOVL(reg[i], m.Get(dst))
			}
		case 8:
			for i, m := range memory {
				reg[i] = GP64()
				MOVQ(m.Get(src), reg[i])
			}
			for i, m := range memory {
				MOVQ(reg[i], m.Get(dst))
			}
		case 16:
			for i, m := range memory {
				reg[i] = XMM()
				VMOVUPS(m.Get(src), reg[i])
			}
			for i, m := range memory {
				VMOVUPS(reg[i], m.Get(dst))
			}
		case 32:
			for i, m := range memory {
				reg[i] = YMM()
				VMOVUPS(m.Get(src), reg[i])
			}
			for i, m := range memory {
				VMOVUPS(reg[i], m.Get(dst))
			}
		}
	})
}
