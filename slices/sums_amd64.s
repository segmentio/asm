// Code generated by command: go run sums_asm.go -pkg slices -out ../slices/sums_amd64.s -stubs ../slices/sums_amd64.go. DO NOT EDIT.

// +build !purego

#include "textflag.h"

// func sumUint64(x []uint64, y []uint64)
// Requires: AVX, AVX2, CMOV
TEXT ·sumUint64(SB), NOSPLIT, $0-48
	XORQ    CX, CX
	MOVQ    x_base+0(FP), DX
	MOVQ    y_base+24(FP), BX
	MOVQ    x_len+8(FP), BP
	MOVQ    y_len+32(FP), AX
	CMPQ    AX, BP
	CMOVQLT AX, BP
	BTL     $0x08, github·com∕segmentio∕asm∕cpu·X86+0(SB)
	JCC     x86_loop

avx2_loop:
	MOVQ    CX, AX
	ADDQ    $0x10, AX
	CMPQ    AX, BP
	JAE     x86_loop
	VMOVDQU (DX)(CX*8), Y0
	VMOVDQU (BX)(CX*8), Y1
	VMOVDQU 32(DX)(CX*8), Y2
	VMOVDQU 32(BX)(CX*8), Y3
	VMOVDQU 64(DX)(CX*8), Y4
	VMOVDQU 64(BX)(CX*8), Y5
	VMOVDQU 96(DX)(CX*8), Y6
	VMOVDQU 96(BX)(CX*8), Y7
	VPADDQ  Y0, Y1, Y0
	VPADDQ  Y2, Y3, Y2
	VPADDQ  Y4, Y5, Y4
	VPADDQ  Y6, Y7, Y6
	VMOVDQU Y0, (DX)(CX*8)
	VMOVDQU Y2, 32(DX)(CX*8)
	VMOVDQU Y4, 64(DX)(CX*8)
	VMOVDQU Y6, 96(DX)(CX*8)
	MOVQ    AX, CX
	JMP     avx2_loop

x86_loop:
	CMPQ CX, BP
	JAE  return
	MOVQ (BX)(CX*8), AX
	ADDQ AX, (DX)(CX*8)
	ADDQ $0x01, CX
	JMP  x86_loop

return:
	RET

// func sumUint32(x []uint32, y []uint32)
// Requires: AVX, AVX2, CMOV
TEXT ·sumUint32(SB), NOSPLIT, $0-48
	XORQ    CX, CX
	MOVQ    x_base+0(FP), DX
	MOVQ    y_base+24(FP), BX
	MOVQ    x_len+8(FP), BP
	MOVQ    y_len+32(FP), AX
	CMPQ    AX, BP
	CMOVQLT AX, BP
	BTL     $0x08, github·com∕segmentio∕asm∕cpu·X86+0(SB)
	JCC     x86_loop

avx2_loop:
	MOVQ    CX, AX
	ADDQ    $0x20, AX
	CMPQ    AX, BP
	JAE     x86_loop
	VMOVDQU (DX)(CX*4), Y0
	VMOVDQU (BX)(CX*4), Y1
	VMOVDQU 32(DX)(CX*4), Y2
	VMOVDQU 32(BX)(CX*4), Y3
	VMOVDQU 64(DX)(CX*4), Y4
	VMOVDQU 64(BX)(CX*4), Y5
	VMOVDQU 96(DX)(CX*4), Y6
	VMOVDQU 96(BX)(CX*4), Y7
	VPADDD  Y0, Y1, Y0
	VPADDD  Y2, Y3, Y2
	VPADDD  Y4, Y5, Y4
	VPADDD  Y6, Y7, Y6
	VMOVDQU Y0, (DX)(CX*4)
	VMOVDQU Y2, 32(DX)(CX*4)
	VMOVDQU Y4, 64(DX)(CX*4)
	VMOVDQU Y6, 96(DX)(CX*4)
	MOVQ    AX, CX
	JMP     avx2_loop

