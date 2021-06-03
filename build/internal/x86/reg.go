package x86

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

var all = map[Spec][]VecPhysical{
	S128: []VecPhysical{X0, X1, X2, X3, X4, X5, X6, X7, X8, X9, X10, X11, X12, X13, X14, X15, X16, X17, X18, X19, X20, X21, X22, X23, X24, X25, X26, X27, X28, X29, X30, X31},
	S256: []VecPhysical{Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7, Y8, Y9, Y10, Y11, Y12, Y13, Y14, Y15, Y16, Y17, Y18, Y19, Y20, Y21, Y22, Y23, Y24, Y25, Y26, Y27, Y28, Y29, Y30, Y31},
	S512: []VecPhysical{Z0, Z1, Z2, Z3, Z4, Z5, Z6, Z7, Z8, Z9, Z10, Z11, Z12, Z13, Z14, Z15, Z16, Z17, Z18, Z19, Z20, Z21, Z22, Z23, Z24, Z25, Z26, Z27, Z28, Z29, Z30, Z31},
}

func gp(c func() GPVirtual, g ...Register) Register {
	if len(g) == 0 {
		return c()
	}
	return g[0]
}

// load either emits a size-specific MOV operation based on the input immediate
// value, or returns the existing register. The optional destination register
// must be a general purpose register if provided. If not provided, a virtual
// register will be allocated if needed.
func load(ir Op, dst ...Register) (Register, uint) {
	switch v := ir.(type) {
	default:
		panic("unsupported input operand")
	case U8, I8:
		g := gp(GP32, dst...)
		MOVB(v, g.(GP).As8())
		return g, 1
	case U16, I16:
		g := gp(GP32, dst...)
		MOVW(v, g.(GP).As16())
		return g, 2
	case U32, I32:
		g := gp(GP32, dst...)
		MOVL(v, g)
		return g, 4
	case U64, I64:
		g := gp(GP64, dst...)
		MOVQ(v, g)
		return g, 8
	case Register:
		return v, v.Size()
	}
}

// VecList returns a slice of vector registers for the given Spec.
func VecList(s Spec, max int) []VecPhysical {
	return all[s][:max]
}

// VecBroadcast broadcasts an immediate or general purpose register into a
// vector register. The broadcast size is based on the input operand size.
// If the input is a register, it may be necessary to convert it to the
// desired size. For example:
//    reg := GP32()
//    XORL(reg, reg)
//    MOVB(U8(0x7F), reg.As8())
//    mask := VecBroadcast(reg, XMM())       // will broadcast 0x0000007F0000007F0000007F0000007F
//    mask := VecBroadcast(reg.As8(), XMM()) // will broadcast 0x7F7F7F7F7F7F7F7F7F7F7F7F7F7F7F7F
//
// If the `reg` register isn't needed, it would preferrable to use:
//    mask := VecBroadcast(U8(0x7F), XMM())
func VecBroadcast(ir Op, xyz Register) Register {
	vec := xyz.(Vec)
	reg, size := load(ir)

	// PINSR{B,W} accept either m{8,16} or r32. If the input was
	// r{8,16} we need to cast to 32 bits.
	if reg.Size() < 4 {
		if gp, ok := reg.(GP); ok {
			reg = gp.As32()
		} else {
			r32 := GP32()
			switch reg.Size() {
			case 1:
				MOVBLZX(reg, r32)
			case 2:
				MOVWLZX(reg, r32)
			}
			reg = r32
		}
	}

	switch size {
	default:
		panic("unsupported register size")
	case 1:
		PINSRB(Imm(0), reg, vec.AsX())
		VPBROADCASTB(vec.AsX(), xyz)
	case 2:
		PINSRW(Imm(0), reg, vec.AsX())
		VPBROADCASTW(vec.AsX(), xyz)
	case 4:
		PINSRD(Imm(0), reg, vec.AsX())
		VPBROADCASTD(vec.AsX(), xyz)
	case 8:
		PINSRQ(Imm(0), reg, vec.AsX())
		VPBROADCASTQ(vec.AsX(), xyz)
	}

	return xyz
}

// VectorLane is an interface for abstracting allocating and loading memory into
// vector registers. This is used during the map phase of the Vectorizer.
type VectorLane interface {
	Read(Mem) Register
	Offset(Mem) Mem
	Alloc() Register
}

// Vectorizer is a map/reduce-based helper for constructing parallelized
// instruction pipelines.
type Vectorizer struct {
	max     int                          // total registers allowed
	mapper  func(VectorLane) Register    // function to map the main operation to an output register
	reducer func(a, b Register) Register // function to reduce the mapped output registers into one
}

