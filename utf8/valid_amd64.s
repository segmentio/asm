#include "textflag.h"

// Need 2 operations because AVX2 operates on 2x16 blocks, not 1x32.
#define push_last_byte_a_to_front_of_b(A, B, D) VPERM2I128 $0x03, A, B, D; VPALIGNR $15, D, B, D
#define push_last_2bytes_a_to_front_of_b(A, B, D) VPERM2I128 $0x03, A, B, D; VPALIGNR $14, D, B, D
#define push_last_3bytes_a_to_front_of_b(A, B, D) VPERM2I128 $0x03, A, B, D; VPALIGNR $13, D, B, D
#define high_nibbles(A, D) VPSRLW $4, A, D ; VPAND D, Y15, D
#define low_nibbles(A, D) VPAND A, Y15, D

// High nibble of the current byte
DATA nibble1_errors<>+0(SB)/4,  $0x02020202
DATA nibble1_errors<>+4(SB)/4,  $0x02020202
DATA nibble1_errors<>+8(SB)/4,  $0x80808080
DATA nibble1_errors<>+12(SB)/4, $0x49150121
DATA nibble1_errors<>+16(SB)/4, $0x02020202
DATA nibble1_errors<>+20(SB)/4, $0x02020202
DATA nibble1_errors<>+24(SB)/4, $0x80808080
DATA nibble1_errors<>+28(SB)/4, $0x49150121
GLOBL nibble1_errors<>(SB), RODATA, $32

// Low nibble of the current byte
DATA nibble2_errors<>+0(SB)/4,  $0x8383A3E7
DATA nibble2_errors<>+4(SB)/4,  $0xCBCBCB8B
DATA nibble2_errors<>+8(SB)/4,  $0xCBCBCBCB
DATA nibble2_errors<>+12(SB)/4, $0xCBCBDBCB
DATA nibble2_errors<>+16(SB)/4, $0x8383A3E7
DATA nibble2_errors<>+20(SB)/4, $0xCBCBCB8B
DATA nibble2_errors<>+24(SB)/4, $0xCBCBCBCB
DATA nibble2_errors<>+28(SB)/4, $0xCBCBDBCB
GLOBL nibble2_errors<>(SB), RODATA, $32

// High nibble of the next byte
DATA nibble3_errors<>+0(SB)/4,  $0x01010101
DATA nibble3_errors<>+4(SB)/4,  $0x01010101
DATA nibble3_errors<>+8(SB)/4,  $0xBABAAEE6
DATA nibble3_errors<>+12(SB)/4, $0x01010101
DATA nibble3_errors<>+16(SB)/4, $0x01010101
DATA nibble3_errors<>+20(SB)/4, $0x01010101
DATA nibble3_errors<>+24(SB)/4, $0xBABAAEE6
DATA nibble3_errors<>+28(SB)/4, $0x01010101
GLOBL nibble3_errors<>(SB), RODATA, $32

DATA nibble_mask<>+0(SB)/8, $0x0F0F0F0F0F0F0F0F
DATA nibble_mask<>+8(SB)/8, $0x0F0F0F0F0F0F0F0F
DATA nibble_mask<>+16(SB)/8, $0x0F0F0F0F0F0F0F0F
DATA nibble_mask<>+24(SB)/8, $0x0F0F0F0F0F0F0F0F
GLOBL nibble_mask<>(SB), RODATA, $32


// Vector to be substracted to input to check for 3 bytes continuations.
// 0xDF = 0b11100000 - 1
DATA cont3_vec<>+0(SB)/8, $0xDFDFDFDFDFDFDFDF
DATA cont3_vec<>+8(SB)/8, $0xDFDFDFDFDFDFDFDF
DATA cont3_vec<>+16(SB)/8, $0xDFDFDFDFDFDFDFDF
DATA cont3_vec<>+24(SB)/8, $0xDFDFDFDFDFDFDFDF
GLOBL cont3_vec<>(SB), RODATA, $32

