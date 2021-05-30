// Code generated by command: go run mask_asm.go -pkg mem -out ../mem/mask_amd64.s -stubs ../mem/mask_amd64.go. DO NOT EDIT.

#include "textflag.h"

// func Mask(dst []byte, src []byte) int
// Requires: AVX, AVX2, CMOV, SSE2
TEXT ·Mask(SB), NOSPLIT, $0-56
	MOVQ    dst_base+0(FP), AX
	MOVQ    src_base+24(FP), CX
	MOVQ    dst_len+8(FP), DX
	MOVQ    src_len+32(FP), BX
	CMPQ    BX, DX
	CMOVQGT BX, DX
	MOVQ    DX, ret+48(FP)

	// Tail copy with special cases for each possible item size.
tail:
	CMPQ DX, $0x00
	JE   done
	CMPQ DX, $0x02
	JBE  copy1to2
	CMPQ DX, $0x03
	JE   copy3
	CMPQ DX, $0x04
	JE   copy4
	CMPQ DX, $0x08
	JB   copy5to7
	JE   copy8
	CMPQ DX, $0x10
	JBE  copy9to16
	CMPQ DX, $0x20
	JBE  copy17to32
	CMPQ DX, $0x40
	JBE  copy33to64
	BTL  $0x08, github·com∕segmentio∕asm∕cpu·X86+0(SB)
	JCS  avx2

	// Generic copy for targets without AVX instructions.
generic:
	MOVQ (CX), BX
	ANDQ BX, (AX)
	ADDQ $0x08, CX
	ADDQ $0x08, AX
	SUBQ $0x08, DX
	CMPQ DX, $0x08
	JBE  tail
	JMP  generic

done:
	RET

copy1to2:
	MOVB (CX), BL
	MOVB -1(CX)(DX*1), CL
	ANDB BL, (AX)
	ANDB CL, -1(AX)(DX*1)
	RET

copy3:
	MOVW (CX), DX
	ANDW DX, (AX)
	MOVB 2(CX), CL
	ANDB CL, 2(AX)
	RET

copy4:
	MOVL (CX), CX
	ANDL CX, (AX)
	RET

copy5to7:
	MOVL (CX), BX
	MOVL -4(CX)(DX*1), CX
	ANDL BX, (AX)
	ANDL CX, -4(AX)(DX*1)
	RET

copy8:
	MOVQ (CX), CX
	ANDQ CX, (AX)
	RET

copy9to16:
	MOVQ (CX), BX
	MOVQ -8(CX)(DX*1), CX
	ANDQ BX, (AX)
	ANDQ CX, -8(AX)(DX*1)
	RET

copy17to32:
	MOVOU (CX), X0
	MOVOU -16(CX)(DX*1), X1
	MOVOU (AX), X2
	MOVOU -16(AX)(DX*1), X3
	PAND  X2, X0
	PAND  X3, X1
	MOVOU X0, (AX)
	MOVOU X1, -16(AX)(DX*1)
	RET

copy33to64:
	MOVOU (CX), X0
	MOVOU 16(CX), X1
	MOVOU -32(CX)(DX*1), X2
	MOVOU -16(CX)(DX*1), X3
	MOVOU (AX), X4
	MOVOU 16(AX), X5
	MOVOU -32(AX)(DX*1), X6
	MOVOU -16(AX)(DX*1), X7
	PAND  X4, X0
	PAND  X5, X1
	PAND  X6, X2
	PAND  X7, X3
	MOVOU X0, (AX)
	MOVOU X1, 16(AX)
	MOVOU X2, -32(AX)(DX*1)
	MOVOU X3, -16(AX)(DX*1)
	RET

	// AVX optimized version for medium to large size inputs.
avx2:
	CMPQ    DX, $0x80
	JB      avx2_tail
	VMOVDQU (CX), Y0
	VMOVDQU 32(CX), Y1
	VMOVDQU 64(CX), Y2
	VMOVDQU 96(CX), Y3
	VPAND   (AX), Y0, Y0
	VPAND   32(AX), Y1, Y1
	VPAND   64(AX), Y2, Y2
	VPAND   96(AX), Y3, Y3
	VMOVDQU Y0, (AX)
	VMOVDQU Y1, 32(AX)
	VMOVDQU Y2, 64(AX)
	VMOVDQU Y3, 96(AX)
	ADDQ    $0x80, AX
	ADDQ    $0x80, CX
	SUBQ    $0x80, DX
	JMP     avx2

avx2_tail:
	JZ      done
	CMPQ    DX, $0x40
	JBE     avx2_tail_1to64
	VMOVDQU (CX), Y0
	VMOVDQU 32(CX), Y1
	VMOVDQU 64(CX), Y2
	VMOVDQU -32(CX)(DX*1), Y3
	VPAND   (AX), Y0, Y0
	VPAND   32(AX), Y1, Y1
	VPAND   64(AX), Y2, Y2
	VPAND   -32(AX)(DX*1), Y3, Y3
	VMOVDQU Y0, (AX)
	VMOVDQU Y1, 32(AX)
	VMOVDQU Y2, 64(AX)
	VMOVDQU Y3, -32(AX)(DX*1)
	RET

avx2_tail_1to64:
	VMOVDQU -64(CX)(DX*1), Y0
	VMOVDQU -32(CX)(DX*1), Y1
	VPAND   -64(AX)(DX*1), Y0, Y0
	VPAND   -32(AX)(DX*1), Y1, Y1
	VMOVDQU Y0, -64(AX)(DX*1)
	VMOVDQU Y1, -32(AX)(DX*1)
	RET
