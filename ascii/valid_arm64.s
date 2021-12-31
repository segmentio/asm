// +build !purego

#include "textflag.h"

// func ValidString(s string) bool
TEXT ·ValidString(SB), NOSPLIT, $0-17
	MOVD s_base+0(FP), R0 // string base
	MOVD s_len+8(FP), R1 // length of string
	MOVD $ret+16(FP), R2 // address for result

	CMP $0x10, R1 // if len(s) < 16 goto cmp8
	BLT cmp8

	// With the current implementation, we haven't found
	// better performance while testing against more than 2 vectors
	// at a time.
	MOVD $0x8080808080808080, R8 // mask for vectors
	VMOV	R8, V5.B16
	B cmp32

cmp8:
	CMP $0x08, R1 // if len(s) < 8 goto cmp4
	BLT cmp4

	MOVD (R0), R3
	AND $0x8080808080808080, R3, R3
	CBNZ R3, invalid

	SUB $0x08, R1, R1
	CBZ R1, done

	ADD $0x08, R0, R0
	JMP cmp8

cmp4:
	CMP $0x04, R1 // if len(s) < 4 goto cmp3
	BLT cmp3

	MOVW (R0), R3
	AND $0x80808080, R3, R3
	CBNZ R3, invalid

	SUB $0x04, R1, R1
	CBZ R1, done

	ADD $0x04, R0, R0
	JMP cmp4


cmp3:
	CMP $0x03, R1 // if len(s) < 3 goto cmp2
	BLT cmp2

	MOVHU (R0), R3
	MOVBU 2(R0), R5
	ORR R5<<16, R3, R6

	AND $0x80808080, R6, R0
	CBZ R0, done

	JMP invalid

cmp2:
	TBZ $0x01, R1, cmp1 // if len(s) == 1 goto cmp1

	MOVHU (R0), R0	
	AND $0x8080, R0, R0
	CMPW ZR, R0
	CSET EQ, R0
	MOVB R0, (R2)
	RET

cmp1:
	TBZ $0x00, R1, done // if len(s) == 0 we are done

	MOVBU (R0), R0	
	AND $0x80, R0, R0
	CMPW ZR, R0
	CSET EQ, R0
	MOVB R0, (R2)
	RET

done:
	ORR $1, ZR, R0
	MOVB R0, (R2)
	RET

invalid:
	MOVD ZR, (R2)
	RET

// Following is using ARM64 ASIMD instructions.
// A better implementation could be written using SVE.
cmp32:
	CMP $0x20, R1 // if len(s) < 32 goto cmp16
	BLT cmp16

	// Load 32B into two vectors, test against mask
	// exit if invalid.
	VLD1  (R0), [V0.D2, V1.D2] 
	VAND V5.B16, V0.B16, V0.B16
	VAND V5.B16, V1.B16, V1.B16
	VORR V1.B16, V0.B16, V0.B16
	VMOV V0.D[0], R3
	VMOV V0.D[1], R4
	ORR R3, R4, R4
	CBNZ R4, invalid

	// Move the pointers to the next 32B.
	ADD $0x20, R0, R0
	SUB $0x20, R1, R1

cmp16:
	CMP $0x10, R1 // if len(s) < 16 goto cmp8
	BLT cmp8 

	// Load 16B into two vectors, test against mask
	// exit if invalid.
	VLD1  (R0), [V0.D1, V1.D1] 
	VAND V5.B16, V0.B16, V0.B16
	VAND V5.B16, V1.B16, V1.B16
	VORR V1.B16, V0.B16, V0.B16
	VMOV V0.D[0], R3
	CBNZ R3, invalid

	// Move the pointers to the next 16B.
	ADD $0x10, R0, R0
	SUB $0x10, R1, R1

cmp_tail:
	SUB $0x10, R1, R1
	ADD R1, R0, R0

	VLD1  (R0), [V0.D1, V1.D1] 
	VAND V5.B16, V0.B16, V0.B16
	VAND V5.B16, V1.B16, V1.B16
	VORR V1.B16, V0.B16, V0.B16
	VMOV V0.D[0], R3

	CBNZ R3, invalid

	B done