// NewVectorizer creates a new vectorizing utility utilizing a max number of
// registers and a mapper function.
func NewVectorizer(max int, mapper func(VectorLane) Register) *Vectorizer {
	return &Vectorizer{max: max, mapper: mapper}
}

// Reduce sets the reducer function in the Vectorizer.
func (v *Vectorizer) Reduce(h func(a, b Register) Register) *Vectorizer {
	v.reducer = h
	return v
}

// Compile runs the map and reduce phases for the given register size and
// parallel lane count. This can be called multiple times using different
// configurations to produce separate execution strides. The returned slice
// is dependent on the presence of a reducer. If no reducer is used, the
// slice will be all of the output registers from the map phase. If a reducer
// is defined, the result is slice containing the final reduced register.
func (v *Vectorizer) Compile(spec Spec, lanes int) []Register {
	alloc := NewVectorAlloc(VecList(spec, v.max), lanes)
	var out []Register

	for alloc.NextLane() {
		r := v.mapper(alloc)
		out = append(out, r)
	}

	if v.reducer != nil {
		for len(out) > 1 {
			r := v.reducer(out[0], out[1])
			out = append(out[2:], r)
		}
	}

	return out
}

// ReduceOr performs a bitwise OR between two registers and returns the result.
// This can be used as the reducer for a Vectorizer.
func ReduceOr(a, b Register) Register {
	VPOR(b, a, a)
	return a
}

// ReduceAnd performs a bitwise AND between two registers and returns the result.
// This can be used as the reducer for a Vectorizer.
func ReduceAnd(a, b Register) Register {
	VPAND(b, a, a)
	return a
}

// VectorAlloc is a lower-level lane-driven register allocator. This pulls
// registers from a fixed list of physical registers for a given number of
// lanes. Registers are allocated in distinct blocks; one block for each lane.
type VectorAlloc struct {
	vec   []VecPhysical   // all available physical registers
	rd    map[Mem]vecRead // loaded register index
	off   map[Mem]int     // offset index
	lanes int             // number of lanes being compiled
	lane  int             // current lane being compiled
	size  int             // register size
}

type vecRead struct {
	reg []Register
	idx int
}

// NewVectorAlloc creates a new VectorAlloc instance.
func NewVectorAlloc(vec []VecPhysical, lanes int) *VectorAlloc {
	return &VectorAlloc{
		vec:   vec,
		rd:    map[Mem]vecRead{},
		off:   map[Mem]int{},
		lanes: lanes,
		lane:  -1,
		size:  int(vec[0].Size()),
	}
}

// NextLane is used to advance the allocator to the next lane available lane.
// This returns false when no more lanes are available.
func (a *VectorAlloc) NextLane() bool {
	next := a.lane + 1
	if next < a.lanes {
		a.lane = next
		return true
	}
	return false
}

// Read implements the VectorLane interface. This reads the next register-sized
// memory region. Each read within a single lane will load adjacent memory
// regions. Subsequent lanes will read adjacent memory after the last read of
// the prior lane. This has special handling so that all reads are batched
// together. Because of this, Read calls should appear first in the mapper.
func (a *VectorAlloc) Read(mem Mem) Register {
	if _, ok := a.off[mem]; ok {
		panic("Offset and Read cannot current be combined for the same memory region")
	}

	rd := a.rd[mem]

	if a.lane == 0 {
		for i := 0; i < a.lanes; i++ {
			r := a.Alloc()
			VMOVDQU(mem.Offset(len(rd.reg)*a.size), r)
			rd.reg = append(rd.reg, r)
		}
	}

	r := rd.reg[rd.idx]
	rd.idx++

	a.rd[mem] = rd

	return r
}

// Offset implements the VectorLane interface. This calculates the next
// register-sized memory offset. Each offset within a single lane will refer
// to adjacent memory regions. Subsequent lanes will obtain an offset of
// adjacent memory after the last offset of the prior lane.
func (a *VectorAlloc) Offset(mem Mem) Mem {
	if _, ok := a.rd[mem]; ok {
		panic("Read and Offset cannot current be combined for the same memory region")
	}

	n := a.off[mem]
	a.off[mem] = n + 1

	return mem.Offset(n * a.size)
}

// Alloc implements the VectorLane interface. This allocates a register for the
// current lane.
func (a *VectorAlloc) Alloc() Register {
	if len(a.vec) == 0 {
		panic("not enough vector registers available")
	}

	r := a.vec[0]
	a.vec = a.vec[1:]

	return r
}
