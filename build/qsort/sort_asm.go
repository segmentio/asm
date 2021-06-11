// +build !amd64

package main

import (
	"fmt"
	"math"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
	. "github.com/segmentio/asm/build/internal/x86"
)

func main() {
	distributeForward(&SortableScalar{reg: GP64, size: 8, mov: MOVQ, cmp: CMPQ})
	distributeBackward(&SortableScalar{reg: GP64, size: 8, mov: MOVQ, cmp: CMPQ})

	insertionsort(&SortableVector{reg: XMM, size: 16})
	distributeForward(&SortableVector{reg: XMM, size: 16})
	distributeBackward(&SortableVector{reg: XMM, size: 16})

	insertionsort(&SortableVector{reg: YMM, size: 32})
	distributeForward(&SortableVector{reg: YMM, size: 32})
	distributeBackward(&SortableVector{reg: YMM, size: 32})

	Generate()
}

type Sortable interface {
	Register() Register
	Size() uint64
	Init()
	Move(Op, Op)
	Compare(Register, Register)
}

type SortableScalar struct {
	reg  func() GPVirtual
	size uint64
	mov  func(Op, Op)
	cmp  func(Op, Op)
}

func (s *SortableScalar) Register() Register {
	return s.reg()
}

func (s *SortableScalar) Size() uint64 {
	return s.size
}

func (s *SortableScalar) Init() {}

func (s *SortableScalar) Move(a, b Op) {
	s.mov(a, b)
}

func (s *SortableScalar) Compare(a, b Register) {
	s.cmp(a, b)
}

type SortableVector struct {
	reg  func() VecVirtual
	size uint64
	msb  Register
}

func (s *SortableVector) Register() Register {
	return s.reg()
}

func (s *SortableVector) Size() uint64 {
	return s.size
}

func (s *SortableVector) Init() {
	s.msb = s.reg()
	VecBroadcast(U64(1<<63), s.msb)
}

func (s *SortableVector) Move(a, b Op) {
	VMOVDQU(a, b)
}

func (s *SortableVector) Compare(a, b Register) {
	// The following is a routine for vectors that yields the same ZF/CF
	// result as a CMP instruction.

	// First compare each packed qword for equality.
	eq := s.Register()
	VPCMPEQQ(a, b, eq)

	// SSE4.2 and AVX2 have a CMPGTQ to compare packed qwords, but
	// unfortunately it's a signed comparison. We know that u64 has
	// range [0,2^64-1] and signed (two's complement) i64 has range
	// [-2^63,2^63-1]. We can add (or subtract) 2^63 to each packed
	// unsigned qword and reinterpret each as a signed qword. Doing so
	// allows us to utilize a signed comparison, and yields the same
	// result as if we were doing an unsigned comparison with the input.
	// As usual, AVX-512 fixes the problem with its VPCMPUQ.
	lt := s.Register()
	aSigned := s.Register()
	bSigned := s.Register()
	VPADDQ(a, s.msb, aSigned)
	VPADDQ(b, s.msb, bSigned)
	VPCMPGTQ(aSigned, bSigned, lt)

	// Extract bit masks.
	eqMask := GP32()
	ltMask := GP32()
	VMOVMSKPD(eq, eqMask)
	VMOVMSKPD(lt, ltMask)

	// Invert the equality mask to find qwords that weren't equal.
	// Bit-scan forward to find the first unequal byte, then test
	// that bit in the less-than mask.
	NOTL(eqMask)
	unequalByteIndex := GP32()
	BSFL(eqMask, unequalByteIndex) // set ZF
	BTSL(unequalByteIndex, ltMask) // set CF
}

func insertionsort(s Sortable) {
	size := s.Size()
	TEXT(fmt.Sprintf("insertionsort%dNoSwap", size*8), NOSPLIT, fmt.Sprintf("func(data []%s, base int, swap func(int, int))", typeFor(size)))
	Pragma("noescape")

	data := Load(Param("data").Base(), GP64())
	end := Load(Param("data").Len(), GP64())
	shift := log2(size)
	SHLQ(Imm(shift), end)
	ADDQ(data, end)
	TESTQ(data, end)
	JE(LabelRef("done"))

	s.Init()

	i := GP64()
	MOVQ(data, i)

	Label("outer")
	ADDQ(Imm(size), i)
	CMPQ(i, end)
	JAE(LabelRef("done"))
	item := s.Register()
	s.Move(Mem{Base: i}, item)
	j := GP64()
	MOVQ(i, j)

	Label("inner")
	prev := s.Register()
	s.Move(Mem{Base: j, Disp: -int(size)}, prev)
	s.Compare(item, prev)
	JAE(LabelRef("outer"))

	s.Move(prev, Mem{Base: j})
	s.Move(item, Mem{Base: j, Disp: -int(size)})
	SUBQ(Imm(size), j)
	CMPQ(j, data)
	JA(LabelRef("inner"))
	JMP(LabelRef("outer"))

	Label("done")
	if size > 16 {
		VZEROUPPER()
	}
	RET()
}

