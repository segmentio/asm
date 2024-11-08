// SPDX-License-Identifier: MIT-0

//go:build !purego

#include "textflag.h"

// func validateNEON(p []byte) byte
TEXT ·validateNEON(SB),NOSPLIT,$0-25
    MOVD    s_base+0(FP), R10
    MOVD    s_len+8(FP), R11
    CBZ     R11, valid
    CMP     $16, R11
    BLT     small

    VMOVQ   $0x8080808080808080, $0x8080808080808080, V0

ascii_loop:
    CMP     $16, R11
    BLT     small

    VLD1    (R10), [V1.B16]
    VCMTST  V1.B16, V0.B16, V2.B16
    VMOV    V2.D[0], R2
    VMOV    V2.D[1], R3
    ORR     R2, R3, R2
    CBNZ    R2, stop_ascii

    ADD     $16, R10
    SUB     $16, R11
    B       ascii_loop

stop_ascii:
    VMOVQ   $0x0202020202020202, $0x4915012180808080, V11
    VMOVQ   $0xcbcbcb8b8383a3e7, $0xcbcbdbcbcbcbcbcb, V13
    VMOVQ   $0x0101010101010101, $0x01010101babaaee6, V15
    VMOVQ   $0x0F0F0F0F0F0F0F0F, $0x0F0F0F0F0F0F0F0F, V18
    VMOVQ   $0x0707070707070707, $0x0707070707070707, V12
    VMOVQ   $0xFFFFFFFFFFFFFFFF, $0xFFFFFFFFFFFFFFFF, V14
    VMOVQ   $0x7F7F7F7F7F7F7F7F, $0x7F7F7F7F7F7F7F7F, V16
    VMOVQ   $0xDFDFDFDFDFDFDFDF, $0xDFDFDFDFDFDFDFDF, V17
    VMOVQ   $0x0808080808080808, $0x0808080808080808, V19
    VMOVQ   $0x8080808080808080, $0x8080808080808080, V20
    VMOVQ   $0x0000000000000000, $0x0000000000000000, V30
    VMOVQ   $0x0000000000000000, $0x0000000000000000, V3

aligned_loop:
    VLD1.P  16(R10), [V4.B16]
    VEXT    $15, V4.B16, V3.B16, V5.B16
    VUSHR   $4, V5.B16, V6.B16
    VTBL    V6.B16, [V11.B16], V6.B16
    VAND    V5.B16, V18.B16, V7.B16
    VTBL    V7.B16, [V13.B16], V7.B16
    VUSHR   $4, V4.B16, V8.B16
    VTBL    V8.B16, [V15.B16], V8.B16
    VAND    V6.B16, V7.B16, V9.B16
    VAND    V9.B16, V8.B16, V10.B16
    VEXT    $14, V4.B16, V3.B16, V5.B16
    VUSHR   $5, V5.B16, V6.B16
    VCMEQ   V12.B16, V6.B16, V6.B16
    VEXT    $13, V4.B16, V3.B16, V5.B16
    VUSHR   $4, V5.B16, V9.B16
    VCMEQ   V18.B16, V9.B16, V9.B16
    VORR    V6.B16, V9.B16, V9.B16
    VAND    V9.B16, V20.B16, V9.B16
    VSUB    V9.B16, V10.B16, V9.B16
    VMOV    V9.D[0], R1
    VMOV    V9.D[1], R2
    ORR     R1, R2, R1
    CBNZ    R1, no_valid
    VMOV    V4.B16, V3.B16
    SUB     $16, R11, R11
    CMP     $16, R11

    BGE     aligned_loop

    B small_no_const

small:
    CBZ     R11, valid_ascii

tail_loop:
    MOVBU   (R10), R2        
    AND     $0x80, R2        
    CBNZ    R2, check_utf8        
    ADD     $1, R10          
    SUB     $1, R11          
    CBNZ    R11, tail_loop        
    B       valid_ascii          


check_utf8:

    VMOVQ   $0x0202020202020202, $0x4915012180808080, V11
    VMOVQ   $0xcbcbcb8b8383a3e7, $0xcbcbdbcbcbcbcbcb, V13
    VMOVQ   $0x0101010101010101, $0x01010101babaaee6, V15
    VMOVQ   $0x0F0F0F0F0F0F0F0F, $0x0F0F0F0F0F0F0F0F, V18
    VMOVQ   $0x0707070707070707, $0x0707070707070707, V12
    VMOVQ   $0xFFFFFFFFFFFFFFFF, $0xFFFFFFFFFFFFFFFF, V14
    VMOVQ   $0x7F7F7F7F7F7F7F7F, $0x7F7F7F7F7F7F7F7F, V16
    VMOVQ   $0xDFDFDFDFDFDFDFDF, $0xDFDFDFDFDFDFDFDF, V17
    VMOVQ   $0x0808080808080808, $0x0808080808080808, V19
    VMOVQ   $0x8080808080808080, $0x8080808080808080, V20
    VMOVQ   $0x0000000000000000, $0x0000000000000000, V30
    VMOVQ   $0x0000000000000000, $0x0000000000000000, V3

