package asm

import (
	"encoding/binary"

	. "github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/operand"
)

func ConstBytes(name string, data []byte) operand.Mem {
	m := GLOBL(name, RODATA|NOPTR)

	switch {
	case len(data)%8 == 0:
		constBytes8(0, data)

	case len(data)%4 == 0:
		constBytes4(0, data)

	default:
		i := (len(data) / 8) * 8
		constBytes8(0, data[:i])
		constBytes1(i, data[i:])
	}

	return m
}

func ConstArray16(name string, elems ...uint16) operand.Mem {
	data := make([]byte, 2*len(elems))
	for i, elem := range elems {
		binary.LittleEndian.PutUint16(data[i*2:], elem)
	}
	return ConstBytes(name, data)
}

func ConstArray32(name string, elems ...uint32) operand.Mem {
	data := make([]byte, 4*len(elems))
	for i, elem := range elems {
		binary.LittleEndian.PutUint32(data[i*4:], elem)
	}
	return ConstBytes(name, data)
}

func ConstArray64(name string, elems ...uint64) operand.Mem {
	data := make([]byte, 8*len(elems))
	for i, elem := range elems {
		binary.LittleEndian.PutUint64(data[i*8:], elem)
	}
	return ConstBytes(name, data)
}

func ConstShuffleMask32(name string, indices ...uint32) operand.Mem {
	data := make([]byte, 4*len(indices))
	for i, index := range indices {
		for j := 0; j < 4; j++ {
			data[i*4+j] = byte(index*4 + uint32(j))
		}
	}
	return ConstBytes(name, data)
}

func ConstShuffleMask64(name string, indices ...uint64) operand.Mem {
	data := make([]byte, 8*len(indices))
	for i, index := range indices {
		for j := 0; j < 8; j++ {
			data[i*8+j] = byte(index*8 + uint64(j))
		}
	}
	return ConstBytes(name, data)
}

func ConstLoadMask32(name string, indices ...uint32) operand.Mem {
	data := make([]uint32, len(indices))
	for i, index := range indices {
		data[i] = index << 31
	}
	return ConstArray32(name, data...)
}

func ConstLoadMask64(name string, indices ...uint64) operand.Mem {
	data := make([]uint64, len(indices))
	for i, index := range indices {
		data[i] = index << 63
	}
	return ConstArray64(name, data...)
}

func constBytes8(offset int, data []byte) {
	for i := 0; i < len(data); i += 8 {
		DATA(offset+i, operand.U64(binary.LittleEndian.Uint64(data[i:i+8])))
	}
}

func constBytes4(offset int, data []byte) {
	for i := 0; i < len(data); i += 4 {
		DATA(offset+i, operand.U32(binary.LittleEndian.Uint32(data[i:i+4])))
	}
}

func constBytes1(offset int, data []byte) {
	for i, b := range data {
		DATA(offset+i, operand.U8(b))
	}
}
