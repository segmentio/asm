// Code generated by command: go run keyset_asm.go -pkg keyset -out ../keyset/keyset_amd64.s -stubs ../keyset/keyset_amd64.go. DO NOT EDIT.

// +build !purego

#include "textflag.h"

// func searchAVX(buffer *byte, lengths []uint32, key []byte) int
// Requires: AVX
TEXT ·searchAVX(SB), NOSPLIT, $0-64
	MOVQ key_base+32(FP), SI
	MOVQ key_len+40(FP), CX
	MOVQ key_cap+48(FP), DI
	CMPQ CX, $0x10
	JA   notfound
	MOVQ buffer+0(FP), AX
	MOVQ lengths_base+8(FP), DX
	MOVQ lengths_len+16(FP), BX
	CMPQ DI, $0x10
	JB   check_input

load:
	VMOVUPS (SI), X0

prepare:
	MOVL $0x00000001, SI
	SHLL CL, SI
	DECL SI
	XORQ DI, DI
	MOVQ BX, R12
	SHRQ $0x02, R12
	SHLQ $0x02, R12

bigloop:
	CMPQ      DI, R12
	JE        loop
	CMPL      CX, (DX)(DI*4)
	JNE       try1
	VPCMPEQB  (AX), X0, X8
	VPMOVMSKB X8, R8
	ANDL      SI, R8
	CMPL      SI, R8
	JNE       try1
	JMP       done

try1:
	CMPL      CX, 4(DX)(DI*4)
	JNE       try2
	VPCMPEQB  16(AX), X0, X9
	VPMOVMSKB X9, R9
	ANDL      SI, R9
	CMPL      SI, R9
	JNE       try2
	ADDQ      $0x01, DI
	JMP       done

try2:
	CMPL      CX, 8(DX)(DI*4)
	JNE       try3
	VPCMPEQB  32(AX), X0, X10
	VPMOVMSKB X10, R10
	ANDL      SI, R10
	CMPL      SI, R10
	JNE       try3
	ADDQ      $0x02, DI
	JMP       done

try3:
	CMPL      CX, 12(DX)(DI*4)
	JNE       try4
	VPCMPEQB  48(AX), X0, X11
	VPMOVMSKB X11, R11
	ANDL      SI, R11
	CMPL      SI, R11
	JNE       try4
	ADDQ      $0x03, DI
	JMP       done

try4:
	ADDQ $0x04, DI
	ADDQ $0x40, AX
	JMP  bigloop

loop:
	CMPQ      DI, BX
	JE        done
	CMPL      CX, (DX)(DI*4)
	JNE       next
	VPCMPEQB  (AX), X0, X1
	VPMOVMSKB X1, R8
	ANDL      SI, R8
	CMPL      SI, R8
	JE        done

next:
	INCQ DI
	ADDQ $0x10, AX
	JMP  loop

done:
	MOVQ DI, ret+56(FP)
	RET

notfound:
	MOVQ BX, ret+56(FP)
	RET

check_input:
	MOVQ    SI, DI
	ANDQ    $0x00000fff, DI
	CMPQ    DI, $0x00000ff0
	JBE     load
	MOVQ    $0xfffffffffffffff0, DI
	ADDQ    CX, DI
	VMOVUPS (SI)(DI*1), X0
	LEAQ    shuffle_masks<>+16(SB), SI
	SUBQ    CX, SI
	VMOVUPS (SI), X1
	VPSHUFB X1, X0, X0
	JMP     prepare

DATA shuffle_masks<>+0(SB)/8, $0x0706050403020100
DATA shuffle_masks<>+8(SB)/8, $0x0f0e0d0c0b0a0908
DATA shuffle_masks<>+16(SB)/8, $0x0706050403020100
DATA shuffle_masks<>+24(SB)/8, $0x0f0e0d0c0b0a0908
GLOBL shuffle_masks<>(SB), RODATA|NOPTR, $32
