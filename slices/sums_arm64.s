//go:build !purego || arm64

#include "textflag.h"

// func sumUint64(x []uint64, y []uint64)
TEXT ·sumUint64(SB), NOSPLIT, $0-48
    MOVD x_base+0(FP), R0
    MOVD x_len+8(FP), R1
    MOVD y_base+24(FP), R2
    MOVD y_len+32(FP), R3

    CMP R1, R3
    CSEL LT, R1, R3, R4
    
    CBZ R4, done
    
    LSR $2, R4, R5
    AND $3, R4, R6

    CBZ R5, remainder
simd_loop:
    VLD1 (R0), [V0.D2, V1.D2]
    VLD1 (R2), [V2.D2, V3.D2]
    
    VADD V2.D2, V0.D2, V0.D2
    VADD V3.D2, V1.D2, V1.D2
    
    VST1.P [V0.D2, V1.D2], 32(R0)
    ADD $32, R2, R2
    
    SUB $1, R5, R5
    CBNZ R5, simd_loop

remainder:
    CBZ R6, done
    
    LSR $1, R6, R7
    CBZ R7, small_remainder
    
    VLD1 (R0), [V0.D2]
    VLD1 (R2), [V1.D2]
    VADD V1.D2, V0.D2, V0.D2
    VST1.P [V0.D2], 16(R0)
    ADD $16, R2, R2
    AND $1, R6, R6

small_remainder:
    CBZ R6, done
    MOVD ZR, R5
rem_loop:
    LSL $3, R5, R7
    MOVD (R0)(R7), R8
    MOVD (R2)(R7), R9
    ADD R8, R9, R8
    MOVD R8, (R0)(R7)
    
    ADD $1, R5, R5
    CMP R6, R5
    BLT rem_loop

done:
    RET

// func sumUint32(x []uint32, y []uint32)
TEXT ·sumUint32(SB), NOSPLIT, $0-48
    MOVD x_base+0(FP), R0
    MOVD x_len+8(FP), R1
    MOVD y_base+24(FP), R2
    MOVD y_len+32(FP), R3

    CMP R1, R3
    CSEL LT, R1, R3, R4
    
    CBZ R4, done
    
    LSR $3, R4, R5
    AND $7, R4, R6

    CBZ R5, remainder
simd_loop:
    VLD1 (R0), [V0.S4, V1.S4]
    VLD1 (R2), [V2.S4, V3.S4]
    
    VADD V2.S4, V0.S4, V0.S4
    VADD V3.S4, V1.S4, V1.S4
    
    VST1.P [V0.S4, V1.S4], 32(R0)
    ADD $32, R2, R2
    
    SUB $1, R5, R5
    CBNZ R5, simd_loop

remainder:
    CBZ R6, done
    
    LSR $2, R6, R7
    CBZ R7, small_remainder
    
    VLD1 (R0), [V0.S4]
    VLD1 (R2), [V1.S4]
    VADD V1.S4, V0.S4, V0.S4
    VST1.P [V0.S4], 16(R0)
    ADD $16, R2, R2
    AND $3, R6, R6

small_remainder:
    CBZ R6, done
    MOVD ZR, R5
rem_loop:
    ADD R5, R5, R5
    ADD R5, R5, R5
    MOVW (R0)(R5), R7
    MOVW (R2)(R5), R8
    ADD R7, R8, R7
    MOVW R7, (R0)(R5)
    LSR $2, R5, R5
    
    ADD $1, R5, R5
    CMP R6, R5
    BLT rem_loop

done:
    RET

// func sumUint16(x []uint16, y []uint16)
TEXT ·sumUint16(SB), NOSPLIT, $0-48
    MOVD x_base+0(FP), R0
    MOVD x_len+8(FP), R1
    MOVD y_base+24(FP), R2
    MOVD y_len+32(FP), R3

    CMP R1, R3
    CSEL LT, R1, R3, R4
    
    CBZ R4, done
    
    LSR $4, R4, R5
    AND $15, R4, R6

    CBZ R5, remainder
simd_loop:
    VLD1 (R0), [V0.H8, V1.H8]
    VLD1 (R2), [V2.H8, V3.H8]
    
    VADD V2.H8, V0.H8, V0.H8
    VADD V3.H8, V1.H8, V1.H8
    
    VST1.P [V0.H8, V1.H8], 32(R0)
    ADD $32, R2, R2
    
    SUB $1, R5, R5
    CBNZ R5, simd_loop

remainder:
    CBZ R6, done
    
    LSR $3, R6, R7
    CBZ R7, small_remainder
    
    VLD1 (R0), [V0.H8]
    VLD1 (R2), [V1.H8]
    VADD V1.H8, V0.H8, V0.H8
    VST1.P [V0.H8], 16(R0)
    ADD $16, R2, R2
    AND $7, R6, R6

small_remainder:
    CBZ R6, done
    MOVD ZR, R5
rem_loop:
    ADD R5, R5, R5
    MOVW (R0)(R5), R7
    MOVW (R2)(R5), R8
    ADD R7, R8, R7
    MOVW R7, (R0)(R5)
    LSR $1, R5, R5
    
    ADD $1, R5, R5
    CMP R6, R5
    BLT rem_loop

done:
    RET

// func sumUint8(x []uint8, y []uint8)
TEXT ·sumUint8(SB), NOSPLIT, $0-48
    MOVD x_base+0(FP), 	R0
    MOVD x_len+8(FP), 	R1
    MOVD y_base+24(FP), R2
    MOVD y_len+32(FP), 	R3

    CMP R1, R3
    CSEL LT, R1, R3, R4
    
    CBZ R4, done
    
    LSR $5,  R4, R5
    AND $31, R4, R6

    CBZ R5, remainder
simd_loop:
    VLD1 (R0), [V0.B16, V1.B16]
    VLD1 (R2), [V2.B16, V3.B16]
    
    VADD V2.B16, V0.B16, V0.B16
    VADD V3.B16, V1.B16, V1.B16
    
    VST1.P [V0.B16, V1.B16], 32(R0)
    ADD $32, R2, R2
    
    SUB $1, R5, R5
    CBNZ R5, simd_loop

remainder:
    CBZ R6, done
    
    LSR $4, R6, R7
    CBZ R7, small_remainder
    
    VLD1 (R0), [V0.B16]
    VLD1 (R2), [V1.B16]
    VADD V1.B16, V0.B16, V0.B16
    VST1.P [V0.B16], 16(R0)
    ADD $16, R2, R2
    AND $15, R6, R6

small_remainder:
    CBZ R6, done
    MOVD ZR, R5
rem_loop:
    MOVB (R0)(R5), R7
    MOVB (R2)(R5), R8
    ADD R7, R8, R7
    MOVB R7, (R0)(R5)
    
    ADD $1, R5, R5
    CMP R6, R5
    BLT rem_loop
done:
    RET