// Vector to be substracted to input to check for 4 bytes continuations.
// 0xEF = 0b11110000 - 1
DATA cont4_vec<>+0(SB)/8, $0xEFEFEFEFEFEFEFEF
DATA cont4_vec<>+8(SB)/8, $0xEFEFEFEFEFEFEFEF
DATA cont4_vec<>+16(SB)/8, $0xEFEFEFEFEFEFEFEF
DATA cont4_vec<>+24(SB)/8, $0xEFEFEFEFEFEFEFEF
GLOBL cont4_vec<>(SB), RODATA, $32


// Most significant bit mask
DATA msb_mask<>+0(SB)/8,  $0x8080808080808080
DATA msb_mask<>+8(SB)/8,  $0x8080808080808080
DATA msb_mask<>+16(SB)/8, $0x8080808080808080
DATA msb_mask<>+24(SB)/8, $0x8080808080808080
GLOBL msb_mask<>(SB), RODATA, $32

// Mask to check whether the last 3 bytes may be incomplete.
DATA incomplete_mask<>+0(SB)/8,  $0xFFFFFFFFFFFFFFFF
DATA incomplete_mask<>+8(SB)/8,  $0xFFFFFFFFFFFFFFFF
DATA incomplete_mask<>+16(SB)/8, $0xFFFFFFFFFFFFFFFF
DATA incomplete_mask<>+24(SB)/8, $0xBFDFEFFFFFFFFFFF
GLOBL incomplete_mask<>(SB), RODATA, $32
	

// TODO: check AVX support with CPUID.
	
	// Valid(p []byte) bool
	// Arguments:
	//   p: []byte, 24 bytes
	// Return: bool, 1 byte
	//
	// (SP) -> 32 bytes of scratch space (b)
TEXT ·Valid(SB),NOSPLIT,$72-25

	MOVQ p+0(FP), AX  // p.Data (uintptr, 8 bytes)
	MOVQ p_len+8(FP), DX  // p.Len (int, 8 bytes)
	// Put the address of the scratch space (b) in DI.
        LEAQ (SP), DI

	// AX contains the address of the next load.
	// DX contains the address of the byte past the end of b.

	// Y0 is the current input low chunk (32 bytes).
	// Y1 stores whether the previous input was incomplete.
	// Y2 is the previous input high chunk (32 bytes).
	// Y3 is used as the sticky error register.
	// Y4 stores the incomplete mask.
	// Y5 is a scratch register
	// Y6 is a scratch register
	// Y7 is a scratch register
	// Y8 is a scratch register
	// Y9 stores the continuation 4 vector.
	// Y10 stores the continuation 3 vector.
	// Y11 unused
	// Y12 stores the current high nibble error mask.
	// Y13 stores the current low nibble error mask.
	// Y14 stores the lookahead high nibble error mask.
	// Y15 stores the nibble mask (0x0F in all bytes).

	// Initialize the constant masks registers.
	VMOVDQU incomplete_mask<>(SB), Y4
	VMOVDQU cont4_vec<>(SB), Y9
	VMOVDQU cont3_vec<>(SB), Y10
	VMOVDQU nibble1_errors<>(SB), Y12
	VMOVDQU nibble2_errors<>(SB), Y13
	VMOVDQU nibble3_errors<>(SB), Y14
	VMOVDQU nibble_mask<>(SB), Y15

	// For the first pass, set the previous block as zero.
	VXORPS Y2, Y2, Y2
	// Zeroes the error vector.
	VXORPS Y3, Y3, Y3
	// Zeroes the "previous block was incomplete" vector.
	VXORPS Y1, Y1, Y1

check_input:
	CMPQ DX, $32
	JGE process

	// < 32 bytes left

	// Fast exit if done
	CMPQ DX, $0x00
	JE   end


	// At that point we know we will need to scratch buffer, so zero it.
	VXORPS Y7, Y7, Y7
	VMOVDQU Y7, (DI)
	MOVQ AX, CX
	MOVQ DI, AX

