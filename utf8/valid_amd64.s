// Code generated by command: go run valid_asm.go -pkg utf8 -out ../utf8/valid_amd64.s -stubs ../utf8/valid_amd64.go. DO NOT EDIT.

//go:build !purego
// +build !purego

#include "textflag.h"

// func Valid(p []byte) bool
// Requires: AVX, AVX2, LZCNT
TEXT ·Valid(SB), NOSPLIT, $32-25
	MOVQ p_base+0(FP), AX
	MOVQ p_len+8(FP), CX
	BTL  $0x08, github·com∕segmentio∕asm∕cpu·X86+0(SB)
	JCC  stdlib

	// if input < 128 bytes
	CMPQ CX, $0x80
	JGE  init_avx

stdlib:
	// Non-vectorized implementation from the stdlib. Used for small inputs.
	MOVQ $0x8080808080808080, DX

	// Fast ascii-check loop
start_loop:
	CMPQ  CX, $0x08
	JL    end_loop
	TESTQ DX, AX
	JNZ   end_loop
	SUBQ  $0x08, CX
	ADDQ  $0x08, AX
	JMP   start_loop

end_loop:
	// UTF-8 check byte-by-byte
	LEAQ (AX)(CX*1), CX
	LEAQ first<>+0(SB), DX
	LEAQ accept_ranges<>+0(SB), SI
	JMP  start_utf8_loop_set

start_utf8_loop:
	MOVQ DI, AX

start_utf8_loop_set:
	CMPQ    AX, CX
	JGE     stdlib_ret_true
	MOVBLZX (AX), DI
	CMPB    DI, $0x80
	JAE     test_first
	LEAQ    1(AX), AX
	JMP     start_utf8_loop_set

test_first:
	MOVB    (DX)(DI*1), BL
	CMPB    BL, $0xf1
	JEQ     stdlib_ret_false
	MOVBLZX BL, R8
	ANDL    $0x07, R8
	LEAQ    (AX)(R8*1), DI
	CMPQ    DI, CX
	JA      stdlib_ret_false
	SHRB    $0x04, BL
	MOVBLZX (SI)(BX*2), R9
	MOVBLZX 1(SI)(BX*2), R10
	MOVB    1(AX), BL
	CMPB    BL, R9
	JB      stdlib_ret_false
	CMPB    R10, BL
	JB      stdlib_ret_false
	CMPL    R8, $0x02
	JEQ     start_utf8_loop
	MOVBLZX 2(AX), R9
	SUBL    $0x80, R9
	CMPB    R9, $0x3f
	JHI     stdlib_ret_false
	CMPL    R8, $0x03
	JEQ     start_utf8_loop
	MOVBLZX 3(AX), AX
	SUBL    $0x80, AX
	CMPB    AL, $0x3f
	JLS     start_utf8_loop

stdlib_ret_false:
	MOVB $0x00, ret+24(FP)
	RET

stdlib_ret_true:
	MOVB $0x01, ret+24(FP)
	RET

	// End of stdlib implementation
init_avx:
	LEAQ (SP), DX

	// Prepare the constant masks
	VMOVDQU incomplete_mask<>+0(SB), Y0
	VMOVDQU cont4_vec<>+0(SB), Y1
	VMOVDQU cont3_vec<>+0(SB), Y2

	// High nibble of current byte
	VMOVDQU nibble1_errors<>+0(SB), Y3

	// Low nibble of current byte
	VMOVDQU nibble2_errors<>+0(SB), Y4

	// High nibble of the next byte
	VMOVDQU nibble3_errors<>+0(SB), Y5

	// Nibble mask
	VMOVDQU nibble_mask<>+0(SB), Y6

	// For the first pass, set the previous block as zero.
	VXORPS Y7, Y7, Y7

	// Zeroes the error vector.
	VXORPS Y8, Y8, Y8

	// Zeroes the "previous block was incomplete" vector.
	VXORPS Y9, Y9, Y9
	XORB   BL, BL

	// Top of the loop.
