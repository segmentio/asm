/// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

func init() {
	ConstraintExpr("!purego")
}

func main() {
	TEXT("search16", NOSPLIT, "func(buffer *byte, lengths []uint32, key []byte) int")

	keyPtr := Load(Param("key").Base(), GP64())
	keyLen := reg.CX
	Load(Param("key").Len(), keyLen.As64())

	CMPQ(keyLen.As64(), Imm(16))
	JA(LabelRef("notfound"))

	buffer := Load(Param("buffer"), GP64())

	lengths := Load(Param("lengths").Base(), GP64())
	count := Load(Param("lengths").Len(), GP64())

	// FIXME: check if near a page boundary and load+shuffle so it doesn't fault
	key := XMM()
	VMOVUPS(Mem{Base: keyPtr}, key)

	keyMask := GP64()
	MOVQ(U32(1), keyMask)
	SHLQ(reg.CL, keyMask) // CL = keyLen.As8L()
	DECQ(keyMask)

	mask := GP64()
	XORQ(mask, mask)

	i := GP64()
	XORQ(i, i)

	Label("loop")
	CMPQ(i, count)
	JE(LabelRef("done"))

	CMPL(keyLen.As32(), Mem{Base: lengths, Index: i, Scale: 4})
	JNE(LabelRef("next"))

	nextKey := XMM()
	maskResult := XMM()
	VMOVUPS(Mem{Base: buffer}, nextKey)
	VPCMPEQB(nextKey, key, maskResult)
	VPMOVMSKB(maskResult, mask.As32())
	ANDQ(keyMask, mask)
	CMPQ(keyMask, mask)
	JE(LabelRef("done"))

	Label("next")
	INCQ(i)
	ADDQ(Imm(16), buffer)
	JMP(LabelRef("loop"))

	Label("done")
	Store(i, ReturnIndex(0))
	RET()

	Label("notfound")
	Store(count, ReturnIndex(0))
	RET()

	Generate()
}