x86_loop:
	CMPQ CX, BP
	JAE  return
	MOVL (BX)(CX*4), AX
	ADDL AX, (DX)(CX*4)
	ADDQ $0x01, CX
	JMP  x86_loop

return:
	RET

// func sumUint16(x []uint16, y []uint16)
// Requires: AVX, AVX2, CMOV
TEXT ·sumUint16(SB), NOSPLIT, $0-48
	XORQ    CX, CX
	MOVQ    x_base+0(FP), DX
	MOVQ    y_base+24(FP), BX
	MOVQ    x_len+8(FP), BP
	MOVQ    y_len+32(FP), AX
	CMPQ    AX, BP
	CMOVQLT AX, BP
	BTL     $0x08, github·com∕segmentio∕asm∕cpu·X86+0(SB)
	JCC     x86_loop

avx2_loop:
	MOVQ    CX, AX
	ADDQ    $0x40, AX
	CMPQ    AX, BP
	JAE     x86_loop
	VMOVDQU (DX)(CX*2), Y0
	VMOVDQU (BX)(CX*2), Y1
	VMOVDQU 32(DX)(CX*2), Y2
	VMOVDQU 32(BX)(CX*2), Y3
	VMOVDQU 64(DX)(CX*2), Y4
	VMOVDQU 64(BX)(CX*2), Y5
	VMOVDQU 96(DX)(CX*2), Y6
	VMOVDQU 96(BX)(CX*2), Y7
	VPADDW  Y0, Y1, Y0
	VPADDW  Y2, Y3, Y2
	VPADDW  Y4, Y5, Y4
	VPADDW  Y6, Y7, Y6
	VMOVDQU Y0, (DX)(CX*2)
	VMOVDQU Y2, 32(DX)(CX*2)
	VMOVDQU Y4, 64(DX)(CX*2)
	VMOVDQU Y6, 96(DX)(CX*2)
	MOVQ    AX, CX
	JMP     avx2_loop

x86_loop:
	CMPQ CX, BP
	JAE  return
	MOVW (BX)(CX*2), AX
	ADDW AX, (DX)(CX*2)
	ADDQ $0x01, CX
	JMP  x86_loop

return:
	RET

// func sumUint8(x []uint8, y []uint8)
// Requires: AVX, AVX2, CMOV
TEXT ·sumUint8(SB), NOSPLIT, $0-48
	XORQ    CX, CX
	MOVQ    x_base+0(FP), DX
	MOVQ    y_base+24(FP), BX
	MOVQ    x_len+8(FP), BP
	MOVQ    y_len+32(FP), AX
	CMPQ    AX, BP
	CMOVQLT AX, BP
	BTL     $0x08, github·com∕segmentio∕asm∕cpu·X86+0(SB)
	JCC     x86_loop

avx2_loop:
	MOVQ    CX, AX
	ADDQ    $0x80, AX
	CMPQ    AX, BP
	JAE     x86_loop
	VMOVDQU (DX)(CX*1), Y0
	VMOVDQU (BX)(CX*1), Y1
	VMOVDQU 32(DX)(CX*1), Y2
	VMOVDQU 32(BX)(CX*1), Y3
	VMOVDQU 64(DX)(CX*1), Y4
	VMOVDQU 64(BX)(CX*1), Y5
	VMOVDQU 96(DX)(CX*1), Y6
	VMOVDQU 96(BX)(CX*1), Y7
	VPADDB  Y0, Y1, Y0
	VPADDB  Y2, Y3, Y2
	VPADDB  Y4, Y5, Y4
	VPADDB  Y6, Y7, Y6
	VMOVDQU Y0, (DX)(CX*1)
	VMOVDQU Y2, 32(DX)(CX*1)
	VMOVDQU Y4, 64(DX)(CX*1)
	VMOVDQU Y6, 96(DX)(CX*1)
	MOVQ    AX, CX
	JMP     avx2_loop

x86_loop:
	CMPQ CX, BP
	JAE  return
	MOVB (BX)(CX*1), AL
	ADDB AL, (DX)(CX*1)
	ADDQ $0x01, CX
	JMP  x86_loop

return:
	RET
