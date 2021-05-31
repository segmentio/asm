package x86

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"

	"github.com/segmentio/asm/cpu"
)

type Memory struct {
	Size   int
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

func (m Memory) Load(base Register) Register {
	r := GetRegister(m.Size)
	m.mov(m.Get(base), r)
	return r
}

func (m Memory) Store(src Register, base Register) {
	m.mov(src, m.Get(base))
}

func (m Memory) mov(src, dst Op) {
	switch m.Size {
	case 1:
		MOVB(src, dst)
	case 2:
		MOVW(src, dst)
	case 4:
		MOVL(src, dst)
	case 8:
		MOVQ(src, dst)
	case 16:
		MOVOU(src, dst)
	case 32:
		VMOVUPS(src, dst)
	}
}

func GetRegister(size int) (r Register) {
	switch size {
	case 1:
		r = GP8()
	case 2:
		r = GP16()
	case 4:
		r = GP32()
	case 8:
		r = GP64()
	case 16:
		r = XMM()
	case 32:
		r = YMM()
	default:
		panic("bad register size")
	}
	return
}

func VariableLengthBytes(inputs []Register, n Register, handle func(inputs []Register, memory ...Memory)) {
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
	handle(inputs, Memory{Size: 8})
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
	handle(inputs,
		Memory{Size: 1},
		Memory{Size: 1, Index: n, Offset: -1})
	RET()

	Label("handle3")
	handle(inputs,
		Memory{Size: 2},
		Memory{Size: 1, Offset: 2})
	RET()

	Label("handle4")
	handle(inputs, Memory{Size: 4})
	RET()

	Label("handle5to7")
	handle(inputs,
		Memory{Size: 4},
		Memory{Size: 4, Index: n, Offset: -4})
	RET()

	Label("handle8")
	handle(inputs, Memory{Size: 8})
	RET()

	Label("handle9to16")
	handle(inputs,
		Memory{Size: 8},
		Memory{Size: 8, Index: n, Offset: -8})
	RET()

	Label("handle17to32")
	handle(inputs,
		Memory{Size: 16},
		Memory{Size: 16, Index: n, Offset: -16})
	RET()

	Label("handle33to64")
	handle(inputs,
		Memory{Size: 16},
		Memory{Size: 16, Offset: 16},
		Memory{Size: 16, Index: n, Offset: -32},
		Memory{Size: 16, Index: n, Offset: -16})
	RET()

	Comment("AVX optimized version for medium to large size inputs.")
	Label("avx2")
	CMPQ(n, Imm(128))
	JB(LabelRef("avx2_tail"))
	handle(inputs,
		Memory{Size: 32},
		Memory{Size: 32, Offset: 32},
		Memory{Size: 32, Offset: 64},
		Memory{Size: 32, Offset: 96})
	for i := range inputs {
		ADDQ(Imm(128), inputs[i])
	}
	SUBQ(Imm(128), n)
	JMP(LabelRef("avx2"))

	Label("avx2_tail")
	JZ(LabelRef("done")) // check flags from last CMPQ
	CMPQ(n, Imm(64)) // n > 0 && n <= 64
	JBE(LabelRef("avx2_tail_1to64"))
	handle(inputs,
		Memory{Size: 32},
		Memory{Size: 32, Offset: 32},
		Memory{Size: 32, Offset: 64},
		Memory{Size: 32, Index: n, Offset: -32})
	RET()

	Label("avx2_tail_1to64")
	handle(inputs,
		Memory{Size: 32, Index: n, Offset: -64},
		Memory{Size: 32, Index: n, Offset: -32})
	RET()

	Generate()
}