check_input:
	// if bytes left >= 32
	CMPQ CX, $0x20

	// go process the next block
	JGE process

	// If < 32 bytes left
	// Fast exit if done
	CMPQ CX, $0x00
	JE   end

	// If 0 < bytes left < 32.
	CMPB      BL, $0x01
	JNE       stdlib
	VPTEST    Y8, Y8
	JNZ       exit
	VXORPS    Y0, Y0, Y0
	VPCMPEQB  Y9, Y0, Y0
	VPMOVMSKB Y0, DX
	NOTL      DX
	LZCNTL    DX, DX
	SUBQ      $0x20, AX
	ADDQ      DX, AX
	ADDQ      $0x20, CX
	SUBQ      DX, CX
	JMP       stdlib

	// Process one 32B block of data
process:
	// Load the next block of bytes
	VMOVDQU (AX), Y10
	SUBQ    $0x20, CX
	ADDQ    $0x20, AX

	// Fast check to see if ASCII
	VPMOVMSKB Y10, DX
	CMPL      DX, $0x00
	JNZ       non_ascii

	// If this whole block is ASCII, there is nothing to do, and it is an error if any of the previous code point was incomplete.
	VPOR Y8, Y9, Y8
	JMP  check_input

non_ascii:
	// Check errors on the high nibble of the previous byte
	VPERM2I128 $0x03, Y7, Y10, Y9
	VPALIGNR   $0x0f, Y9, Y10, Y9
	VPSRLW     $0x04, Y9, Y11
	VPAND      Y11, Y6, Y11
	VPSHUFB    Y11, Y3, Y11

	// Check errors on the low nibble of the previous byte
	VPAND   Y9, Y6, Y9
	VPSHUFB Y9, Y4, Y9
	VPAND   Y9, Y11, Y11

	// Check errors on the high nibble on the current byte
	VPSRLW  $0x04, Y10, Y9
	VPAND   Y9, Y6, Y9
	VPSHUFB Y9, Y5, Y9
	VPAND   Y9, Y11, Y11

	// Find 3 bytes continuations
	VPERM2I128 $0x03, Y7, Y10, Y9
	VPALIGNR   $0x0e, Y9, Y10, Y9
	VPSUBUSB   Y2, Y9, Y9

	// Find 4 bytes continuations
	VPERM2I128 $0x03, Y7, Y10, Y7
	VPALIGNR   $0x0d, Y7, Y10, Y7
	VPSUBUSB   Y1, Y7, Y7

	// Combine them to have all continuations
	VPOR Y9, Y7, Y7

	// Perform a byte-sized signed comparison with zero to turn any non-zero bytes into 0xFF.
	VXORPS   Y9, Y9, Y9
	VPCMPGTB Y9, Y7, Y7

	// Find bytes that are continuations by looking at their most significant bit.
	VMOVDQU msb_mask<>+0(SB), Y9
	VPAND   Y9, Y7, Y7

	// Find mismatches between expected and actual continuation bytes
	VPXOR Y7, Y11, Y7

	// Store result in sticky error
	VPOR Y8, Y7, Y8

	// Prepare for next iteration
	VPSUBUSB Y0, Y10, Y9
	VMOVDQU  Y10, Y7
	MOVB     $0x01, BL

	// End of loop
	JMP check_input

end:
	// If the previous block was incomplete, this is an error.
	VPOR Y9, Y8, Y8

	// Return whether any error bit was set
	VPTEST Y8, Y8

exit:
	SETEQ ret+24(FP)
	VZEROUPPER
	RET

