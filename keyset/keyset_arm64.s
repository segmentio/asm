//go:build !purego
// +build !purego

#include "textflag.h"

// func Lookup(keyset []byte, key []byte) int
TEXT ·Lookup(SB), NOSPLIT, $0-56
	MOVD keyset+0(FP), R0
	MOVD keyset_len+8(FP), R1
	MOVD key+24(FP), R2
	MOVD key_len+32(FP), R3

	// None of the keys in the set are greater than 16 bytes, so if the input
	// key is we can jump straight to not found.
	CMP $16, R3
	BHI notfound

	// Load the input key and pad with zero bytes. We first load the key into V0
	// and then zero out V1. To blend the two vector registers, we load a mask
	// for the particular key length and then use TBL to select the appropriate
	// bytes from either V0 or V1.
	VLD1 (R2), [V0.B16]
	VMOV ZR, V1.B16
	MOVD $blend_masks<>(SB), R10
	ADD  R3<<4, R10, R10
	VLD1 (R10), [V2.B16]
	VTBL V2.B16, [V0.B16, V1.B16], V3.B16

	// We'll be moving the keyset pointer (R0) forward as we compare keys, so
	// make a copy of the starting point (R5). Add the byte length (R1) to
	// obtain a pointer to the end of the keyset (R4).
	MOVD R0, R5
	ADD  R0, R1, R4

	// Prepare a mask.
	MOVD ZR, R6
	SUB  $1, R6

loop:
	// Loop through each 16 byte key in the keyset.
	CMP R0, R4
	BEQ notfound

	// Load and compare the next key.
	VLD1.P 16(R0), [V4.B16]
	VCMEQ  V3.B16, V4.B16, V5.B16
	VMOV   V5.D[0], R7
	VMOV   V5.D[1], R8
	AND    R7, R8, R9

	// If the masks match, we found the key.
	CMP R9, R6
	BEQ found
	JMP loop

notfound:
	// Return the number of keys in the keyset, which is the byte length (R1)
	// divided by 16.
	ADD  R1>>4, ZR, R1
	MOVD R1, ret+48(FP)
	RET

found:
	// If the key was found, take the position in the keyset and convert it
	// to an index. The keyset pointer (R0) will be 1 key past the match, so
	// subtract the starting pointer (R5), divide by 16 to convert from byte
	// length to an index, and then subtract one.
	SUB  R5, R0, R0
	ADD  R0>>4, ZR, R0
	SUB  $1, R0, R0
	MOVD R0, ret+48(FP)
	RET

DATA blend_masks<>+0(SB)/8, $0x1010101010101010
DATA blend_masks<>+8(SB)/8, $0x1010101010101010
DATA blend_masks<>+16(SB)/8, $0x1010101010101000
DATA blend_masks<>+24(SB)/8, $0x1010101010101010
DATA blend_masks<>+32(SB)/8, $0x1010101010100100
DATA blend_masks<>+40(SB)/8, $0x1010101010101010
DATA blend_masks<>+48(SB)/8, $0x1010101010020100
DATA blend_masks<>+56(SB)/8, $0x1010101010101010
DATA blend_masks<>+64(SB)/8, $0x1010101003020100
DATA blend_masks<>+72(SB)/8, $0x1010101010101010
DATA blend_masks<>+80(SB)/8, $0x1010100403020100
DATA blend_masks<>+88(SB)/8, $0x1010101010101010
DATA blend_masks<>+96(SB)/8, $0x1010050403020100
DATA blend_masks<>+104(SB)/8, $0x1010101010101010
DATA blend_masks<>+112(SB)/8, $0x1006050403020100
DATA blend_masks<>+120(SB)/8, $0x1010101010101010
DATA blend_masks<>+128(SB)/8, $0x0706050403020100
DATA blend_masks<>+136(SB)/8, $0x1010101010101010
DATA blend_masks<>+144(SB)/8, $0x0706050403020100
DATA blend_masks<>+152(SB)/8, $0x1010101010101008
DATA blend_masks<>+160(SB)/8, $0x0706050403020100
DATA blend_masks<>+168(SB)/8, $0x1010101010100908
DATA blend_masks<>+176(SB)/8, $0x0706050403020100
DATA blend_masks<>+184(SB)/8, $0x10101010100A0908
DATA blend_masks<>+192(SB)/8, $0x0706050403020100
DATA blend_masks<>+200(SB)/8, $0x101010100B0A0908
DATA blend_masks<>+208(SB)/8, $0x0706050403020100
DATA blend_masks<>+216(SB)/8, $0x1010100C0B0A0908
DATA blend_masks<>+224(SB)/8, $0x0706050403020100
DATA blend_masks<>+232(SB)/8, $0x10100D0C0B0A0908
DATA blend_masks<>+240(SB)/8, $0x0706050403020100
DATA blend_masks<>+248(SB)/8, $0x100E0D0C0B0A0908
DATA blend_masks<>+256(SB)/8, $0x0706050403020100
DATA blend_masks<>+264(SB)/8, $0x0F0E0D0C0B0A0908
GLOBL blend_masks<>(SB), RODATA|NOPTR, $272
