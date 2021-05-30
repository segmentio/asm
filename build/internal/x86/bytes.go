package x86

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"

	"github.com/segmentio/asm/cpu"
)

type Memory struct {
	Index  Register
	Offset int
}

func (m Memory) Get(base Register) Mem {
	memory := Mem{Base: base, Disp: m.Offset, Scale: 1}
	if m.Index != nil {
		memory.Index = m.Index
	}
	return memory
}

func VariableLengthBytes(inputs []Register, n Register, handle func(length int, inputs []Register, mem Memory)) {
	zero := GP64()
	XORQ(zero, zero)

	Label("tail")

	CMPQ(n, Imm(0))
	JE(LabelRef("done"))

	CMPQ(n, Imm(2))
	JBE(LabelRef("handle1to2"))

	CMPQ(n, Imm(3))
	JE(LabelRef("handle3"))

	CMPQ(n, Imm(4))
	JE(LabelRef("handle4"))

	CMPQ(n, Imm(8))
	JB(LabelRef("handle5to7"))
	JE(LabelRef("handle8"))

	CMPQ(n, Imm(16))
	JBE(LabelRef("handle9to16"))

	CMPQ(n, Imm(32))
	JBE(LabelRef("handle17to32"))

	CMPQ(n, Imm(64))
	JBE(LabelRef("handle33to64"))

	JumpIfFeature("avx2", cpu.AVX2)

	Label("generic")
	handle(8, inputs, Memory{})
	for i := range inputs {
		ADDQ(Imm(8), inputs[i])
	}
	SUBQ(Imm(8), n)
	CMPQ(n, Imm(8))
	JBE(LabelRef("tail"))
	JMP(LabelRef("generic"))

	Label("done")
	RET()

	Label("handle1to2")
	handle(1, inputs, Memory{})
	handle(1, inputs, Memory{Index: n, Offset: -1})
	RET()

	Label("handle3")
	handle(2, inputs, Memory{})
	handle(1, inputs, Memory{Offset: 2})
	RET()

	Label("handle4")
	handle(4, inputs, Memory{})
	RET()

	Label("handle5to7")
	handle(4, inputs, Memory{})
	handle(4, inputs, Memory{Index: n, Offset: -4})
	RET()

	Label("handle8")
	handle(8, inputs, Memory{})
	RET()

	Label("handle9to16")
	handle(8, inputs, Memory{})
	handle(8, inputs, Memory{Index: n, Offset: -8})
	RET()

	Label("handle17to32")
	handle(16, inputs, Memory{})
	handle(16, inputs, Memory{Index: n, Offset: -16})
	RET()

	Label("handle33to64")
	handle(32, inputs, Memory{})
	handle(32, inputs, Memory{Index: n, Offset: -32})
	RET()

	Comment("AVX optimized version for medium to large size inputs.")
	Label("avx2")
	CMPQ(n, Imm(128))
	JB(LabelRef("avx2_tail"))
	handle(32, inputs, Memory{})
	handle(32, inputs, Memory{Offset: 32})
	handle(32, inputs, Memory{Offset: 64})
	handle(32, inputs, Memory{Offset: 96})
	for i := range inputs {
		ADDQ(Imm(128), inputs[i])
	}
	SUBQ(Imm(128), n)
	JMP(LabelRef("avx2"))

	Label("avx2_tail")
	JZ(LabelRef("done")) // check flags from last CMPQ
	CMPQ(n, Imm(64)) // n > 0 && n <= 64
	JBE(LabelRef("avx2_tail_1to64"))
	handle(32, inputs, Memory{})
	handle(32, inputs, Memory{Offset: 32})
	handle(32, inputs, Memory{Offset: 64})
	handle(32, inputs, Memory{Index: n, Offset: -32})
	RET()

	Label("avx2_tail_1to64")
	handle(32, inputs, Memory{Index: n, Offset: -64})
	handle(32, inputs, Memory{Index: n, Offset: -32})
	RET()

	Generate()
}