DATA first<>+0(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+8(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+16(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+24(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+32(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+40(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+48(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+56(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+64(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+72(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+80(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+88(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+96(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+104(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+112(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+120(SB)/8, $0xf0f0f0f0f0f0f0f0
DATA first<>+128(SB)/8, $0xf1f1f1f1f1f1f1f1
DATA first<>+136(SB)/8, $0xf1f1f1f1f1f1f1f1
DATA first<>+144(SB)/8, $0xf1f1f1f1f1f1f1f1
DATA first<>+152(SB)/8, $0xf1f1f1f1f1f1f1f1
DATA first<>+160(SB)/8, $0xf1f1f1f1f1f1f1f1
DATA first<>+168(SB)/8, $0xf1f1f1f1f1f1f1f1
DATA first<>+176(SB)/8, $0xf1f1f1f1f1f1f1f1
DATA first<>+184(SB)/8, $0xf1f1f1f1f1f1f1f1
DATA first<>+192(SB)/8, $0x020202020202f1f1
DATA first<>+200(SB)/8, $0x0202020202020202
DATA first<>+208(SB)/8, $0x0202020202020202
DATA first<>+216(SB)/8, $0x0202020202020202
DATA first<>+224(SB)/8, $0x0303030303030313
DATA first<>+232(SB)/8, $0x0303230303030303
DATA first<>+240(SB)/8, $0xf1f1f14404040434
DATA first<>+248(SB)/8, $0xf1f1f1f1f1f1f1f1
GLOBL first<>(SB), RODATA|NOPTR, $256

DATA accept_ranges<>+0(SB)/8, $0xbf909f80bfa0bf80
DATA accept_ranges<>+8(SB)/8, $0x0000000000008f80
DATA accept_ranges<>+16(SB)/8, $0x0000000000000000
DATA accept_ranges<>+24(SB)/8, $0x0000000000000000
GLOBL accept_ranges<>(SB), RODATA|NOPTR, $32

DATA incomplete_mask<>+0(SB)/8, $0xffffffffffffffff
DATA incomplete_mask<>+8(SB)/8, $0xffffffffffffffff
DATA incomplete_mask<>+16(SB)/8, $0xffffffffffffffff
DATA incomplete_mask<>+24(SB)/8, $0xbfdfefffffffffff
GLOBL incomplete_mask<>(SB), RODATA|NOPTR, $32

DATA cont4_vec<>+0(SB)/8, $0xefefefefefefefef
DATA cont4_vec<>+8(SB)/8, $0xefefefefefefefef
DATA cont4_vec<>+16(SB)/8, $0xefefefefefefefef
DATA cont4_vec<>+24(SB)/8, $0xefefefefefefefef
GLOBL cont4_vec<>(SB), RODATA|NOPTR, $32

DATA cont3_vec<>+0(SB)/8, $0xdfdfdfdfdfdfdfdf
DATA cont3_vec<>+8(SB)/8, $0xdfdfdfdfdfdfdfdf
DATA cont3_vec<>+16(SB)/8, $0xdfdfdfdfdfdfdfdf
DATA cont3_vec<>+24(SB)/8, $0xdfdfdfdfdfdfdfdf
GLOBL cont3_vec<>(SB), RODATA|NOPTR, $32

DATA nibble1_errors<>+0(SB)/8, $0x0202020202020202
DATA nibble1_errors<>+8(SB)/8, $0x4915012180808080
DATA nibble1_errors<>+16(SB)/8, $0x0202020202020202
DATA nibble1_errors<>+24(SB)/8, $0x4915012180808080
GLOBL nibble1_errors<>(SB), RODATA|NOPTR, $32

DATA nibble2_errors<>+0(SB)/8, $0xcbcbcb8b8383a3e7
DATA nibble2_errors<>+8(SB)/8, $0xcbcbdbcbcbcbcbcb
DATA nibble2_errors<>+16(SB)/8, $0xcbcbcb8b8383a3e7
DATA nibble2_errors<>+24(SB)/8, $0xcbcbdbcbcbcbcbcb
GLOBL nibble2_errors<>(SB), RODATA|NOPTR, $32

DATA nibble3_errors<>+0(SB)/8, $0x0101010101010101
DATA nibble3_errors<>+8(SB)/8, $0x01010101babaaee6
DATA nibble3_errors<>+16(SB)/8, $0x0101010101010101
DATA nibble3_errors<>+24(SB)/8, $0x01010101babaaee6
GLOBL nibble3_errors<>(SB), RODATA|NOPTR, $32

DATA nibble_mask<>+0(SB)/8, $0x0f0f0f0f0f0f0f0f
DATA nibble_mask<>+8(SB)/8, $0x0f0f0f0f0f0f0f0f
DATA nibble_mask<>+16(SB)/8, $0x0f0f0f0f0f0f0f0f
DATA nibble_mask<>+24(SB)/8, $0x0f0f0f0f0f0f0f0f
GLOBL nibble_mask<>(SB), RODATA|NOPTR, $32

DATA msb_mask<>+0(SB)/8, $0x8080808080808080
DATA msb_mask<>+8(SB)/8, $0x8080808080808080
DATA msb_mask<>+16(SB)/8, $0x8080808080808080
DATA msb_mask<>+24(SB)/8, $0x8080808080808080
GLOBL msb_mask<>(SB), RODATA|NOPTR, $32
