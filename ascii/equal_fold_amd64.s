// Code generated by command: go run equal_fold_asm.go -pkg ascii -out ../ascii/equal_fold_amd64.s -stubs ../ascii/equal_fold_amd64.go. DO NOT EDIT.

#include "textflag.h"

// func EqualFoldString(a string, b string) bool
// Requires: AVX, AVX2, SSE4.1
TEXT ·EqualFoldString(SB), NOSPLIT, $0-33
	MOVQ         a_base+0(FP), AX
	MOVQ         a_len+8(FP), CX
	MOVQ         b_base+16(FP), DX
	CMPQ         CX, b_len+24(FP)
	JNE          done
	XORQ         DI, DI
	MOVQ         $0xdfdfdfdfdfdfdfdf, SI
	BTL          $0x08, github·com∕segmentio∕asm∕cpu·X86+0(SB)
	JCC          eq8
	PINSRQ       $0x00, SI, X1
	VPBROADCASTQ X1, Y1

eq64:
	CMPQ      CX, $0x40
	JB        eq32
	VPAND     (AX)(DI*1), Y1, Y0
	VPAND     (DX)(DI*1), Y1, Y2
	VPCMPEQB  Y2, Y0, Y0
	VPAND     32(AX)(DI*1), Y1, Y2
	VPAND     32(DX)(DI*1), Y1, Y3
	VPCMPEQB  Y3, Y2, Y2
	VPAND     Y2, Y0, Y0
	VPMOVMSKB Y0, BX
	ADDQ      $0x40, DI
	SUBQ      $0x40, CX
	CMPL      BX, $0xffffffff
	JNE       done
	JMP       eq64

eq32:
	CMPQ      CX, $0x20
	JB        eq16
	VPAND     (AX)(DI*1), Y1, Y0
	VPAND     (DX)(DI*1), Y1, Y2
	VPCMPEQB  Y2, Y0, Y0
	VPMOVMSKB Y0, BX
	ADDQ      $0x20, DI
	SUBQ      $0x20, CX
	CMPL      BX, $0xffffffff
	JNE       done

eq16:
	CMPQ      CX, $0x10
	JB        eq8
	VPAND     (AX)(DI*1), X1, X0
	VPAND     (DX)(DI*1), X1, X1
	VPCMPEQB  X1, X0, X0
	VPMOVMSKB X0, BX
	ADDQ      $0x10, DI
	SUBQ      $0x10, CX
	CMPL      BX, $0x0000ffff
	JNE       done

eq8:
	CMPQ  CX, $0x08
	JB    eq4
	MOVQ  (AX)(DI*1), BX
	XORQ  (DX)(DI*1), BX
	ADDQ  $0x08, DI
	SUBQ  $0x08, CX
	TESTQ SI, BX
	JNE   done
	JMP   eq8

eq4:
	CMPQ  CX, $0x04
	JB    eq3
	MOVL  (AX)(DI*1), BX
	XORL  (DX)(DI*1), BX
	ADDQ  $0x04, DI
	SUBQ  $0x04, CX
	TESTL $0xdfdfdfdf, BX
	JNE   done

eq3:
	CMPQ    CX, $0x03
	JB      eq2
	MOVWLZX (AX)(DI*1), BX
	MOVBLZX 2(AX)(DI*1), AX
	SHLL    $0x10, AX
	ORL     BX, AX
	MOVWLZX (DX)(DI*1), BX
	MOVBLZX 2(DX)(DI*1), CX
	SHLL    $0x10, CX
	ORL     BX, CX
	XORL    AX, CX
	TESTL   $0x00dfdfdf, CX
	JMP     done

eq2:
	CMPQ  CX, $0x02
	JB    eq1
	MOVW  (AX)(DI*1), BX
	XORW  (DX)(DI*1), BX
	TESTW $0xdfdf, BX
	JMP   done

eq1:
	CMPQ  CX, $0x00
	JE    done
	MOVB  (AX)(DI*1), BL
	XORB  (DX)(DI*1), BL
	TESTB $0xdf, BL

done:
	SETEQ ret+32(FP)
	RET
