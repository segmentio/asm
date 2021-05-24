// Code generated by command: go run valid_print_asm.go -pkg ascii -out ../ascii/valid_print_amd64.s -stubs ../ascii/valid_print_amd64.go. DO NOT EDIT.

#include "textflag.h"

// func ValidPrintString(s string) bool
// Requires: AVX, AVX2, SSE4.1
TEXT ·ValidPrintString(SB), NOSPLIT, $0-17
	MOVQ s_base+0(FP), AX
	MOVQ s_len+8(FP), CX
	CMPQ CX, $0x10
	JB   init
	BTL  $0x08, github·com∕segmentio∕asm∕cpu·X86+0(SB)
	JCS  avx

init:
	CMPQ CX, $0x08
	JB   cmp4
	MOVQ $0xdfdfdfdfdfdfdfe0, DX
	MOVQ $0x0101010101010101, BX
	MOVQ $0x8080808080808080, SI

cmp8:
	MOVQ  (AX), DI
	MOVQ  DI, R8
	LEAQ  (DI)(DX*1), R9
	NOTQ  R8
	ANDQ  R8, R9
	LEAQ  (DI)(BX*1), R8
	ORQ   R8, DI
	ORQ   R9, DI
	ADDQ  $0x08, AX
	SUBQ  $0x08, CX
	TESTQ SI, DI
	JNE   done
	CMPQ  CX, $0x08
	JB    cmp4
	JMP   cmp8

cmp4:
	CMPQ  CX, $0x04
	JB    cmp3
	MOVL  (AX), DX
	MOVL  DX, BX
	LEAL  3755991008(DX), SI
	NOTL  BX
	ANDL  BX, SI
	LEAL  16843009(DX), BX
	ORL   BX, DX
	ORL   SI, DX
	ADDQ  $0x04, AX
	SUBQ  $0x04, CX
	TESTL $0x80808080, DX
	JNE   done

cmp3:
	CMPQ    CX, $0x03
	JB      cmp2
	MOVWLZX (AX), CX
	MOVBLZX 2(AX), AX
	SHLL    $0x10, AX
	ORL     CX, AX
	ORL     $0x20000000, AX
	JMP     final

cmp2:
	CMPQ    CX, $0x02
	JB      cmp1
	MOVWLZX (AX), AX
	ORL     $0x20200000, AX
	JMP     final

cmp1:
	CMPQ    CX, $0x00
	JE      done
	MOVBLZX (AX), AX
	ORL     $0x20202000, AX

final:
	MOVL  AX, CX
	LEAL  3755991008(AX), DX
	NOTL  CX
	ANDL  CX, DX
	LEAL  16843009(AX), CX
	ORL   CX, AX
	ORL   DX, AX
	TESTL $0x80808080, AX

done:
	SETEQ ret+16(FP)
	RET

avx:
	MOVQ         $0x1f1f1f1f1f1f1f1f, DX
	PINSRQ       $0x00, DX, X8
	VPBROADCASTQ X8, Y8
	MOVQ         $0x7e7e7e7e7e7e7e7e, DX
	PINSRQ       $0x00, DX, X9
	VPBROADCASTQ X9, Y9

cmp128:
	CMPQ      CX, $0x80
	JB        cmp64
	VMOVDQU   (AX), Y0
	VPCMPGTB  Y8, Y0, Y1
	VPCMPGTB  Y9, Y0, Y0
	VPANDN    Y1, Y0, Y0
	VMOVDQU   32(AX), Y2
	VPCMPGTB  Y8, Y2, Y3
	VPCMPGTB  Y9, Y2, Y2
	VPANDN    Y3, Y2, Y2
	VMOVDQU   64(AX), Y4
	VPCMPGTB  Y8, Y4, Y5
	VPCMPGTB  Y9, Y4, Y4
	VPANDN    Y5, Y4, Y4
	VMOVDQU   96(AX), Y6
	VPCMPGTB  Y8, Y6, Y7
	VPCMPGTB  Y9, Y6, Y6
	VPANDN    Y7, Y6, Y6
	VPAND     Y2, Y0, Y0
	VPAND     Y6, Y4, Y4
	VPAND     Y4, Y0, Y0
	ADDQ      $0x80, AX
	SUBQ      $0x80, CX
	VPMOVMSKB Y0, DX
	XORL      $0xffffffff, DX
	JNE       done
	JMP       cmp128

cmp64:
	CMPQ      CX, $0x40
	JB        cmp32
	VMOVDQU   (AX), Y0
	VPCMPGTB  Y8, Y0, Y1
	VPCMPGTB  Y9, Y0, Y0
	VPANDN    Y1, Y0, Y0
	VMOVDQU   32(AX), Y2
	VPCMPGTB  Y8, Y2, Y3
	VPCMPGTB  Y9, Y2, Y2
	VPANDN    Y3, Y2, Y2
	VPAND     Y2, Y0, Y0
	ADDQ      $0x40, AX
	SUBQ      $0x40, CX
	VPMOVMSKB Y0, DX
	XORL      $0xffffffff, DX
	JNE       done

cmp32:
	CMPQ      CX, $0x20
	JB        cmp16
	VMOVDQU   (AX), Y0
	VPCMPGTB  Y8, Y0, Y1
	VPCMPGTB  Y9, Y0, Y0
	VPANDN    Y1, Y0, Y0
	ADDQ      $0x20, AX
	SUBQ      $0x20, CX
	VPMOVMSKB Y0, DX
	XORL      $0xffffffff, DX
	JNE       done

cmp16:
	CMPQ      CX, $0x10
	JB        init
	VMOVDQU   (AX), X0
	VPCMPGTB  X8, X0, X1
	VPCMPGTB  X9, X0, X0
	VPANDN    X1, X0, X0
	ADDQ      $0x10, AX
	SUBQ      $0x10, CX
	VPMOVMSKB X0, DX
	XORL      $0x0000ffff, DX
	JNE       done
	CMPQ      CX, $0x00
	JE        done
	JMP       init
