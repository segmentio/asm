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

	x86.VariableLengthBytes([]Register{src, dst}, n, func (regs []Register, memory ...x86.Memory) {
		src, dst := regs[0], regs[1]

		tmp := make([]Register, len(memory))
		for i, m := range memory {
			tmp[i] = m.Load(src)
		}
		for i, m := range memory {
			m.Store(tmp[i], dst)
		}
	})
}