// Copy logic copied from https://github.com/segmentio/asm/blob/main/mem/copy_amd64.s
	CMPQ DX, $0x01
	JE   handle1
	CMPQ DX, $0x03
	JBE  handle2to3
	CMPQ DX, $0x04
	JE   handle4
	CMPQ DX, $0x08
	JB   handle5to7
	JE   handle8
	CMPQ DX, $0x10
	JBE  handle9to16
	CMPQ DX, $0x20
	JBE  handle17to32

	// should panic

handle1:
	MOVB (CX), CL
	MOVB CL, (AX)
	JMP after_copy

handle2to3:
	MOVW (CX), BX
	MOVW -2(CX)(DX*1), CX
	MOVW BX, (AX)
	MOVW CX, -2(AX)(DX*1)
	JMP after_copy

handle4:
	MOVL (CX), CX
	MOVL CX, (AX)
	JMP after_copy

handle5to7:
	MOVL (CX), BX
	MOVL -4(CX)(DX*1), CX
	MOVL BX, (AX)
	MOVL CX, -4(AX)(DX*1)
	JMP after_copy

handle8:
	MOVQ (CX), CX
	MOVQ CX, (AX)
	JMP after_copy

handle9to16:
	MOVQ (CX), BX
	MOVQ -8(CX)(DX*1), CX
	MOVQ BX, (AX)
	MOVQ CX, -8(AX)(DX*1)
	JMP after_copy

handle17to32:
	MOVOU (CX), X0
	MOVOU -16(CX)(DX*1), X1
	MOVOU X0, (AX)
	MOVOU X1, -16(AX)(DX*1)

after_copy:
	MOVQ $32, DX

process:
	// Load the next block of bytes.
	VMOVDQU (AX), Y0
	SUBQ $32, DX
	ADDQ $32, AX

	// Fast check to see if ASCII
	VPMOVMSKB Y0, CX
	CMPQ CX, $0
	JNZ non_ascii

	// If this all block is ASCII, there is nothing to do, and it is an
	// error if any of the previous code point was incomplete.
	VPOR Y3, Y1 , Y3

	JMP check_input



non_ascii:
	// Check low chunk of current input (Y0) against high chunk of previous
	// input (Y2).

	push_last_byte_a_to_front_of_b(Y2, Y0, Y5) // Y5 = prev1

	high_nibbles(Y5, Y8)
	VPSHUFB Y8, Y12, Y8

	low_nibbles(Y5, Y7)
	VPSHUFB Y7, Y13, Y7
	VPAND Y7, Y8, Y8

	high_nibbles(Y0, Y7)
	VPSHUFB Y7, Y14, Y7
	VPAND Y7, Y8, Y8


	push_last_2bytes_a_to_front_of_b(Y2, Y0, Y7)
	push_last_3bytes_a_to_front_of_b(Y2, Y0, Y6)

	VPSUBUSB Y10, Y7, Y7
	VPSUBUSB Y9, Y6, Y6
	VPOR Y7, Y6, Y7 // combine both continuation bits

	// Perform a byte-sized signed comparison with zero to turn any non-zero
	// bytes into 0xFF.
	VPXOR Y6, Y6, Y6
	VPCMPGTB Y6, Y7, Y7
	VMOVDQU msb_mask<>(SB), Y6
	VPAND Y6, Y7, Y7
	VPXOR Y7, Y8, Y7

	// Store result in sticky error
	VPOR Y3, Y7, Y3

	// Check if incomplete
	VPSUBUSB Y4, Y0, Y1 // stores incomplete in Y1
	VMOVDQU Y0, Y2 // Move current block into next block.


	// Move to the next block if there is any left.
	JMP check_input
	
end:
	// If the previous block was incomplete, this is an error.
	VPOR Y1, Y3, Y3

	// Set return value to true if Y2 is zero (no error).
	VPTEST Y3, Y3
	SETEQ ret+24(FP)
	RET