func distributeForward(s Sortable) {
	size := s.Size()
	TEXT(fmt.Sprintf("distributeForward%d", size*8), NOSPLIT, fmt.Sprintf("func(data, scratch []%s, limit, lo, hi int) int", typeFor(size)))
	Pragma("noescape")

	// Load inputs.
	data := Load(Param("data").Base(), GP64())
	scratch := Load(Param("scratch").Base(), GP64())
	limit := Load(Param("limit"), GP64())
	loIndex := Load(Param("lo"), GP64())
	hiIndex := Load(Param("hi"), GP64())

	// Convert indices to byte offsets if necessary. We can use indices
	// only if the size is a valid scale (1/2/4/8).
	shift := log2(size)
	var scale uint8
	if size <= 8 {
		scale = uint8(size)
	} else {
		scale = 1
		SHLQ(Imm(shift), limit)
		SHLQ(Imm(shift), loIndex)
		SHLQ(Imm(shift), hiIndex)
	}

	// Prepare read/cmp pointers.
	lo := GP64()
	hi := GP64()
	tail := GP64()
	LEAQ(Mem{Base: data, Index: loIndex, Scale: scale}, lo)
	LEAQ(Mem{Base: data, Index: hiIndex, Scale: scale}, hi)
	LEAQ(Mem{Base: scratch, Index: limit, Scale: scale, Disp: -int(size)}, tail)

	s.Init()

	// Load the pivot item.
	pivot := s.Register()
	s.Move(Mem{Base: data}, pivot)

	offset := GP64()
	isLess := GP64()
	XORQ(offset, offset)
	XORQ(isLess, isLess)

	// We'll be keeping a negative offset. Negate the limit so we can
	// compare the two in the loop.
	NEGQ(limit)

	Label("loop")

	// Load the next item.
	next := s.Register()
	s.Move(Mem{Base: lo}, next)

	// Compare the item with the pivot.
	s.Compare(next, pivot)
	SETCS(isLess.As8())

	// Conditionally write to either the beginning of the data slice, or
	// end of the scratch slice.
	dst := GP64()
	MOVQ(lo, dst)
	CMOVQCC(tail, dst)
	s.Move(next, Mem{Base: dst, Index: offset, Scale: scale})
	XORB(Imm(1), isLess.As8())
	if size > 8 {
		SHLQ(Imm(shift), isLess)
	}
	SUBQ(isLess, offset)
	ADDQ(Imm(size), lo)

	// Loop while we have more input, and enough room in the scratch slice.
	CMPQ(lo, hi)
	JA(LabelRef("done"))
	CMPQ(offset, limit)
	JNE(LabelRef("loop"))

	// Return the number of items written to the data slice.
	Label("done")
	SUBQ(data, lo)
	LEAQ(Mem{Base: lo, Index: offset, Scale: scale}, lo)
	SHRQ(Imm(shift), lo)
	DECQ(lo)
	Store(lo, ReturnIndex(0))
	if size > 16 {
		VZEROUPPER()
	}
	RET()
}

func distributeBackward(s Sortable) {
	size := s.Size()
	TEXT(fmt.Sprintf("distributeBackward%d", size*8), NOSPLIT, fmt.Sprintf("func(data, scratch []%s, limit, lo, hi int) int", typeFor(size)))
	Pragma("noescape")

	// Load inputs.
	data := Load(Param("data").Base(), GP64())
	scratch := Load(Param("scratch").Base(), GP64())
	limit := Load(Param("limit"), GP64())
	loIndex := Load(Param("lo"), GP64())
	hiIndex := Load(Param("hi"), GP64())

	// Convert indices to byte offsets if necessary. We can use indices
	// only if the size is a valid scale (1/2/4/8).
	shift := log2(size)
	var scale uint8
	if size <= 8 {
		scale = uint8(size)
	} else {
		scale = 1
		SHLQ(Imm(shift), limit)
		SHLQ(Imm(shift), loIndex)
		SHLQ(Imm(shift), hiIndex)
	}

	// Prepare read/cmp pointers.
	lo := GP64()
	hi := GP64()
	LEAQ(Mem{Base: data, Index: loIndex, Scale: scale}, lo)
	LEAQ(Mem{Base: data, Index: hiIndex, Scale: scale}, hi)

	s.Init()

	// Load the pivot item.
	pivot := s.Register()
	s.Move(Mem{Base: data}, pivot)

	offset := GP64()
	isLess := GP64()
	XORQ(offset, offset)
	XORQ(isLess, isLess)

	CMPQ(hi, lo)
	JBE(LabelRef("done"))

	Label("loop")

	// Load the next item.
	next := s.Register()
	s.Move(Mem{Base: hi}, next)

	// Compare the item with the pivot.
	s.Compare(next, pivot)
	SETCS(isLess.As8())

	// Conditionally write to either the end of the data slice, or
	// beginning of the scratch slice.
	dst := GP64()
	MOVQ(scratch, dst)
	CMOVQCC(hi, dst)
	s.Move(next, Mem{Base: dst, Index: offset, Scale: scale})
	if size > 8 {
		SHLQ(Imm(shift), isLess)
	}
	ADDQ(isLess, offset)
	SUBQ(Imm(size), hi)

	// Loop while we have more input, and enough room in the scratch slice.
	CMPQ(hi, lo)
	JBE(LabelRef("done"))
	CMPQ(offset, limit)
	JNE(LabelRef("loop"))

	// Return the number of items written to the data slice.
	Label("done")
	SUBQ(data, hi)
	LEAQ(Mem{Base: hi, Index: offset, Scale: scale}, hi)
	SHRQ(Imm(shift), hi)
	Store(hi, ReturnIndex(0))
	if size > 16 {
		VZEROUPPER()
	}
	RET()
}

func log2(size uint64) uint64 {
	return uint64(math.Log2(float64(size)))
}

func typeFor(size uint64) string {
	if size <= 8 {
		return fmt.Sprintf("uint%d", size*8)
	}
	return fmt.Sprintf("[%d]uint64", size/8)
}
