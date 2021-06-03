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

func (m Memory) Resolve(base Register) Mem {
	memory := Mem{Base: base, Disp: m.Offset, Scale: 1}
	if m.Index != nil {
		memory.Index = m.Index
	}
	return memory
}

func (m Memory) Load(base Register) Register {
	r := GetRegister(m.Size)
	m.mov(m.Resolve(base), r)
	return r
}

func (m Memory) Store(src Register, base Register) {
	m.mov(src, m.Resolve(base))
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
		VMOVDQU(src, dst)
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

func BinaryOpTable(B, W, L, Q, X func(Op, Op), VEX func(Op, Op, Op)) []func(Op, Op) {
	return []func(Op, Op){
		1:  B,
		2:  W,
		4:  L,
		8:  Q,
		16: X,
		32: func(src, dst Op) { VEX(src, dst, dst) },
	}
}

func GenerateCopy(name, doc string, transform []func(Op, Op)) {
	TEXT(name, NOSPLIT, "func(dst, src []byte) int")
	Doc(name + " " + doc)

	dst := Load(Param("dst").Base(), GP64())
	src := Load(Param("src").Base(), GP64())

	n := Load(Param("dst").Len(), GP64())
	x := Load(Param("src").Len(), GP64())

	CMPQ(x, n)
	CMOVQLT(x, n)
	Store(n, ReturnIndex(0))

	VariableLengthBytes{
		Unroll: 128,
		Process: func(regs []Register, memory ...Memory) {
			src, dst := regs[0], regs[1]

			count := len(memory)
			operands := make([]Op, count*2)

			for i, m := range memory {
				operands[i] = m.Load(src)
				if transform != nil {
					if m.Size == 32 {
						// For AVX2, avoid loading the destination into a register
						// before transforming it; pass the memory argument directly
						// to the transform instruction.
						operands[i+count] = m.Resolve(dst)
					} else {
						operands[i+count] = m.Load(dst)
					}
				}
			}

			if transform != nil {
				for i, m := range memory {
					transform[m.Size](operands[i+count], operands[i])
				}
			}

			for i, m := range memory {
				m.Store(operands[i].(Register), dst)
			}
		},
	}.Generate([]Register{src, dst}, n)
}

type VariableLengthBytes struct {
	SetupXMM func()
	SetupYMM func()
	Process  func(inputs []Register, memory ...Memory)
	Epilogue func()
	Unroll   int
}

func (v VariableLengthBytes) Generate(inputs []Register, n Register) {
	unroll := uint64(v.Unroll)
	if unroll != 128 && unroll != 256 {
		panic("unsupported unroll")
	}

	Label("start")

	if v.SetupXMM != nil {
		CMPQ(n, Imm(16))
		JBE(LabelRef("tail"))

		v.SetupXMM()
	}

	Label("tail")

	CMPQ(n, Imm(0))
	JE(LabelRef("done"))

	CMPQ(n, Imm(1))
	JE(LabelRef("handle1"))

	CMPQ(n, Imm(3))
	JBE(LabelRef("handle2to3"))

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

	JumpUnlessFeature("generic", cpu.AVX2)

	if v.SetupYMM != nil {
		v.SetupYMM()
	}

	CMPQ(n, U32(unroll))
	JB(LabelRef("avx2_tail"))
	JMP(LabelRef("avx2"))

	Label("generic")
	v.Process(inputs,
		Memory{Size: 16},
		Memory{Size: 16, Offset: 16},
		Memory{Size: 16, Offset: 32},
		Memory{Size: 16, Offset: 48},
	)
	for i := range inputs {
		ADDQ(Imm(64), inputs[i])
	}
	SUBQ(Imm(64), n)
	CMPQ(n, Imm(64))
	JBE(LabelRef("tail"))
	JMP(LabelRef("generic"))

	Label("done")
	RET()

	Label("handle1")
	v.Process(inputs, Memory{Size: 1})
	RET()

	Label("handle2to3")
	v.Process(inputs,
		Memory{Size: 2},
		Memory{Size: 2, Index: n, Offset: -2})
	RET()

	Label("handle4")
	v.Process(inputs, Memory{Size: 4})
	RET()

	Label("handle5to7")
	v.Process(inputs,
		Memory{Size: 4},
		Memory{Size: 4, Index: n, Offset: -4})
	RET()

	Label("handle8")
	v.Process(inputs, Memory{Size: 8})
	RET()

	Label("handle9to16")
	v.Process(inputs,
		Memory{Size: 8},
		Memory{Size: 8, Index: n, Offset: -8})
	RET()

	Label("handle17to32")
	v.Process(inputs,
		Memory{Size: 16},
		Memory{Size: 16, Index: n, Offset: -16})
	RET()

	Label("handle33to64")
	v.Process(inputs,
		Memory{Size: 16},
		Memory{Size: 16, Offset: 16},
		Memory{Size: 16, Index: n, Offset: -32},
		Memory{Size: 16, Index: n, Offset: -16})
	RET()

	// We have at least `unroll` bytes.
	Comment("AVX optimized version for medium to large size inputs.")
	Label("avx2")
	var memory []Memory
	for i := 0; i < int(unroll / 32); i++ {
		memory = append(memory, Memory{Size: 32, Offset: i * 32})
	}
	v.Process(inputs, memory...)
	for i := range inputs {
		ADDQ(U32(unroll), inputs[i])
	}
	SUBQ(U32(unroll), n)
	JZ(LabelRef("avx2_done"))
	CMPQ(n, U32(unroll))
	JAE(LabelRef("avx2"))

	// We have between [1, unroll) bytes.
	Label("avx2_tail")
	CMPQ(n, Imm(64))
	JB(LabelRef("avx2_tail_1to63"))
	v.Process(inputs,
		Memory{Size: 32},
		Memory{Size: 32, Offset: 32})
	for i := range inputs {
		ADDQ(Imm(64), inputs[i])
	}
	SUBQ(Imm(64), n)
	JMP(LabelRef("avx2_tail"))

	Label("avx2_tail_1to63")
	v.Process(inputs,
		Memory{Size: 32, Index: n, Offset: -64},
		Memory{Size: 32, Index: n, Offset: -32})

	Label("avx2_done")
	VZEROUPPER()
	RET()

	if v.Epilogue != nil {
		v.Epilogue()
	}

	Generate()
}
