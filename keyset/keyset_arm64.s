//go:build !purego
// +build !purego

#include "textflag.h"

// func Lookup(keyset []byte, key []byte) int
TEXT ·Lookup(SB), NOSPLIT, $0-56
	MOVD    keyset+0(FP), R0
	MOVD    keyset_len+8(FP), R1
	MOVD    key+24(FP), R2
	MOVD    key_len+32(FP), R3

	CMP     $16, R3
	BHI     notfound

	VLD1    (R2), [V0.B16]
	// TODO: pad with zeroes if necessary

	MOVD    R0, R5
	ADD     R0, R1, R4

	MOVD    ZR, R6
	SUB     $1, R6

loop:
	CMP     R0, R4
	BEQ     notfound

	VLD1.P  16(R0), [V1.B16]
	VCMEQ	V0.B16, V1.B16, V2.B16
    VMOV	V2.D[0], R7
    VMOV	V2.D[1], R8
    AND     R7, R8, R9
    CMP     R9, R6
    BEQ     found
    JMP     loop

notfound:
    ADD     R1>>4, ZR, R1
    MOVD    R1, ret+48(FP)
	RET

found:
    SUB     R5, R0, R0
    ADD     R0>>4, ZR, R0
    SUB     $1, R0, R0
    MOVD    R0, ret+48(FP)
    RET
