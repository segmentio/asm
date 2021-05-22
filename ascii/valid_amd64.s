// Code generated by command: go run valid_asm.go -pkg ascii -out ../ascii/valid_amd64.s -stubs ../ascii/valid_amd64.go. DO NOT EDIT.

#include "textflag.h"

// func ValidString(s string) bool
// Requires: AVX, AVX2, SSE4.1
TEXT ·ValidString(SB), NOSPLIT, $0-17
	MOVQ         s_base+0(FP), AX
	MOVQ         s_len+8(FP), CX
	MOVQ         $0x8080808080808080, DX
	BTL          $0x08, github·com∕segmentio∕asm∕cpu·X86+0(SB)
	JCC          cmp8
	PINSRQ       $0x00, DX, X4
	VPBROADCASTQ X4, Y4
	CMPQ         CX, $0x80
	JNB          cmp256

cmp64:
	CMPQ    CX, $0x40
	JB      cmp32
	VMOVDQU (AX), Y0
	VPOR    32(AX), Y0, Y0
	VPTEST  Y0, Y4
	JNZ     invalid
	ADDQ    $0x40, AX
	SUBQ    $0x40, CX

cmp32:
	CMPQ   CX, $0x20
	JB     cmp16
	VPTEST (AX), Y4
	JNZ    invalid
	ADDQ   $0x20, AX
	SUBQ   $0x20, CX

cmp16:
	CMPQ   CX, $0x10
	JB     cmp8
	VPTEST (AX), X4
	JNZ    invalid
	ADDQ   $0x10, AX
	SUBQ   $0x10, CX

cmp8:
	CMPQ  CX, $0x08
	JB    cmp4
	TESTQ DX, (AX)
	JNZ   invalid
	ADDQ  $0x08, AX
	SUBQ  $0x08, CX
	JMP   cmp8

cmp4:
	CMPQ  CX, $0x04
	JB    cmp3
	TESTL $0x80808080, (AX)
	JNZ   invalid
	ADDQ  $0x04, AX
	SUBQ  $0x04, CX

cmp3:
	CMPQ    CX, $0x03
	JB      cmp2
	MOVWLZX (AX), CX
	MOVBLZX 2(AX), AX
	SHLL    $0x10, AX
	ORL     CX, AX
	TESTL   $0x80808080, AX
	JMP     done

cmp2:
	CMPQ  CX, $0x02
	JB    cmp1
	TESTW $0x8080, (AX)
	JMP   done

cmp1:
	CMPQ  CX, $0x00
	JE    done
	TESTB $0x80, (AX)

done:
	SETEQ ret+16(FP)
	RET

invalid:
	MOVB $0x00, ret+16(FP)
	RET

cmp256:
	CMPQ    CX, $0x00000100
	JB      cmp128
	VMOVDQU (AX), Y0
	VPOR    32(AX), Y0, Y0
	VMOVDQU 64(AX), Y1
	VPOR    96(AX), Y1, Y1
	VMOVDQU 128(AX), Y2
	VPOR    160(AX), Y2, Y2
	VMOVDQU 192(AX), Y3
	VPOR    224(AX), Y3, Y3
	VPOR    Y1, Y0, Y0
	VPOR    Y3, Y2, Y2
	VPOR    Y2, Y0, Y0
	VPTEST  Y0, Y4
	JNZ     invalid
	ADDQ    $0x00000100, AX
	SUBQ    $0x00000100, CX
	JMP     cmp256

cmp128:
	CMPQ    CX, $0x80
	JB      cmp64
	VMOVDQU (AX), Y0
	VPOR    32(AX), Y0, Y0
	VMOVDQU 64(AX), Y1
	VPOR    96(AX), Y1, Y1
	VPOR    Y1, Y0, Y0
	VPTEST  Y0, Y4
	JNZ     invalid
	ADDQ    $0x80, AX
	SUBQ    $0x80, CX
	JMP     cmp64
