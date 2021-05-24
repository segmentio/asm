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

func VecList(s Spec, max int) []VecPhysical {
	return all[s][:max]
}

func VecBroadcast(ir Op, xyz Register) Register {
	vec := xyz.(Vec)
	var reg Register
	var size uint

	switch v := ir.(type) {
	default:
		panic("unsupported input operand")
	case U8, I8:
		g := GP32()
		MOVB(v, g.As8())
		reg = g
		size = 1
	case U16, I16:
		g := GP32()
		MOVW(v, g.As16())
		reg = g
		size = 2
	case U32, I32:
		g := GP32()
		MOVL(v, g)
		reg = g
		size = 4
	case U64, I64:
		g := GP64()
		MOVQ(v, g)
		reg = g
		size = 8
	case Register:
		reg = v
		size = v.Size()
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
