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

	x86.VariableLengthBytes([]Register{src, dst}, n, func (length int, regs []Register, mem x86.Memory) {
		src, dst := regs[0], regs[1]

		switch length {
		case 1:
			v := GP8()
			MOVB(mem.Get(src), v)
			MOVB(v, mem.Get(dst))
		case 2:
			v := GP16()
			MOVW(mem.Get(src), v)
			MOVW(v, mem.Get(dst))
		case 4:
			v := GP32()
			MOVL(mem.Get(src), v)
			MOVL(v, mem.Get(dst))
		case 8:
			v := GP64()
			MOVQ(mem.Get(src), v)
			MOVQ(v, mem.Get(dst))
		case 16:
			v := XMM()
			VMOVUPS(mem.Get(src), v)
			VMOVUPS(v, mem.Get(dst))
		case 32:
			v := YMM()
			VMOVUPS(mem.Get(src), v)
			VMOVUPS(v, mem.Get(dst))
		}
	})
}
