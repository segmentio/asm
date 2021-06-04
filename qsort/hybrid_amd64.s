// Code generated by command: go run hybrid_asm.go -pkg qsort -out ../qsort/hybrid_amd64.s -stubs ../qsort/hybrid_amd64.go. DO NOT EDIT.

#include "textflag.h"

// func insertionsort32(data *byte, lo int, hi int)
// Requires: AVX, AVX2
TEXT ·insertionsort32(SB), NOSPLIT, $0-24
	MOVQ data+0(FP), AX
	MOVQ lo+8(FP), CX
	MOVQ hi+16(FP), DX
	SHLQ $0x05, CX
	SHLQ $0x05, DX
	LEAQ (AX)(CX*1), CX
	LEAQ (AX)(DX*1), AX
	MOVQ CX, DX

outer:
	ADDQ    $0x20, DX
	CMPQ    DX, AX
	JA      done
	VMOVDQU (DX), Y0
	MOVQ    DX, BX

inner:
	VMOVDQU   -32(BX), Y1
	VPMINUB   Y0, Y1, Y2
	VPCMPEQB  Y0, Y1, Y3
	VPCMPEQB  Y0, Y2, Y2
	VPMOVMSKB Y2, DI
	VPMOVMSKB Y3, SI
	XORL      $0xffffffff, SI
	JZ        outer
	ANDL      SI, DI
	BSFL      SI, SI
	BSFL      DI, DI
	CMPL      SI, DI
	JNE       outer
	VMOVDQU   Y1, (BX)
	VMOVDQU   Y0, -32(BX)
	SUBQ      $0x20, BX
	CMPQ      BX, CX
	JA        inner
	JMP       outer

done:
	VZEROUPPER
	RET

// func distributeForward32(data *byte, scratch *byte, limit int, lo int, hi int, pivot int) int
// Requires: AVX, AVX2, CMOV
TEXT ·distributeForward32(SB), NOSPLIT, $0-56
	MOVQ    data+0(FP), AX
	MOVQ    scratch+8(FP), CX
	MOVQ    limit+16(FP), DX
	MOVQ    lo+24(FP), BX
	MOVQ    hi+32(FP), SI
	MOVQ    pivot+40(FP), DI
	SHLQ    $0x05, DX
	SHLQ    $0x05, BX
	SHLQ    $0x05, SI
	SHLQ    $0x05, DI
	LEAQ    (AX)(BX*1), BX
	LEAQ    (AX)(SI*1), SI
	LEAQ    -32(CX)(DX*1), CX
	VMOVDQU (AX)(DI*1), Y0
	XORQ    DI, DI
	XORQ    R8, R8
	NEGQ    DX

loop:
	VMOVDQU   (BX), Y1
	VPMINUB   Y1, Y0, Y2
	VPCMPEQB  Y1, Y0, Y3
	VPCMPEQB  Y1, Y2, Y2
	VPMOVMSKB Y2, R10
	VPMOVMSKB Y3, R9
	XORL      $0xffffffff, R9
	ANDL      R9, R10
	SETNE     R11
	BSFL      R9, R9
	BSFL      R10, R10
	CMPL      R9, R10
	SETEQ     R8
	ANDB      R11, R8
	XORB      $0x01, R8
	MOVQ      BX, R9
	CMOVQNE   CX, R9
	VMOVDQU   Y1, (R9)(DI*1)
	SHLQ      $0x05, R8
	SUBQ      R8, DI
	ADDQ      $0x20, BX
	CMPQ      BX, SI
	JA        done
	CMPQ      DI, DX
	JNE       loop

done:
	SUBQ AX, BX
	ADDQ DI, BX
	SHRQ $0x05, BX
	DECQ BX
	MOVQ BX, ret+48(FP)
	VZEROUPPER
	RET

// func distributeBackward32(data *byte, scratch *byte, limit int, lo int, hi int, pivot int) int
// Requires: AVX, AVX2, CMOV
TEXT ·distributeBackward32(SB), NOSPLIT, $0-56
	MOVQ    data+0(FP), AX
	MOVQ    scratch+8(FP), CX
	MOVQ    limit+16(FP), DX
	MOVQ    lo+24(FP), BX
	MOVQ    hi+32(FP), SI
	MOVQ    pivot+40(FP), DI
	SHLQ    $0x05, DX
	SHLQ    $0x05, BX
	SHLQ    $0x05, SI
	SHLQ    $0x05, DI
	LEAQ    (AX)(BX*1), BX
	LEAQ    (AX)(SI*1), SI
	VMOVDQU (AX)(DI*1), Y0
	XORQ    DI, DI
	XORQ    R8, R8
	CMPQ    SI, BX
	JBE     done

loop:
	VMOVDQU   (SI), Y1
	VPMINUB   Y1, Y0, Y2
	VPCMPEQB  Y1, Y0, Y3
	VPCMPEQB  Y1, Y2, Y2
	VPMOVMSKB Y2, R10
	VPMOVMSKB Y3, R9
	XORL      $0xffffffff, R9
	ANDL      R9, R10
	SETNE     R11
	BSFL      R9, R9
	BSFL      R10, R10
	CMPL      R9, R10
	SETEQ     R8
	ANDB      R11, R8
	MOVQ      CX, R9
	CMOVQEQ   SI, R9
	VMOVDQU   Y1, (R9)(DI*1)
	SHLQ      $0x05, R8
	ADDQ      R8, DI
	SUBQ      $0x20, SI
	CMPQ      SI, BX
	JBE       done
	CMPQ      DI, DX
	JNE       loop

done:
	SUBQ AX, SI
	ADDQ DI, SI
	SHRQ $0x05, SI
	MOVQ SI, ret+48(FP)
	VZEROUPPER
	RET