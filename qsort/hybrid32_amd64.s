#include "go_asm.h"
#include "textflag.h"

// Byte-wise comparison of two ymm registers: out = (YL >= YR)
// Registers R14,R15,Y14,Y15 are clobbered.
#define LESS32(YL, YR, out) \
    VPMINUB     YL, YR, Y14; \
    VPCMPEQB    YL, YR, Y15; \
    VPCMPEQB    YL, Y14, Y14; \
    VPMOVMSKB   Y15, R14; \
    VPMOVMSKB   Y14, out; \
    XORL        $0xFFFFFFFF, R14; \
    ANDL        R14, out; \
    SETNE       R15; \
    BSFL        R14, R14; \
    BSFL        out, out; \
    CMPL        R14, out; \
    SETEQ       out;      \
    ANDL        R15, out; \
    XORL        $1, out

TEXT ·insertionsort32(SB),NOSPLIT,$0-40
    MOVQ        data_base+0(FP), AX
    MOVQ        lo+24(FP), BX
    MOVQ        hi+32(FP), CX
    SHLQ        $5, BX
    SHLQ        $5, CX
    LEAQ        (AX)(BX*1), BX
    LEAQ        (AX)(CX*1), CX
    LEAQ        32(BX), DX
outer:
    CMPQ        DX, CX
    JA          done
    VMOVDQU     (DX), Y0
    MOVQ        DX, SI
inner:
    CMPQ        SI, BX
    JBE         next
    VMOVDQU     -32(SI), Y1
    VPMINUB     Y0, Y1, Y2
    VPCMPEQB    Y0, Y1, Y3
    VPCMPEQB    Y0, Y2, Y2
    VPMOVMSKB   Y2, R8
    VPMOVMSKB   Y3, R9
    XORL        $0xFFFFFFFF, R9
    JZ          next
    ANDL        R9, R8
    BSFL        R9, R9
    BSFL        R8, R8
    CMPL        R9, R8
    JNE         next
    VMOVDQU     Y1, (SI)
    VMOVDQU     Y0, -32(SI)
    SUBQ        $32, SI
    JMP         inner
next:
    ADDQ        $32, DX
    JMP         outer
done:
    RET

TEXT ·medianOfThree32(SB),NOSPLIT,$0-48
    MOVQ        data_base+0(FP), DX
    MOVQ        a+24(FP), AX
    MOVQ        b+32(FP), BX
    MOVQ        c+40(FP), CX
    SHLQ        $5, AX
    SHLQ        $5, BX
    SHLQ        $5, CX
    ADDQ        DX, AX
    ADDQ        DX, BX
    ADDQ        DX, CX
    VMOVDQU     (AX), Y0
    VMOVDQU     (BX), Y1
    VMOVDQU     (CX), Y2
    LESS32      (Y1, Y0, R8) // B < A?
    JNE         part2
    VMOVDQU     Y1, (AX)
    VMOVDQU     Y0, (BX)
    VMOVDQA     Y1, Y3
    VMOVDQA     Y0, Y1
    VMOVDQA     Y3, Y0
part2:
    LESS32      (Y2, Y1, R8) // C < B?
    JNE         done
    VMOVDQU     Y2, (BX)
    VMOVDQU     Y1, (CX)
    LESS32      (Y2, Y0, R8) // B < A?
    JNE         done
    VMOVDQU     Y2, (AX)
    VMOVDQU     Y0, (BX)
done:
    VZEROUPPER
    RET

TEXT ·distributeForward32(SB),NOSPLIT,$0-72
    MOVQ        data_base+0(FP), AX
    MOVQ        tmp_base+24(FP), R12
    MOVQ        tmp_len+32(FP), R8
    MOVQ        lo+48(FP), SI
    MOVQ        hi+56(FP), DI
    MOVQ        pivot+64(FP), R11
    SHLQ        $5, R8
    SHLQ        $5, SI
    SHLQ        $5, DI
    SHLQ        $5, R11
    VMOVDQU     (AX)(R11*1), Y1     // Y1: pivot = data[pivot]
    LEAQ        (AX)(DI*1), CX      // CX: end = &data[hi]
    LEAQ        -32(R12)(R8*1), BX   // BX: tmp = &tmp[(len(tmp)-1)*32]
    LEAQ        (AX)(SI*1), AX      // AX: data = &data[lo]
    XORQ        DX, DX              // DX: offset = 0
    XORQ        R8, R8              // R8: isLarger = 0
loop:
    CMPQ        AX, CX              // data <= end ?
    JA          done
    VMOVDQU     (AX), Y0            // Y0: item = *data
    LESS32      (Y0, Y1, R8)        // R9: isLarger = (Y0 >= Y1)
    MOVQ        AX, R9              // R8: dest = isLarger ? tmp : data
    CMOVQNE     BX, R9
    VMOVDQU     Y0, (R9)(DX*1)      // dest[offset] = *data
    SHLQ        $5, R8
    SUBQ        R8, DX              // offset -= isLarger * 32
    ADDQ        $32, AX             // data += 32
    LEAQ        32(BX)(DX*1), R13
    CMPQ        R13, R12
    JBE         done
    JMP         loop
done:
    MOVQ        data_base+0(FP), BX
    SUBQ        BX, AX
    ADDQ        DX, AX
    SHRQ        $5, AX
    DECQ        AX
    MOVQ        AX, ret+72(FP)
    VZEROUPPER
    RET

TEXT ·distributeBackward32(SB),NOSPLIT,$0-80
    MOVQ        data_base+0(FP), AX
    MOVQ        tmp_base+24(FP), BX
    MOVQ        tmp_len+32(FP), R10
    MOVQ        lo+48(FP), SI
    MOVQ        hi+56(FP), DI
    MOVQ        pivot+64(FP), R11
    SHLQ        $5, R10
    SHLQ        $5, R11
    SHLQ        $5, SI
    SHLQ        $5, DI
    VMOVDQU     (AX)(R11*1), Y1
    LEAQ        (AX)(SI*1), CX
    LEAQ        (AX)(DI*1), AX
    XORQ        DX, DX
    XORQ        R8, R8
loop:
    CMPQ        AX, CX
    JBE         done
    VMOVDQU     (AX), Y0
    LESS32      (Y0, Y1, R8)
    MOVQ        BX, R9
    CMOVQNE     AX, R9
    VMOVDQU     Y0, (R9)(DX*1)
    XORQ        $1, R8
    SHLQ        $5, R8
    ADDQ        R8, DX
    SUBQ        $32, AX
    CMPQ        DX, R10
    JE          done
    JMP         loop
done:
    MOVQ        data_base+0(FP), BX
    SUBQ        BX, AX
    ADDQ        DX, AX
    SHRQ        $5, AX
    MOVQ        AX, ret+72(FP)
    VZEROUPPER
    RET
