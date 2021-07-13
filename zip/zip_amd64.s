// Code generated by command: go run zip_asm.go -pkg zip -out ../zip/zip_amd64.s -stubs ../zip/zip_amd64.go. DO NOT EDIT.

#include "textflag.h"

// func sumUint64(x []uint64, y []uint64)
// Requires: AVX, AVX2, CMOV
TEXT ·sumUint64(SB), NOSPLIT, $0-48
	XORQ    AX, AX
	MOVQ    x_base+0(FP), CX
	MOVQ    y_base+24(FP), DX
	MOVQ    x_len+8(FP), BX
	MOVQ    y_len+32(FP), SI
	CMPQ    SI, BX
	CMOVQLT SI, BX
	BTL     $0x08, github·com∕segmentio∕asm∕cpu·X86+0(SB)
	JCC     x86_loop

avx2_loop:
	MOVQ    AX, SI
	ADDQ    $0x10, SI
	CMPQ    SI, BX
	JAE     x86_loop
	VMOVDQU (CX)(AX*8), Y0
	VMOVDQU (DX)(AX*8), Y1
	VMOVDQU 32(CX)(AX*8), Y2
	VMOVDQU 32(DX)(AX*8), Y3
	VMOVDQU 64(CX)(AX*8), Y4
	VMOVDQU 64(DX)(AX*8), Y5
	VMOVDQU 96(CX)(AX*8), Y6
	VMOVDQU 96(DX)(AX*8), Y7
	VPADDQ  Y0, Y1, Y0
	VPADDQ  Y2, Y3, Y2
	VPADDQ  Y4, Y5, Y4
	VPADDQ  Y6, Y7, Y6
	VMOVDQU Y0, (CX)(AX*8)
	VMOVDQU Y2, 32(CX)(AX*8)
	VMOVDQU Y4, 64(CX)(AX*8)
	VMOVDQU Y6, 96(CX)(AX*8)
	MOVQ    SI, AX
	JMP     avx2_loop

x86_loop:
	CMPQ AX, BX
	JAE  return
	MOVQ (DX)(AX*8), SI
	ADDQ SI, (CX)(AX*8)
	ADDQ $0x01, AX
	JMP  x86_loop

return:
	RET