small_no_const:

    SUB $16, R10, R10
    ADD R11, R10, R10
    VLD1.P  16(R10), [V4.B16]

    ADR  shift_table, R2
    MOVW R11, R3
    LSL $2,  R3
    ADD R3, R2
    B (R2)


shift_table:
    B do_shift_0
    B do_shift_1
    B do_shift_2
    B do_shift_3
    B do_shift_4
    B do_shift_5
    B do_shift_6
    B do_shift_7
    B do_shift_8
    B do_shift_9
    B do_shift_10
    B do_shift_11
    B do_shift_12
    B do_shift_13
    B do_shift_14
    B do_shift_15

do_shift_0:
    VMOVQ   $0x6161616161616161, $0x6161616161616161, V4
    B end_swith
do_shift_1:
    VEXT    $15, V30.B16, V4.B16, V4.B16
    B end_swith
do_shift_2:
    VEXT    $14, V30.B16, V4.B16, V4.B16
    B end_swith
do_shift_3:
    VEXT    $13, V30.B16, V4.B16, V4.B16
    B end_swith
do_shift_4:
    VEXT    $12, V30.B16, V4.B16, V4.B16
    B end_swith
do_shift_5:
    VEXT    $11, V30.B16, V4.B16, V4.B16
    B end_swith
do_shift_6:
    VEXT    $10, V30.B16, V4.B16, V4.B16
    B end_swith
do_shift_7:
    VEXT    $9, V30.B16, V4.B16, V4.B16
    B end_swith
do_shift_8:
    VEXT    $8, V30.B16, V4.B16, V4.B16
    B end_swith
do_shift_9:
    VEXT    $7, V30.B16, V4.B16, V4.B16
    B end_swith
do_shift_10:
    VEXT    $6, V30.B16, V4.B16, V4.B16
    B end_swith
do_shift_11:
    VEXT    $5, V30.B16, V4.B16, V4.B16
    B end_swith
do_shift_12:
    VEXT    $4, V30.B16, V4.B16, V4.B16
    B end_swith
do_shift_13:
    VEXT    $3, V30.B16, V4.B16, V4.B16
    B end_swith
do_shift_14:
    VEXT    $2, V30.B16, V4.B16, V4.B16
    B end_swith
do_shift_15:
    VEXT    $1, V30.B16, V4.B16, V4.B16
    B end_swith

end_swith:
    VEXT    $15, V4.B16, V3.B16, V5.B16 
    VUSHR   $4, V5.B16, V6.B16
    VTBL    V6.B16, [V11.B16], V6.B16
    VAND    V5.B16, V18.B16, V7.B16
    VTBL    V7.B16, [V13.B16], V7.B16
    VUSHR   $4, V4.B16, V8.B16
    VTBL    V8.B16, [V15.B16], V8.B16
    VAND    V6.B16, V7.B16, V9.B16
    VAND    V9.B16, V8.B16, V10.B16

    VEXT    $14, V4.B16, V3.B16, V5.B16
    VUSHR   $5, V5.B16, V6.B16
    VCMEQ   V12.B16, V6.B16, V6.B16

    VEXT    $13, V4.B16, V3.B16, V5.B16
    VUSHR   $4, V5.B16, V9.B16
    VCMEQ   V18.B16, V9.B16, V9.B16
    VORR    V6.B16, V9.B16, V9.B16

    VAND    V9.B16, V20.B16, V9.B16
    VSUB    V9.B16, V10.B16, V9.B16
    VMOV    V9.D[0], R1
    VMOV    V9.D[1], R2
    ORR     R1, R2, R1
    CBNZ    R1, no_valid

valid:
    MOVD    $1, R0
    MOVD    R0, ret+24(FP)
    RET

no_valid:
    MOVD    $0, R0
    MOVD    R0, ret+24(FP)
    RET

valid_ascii:
    MOVD    $3, R0
    MOVD    R0, ret+24(FP)
    RET


