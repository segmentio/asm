// Code generated by command: go run equal_fold_asm.go -pkg ascii -out ../ascii/equal_fold_amd64.s -stubs ../ascii/equal_fold_amd64.go. DO NOT EDIT.

#include "textflag.h"

// func EqualFoldString(a string, b string) bool
// Requires: AVX, AVX2, SSE4.1
TEXT ·EqualFoldString(SB), NOSPLIT, $0-33
	MOVQ a_base+0(FP), CX
	MOVQ a_len+8(FP), DX
	MOVQ b_base+16(FP), BX
	CMPQ DX, b_len+24(FP)
	JNE  done
	XORQ AX, AX
	CMPQ DX, $0x10
	JB   init_x86
	BTL  $0x08, github·com∕segmentio∕asm∕cpu·X86+0(SB)
	JCS  init_avx

init_x86:
	LEAQ caseMap<>+0(SB), R9
	XORL SI, SI

cmp8:
	CMPQ    DX, $0x08
	JB      cmp7
	MOVBLZX (CX)(AX*1), DI
	MOVBLZX (BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI
	MOVBLZX 1(CX)(AX*1), DI
	MOVBLZX 1(BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI
	MOVBLZX 2(CX)(AX*1), DI
	MOVBLZX 2(BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI
	MOVBLZX 3(CX)(AX*1), DI
	MOVBLZX 3(BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI
	MOVBLZX 4(CX)(AX*1), DI
	MOVBLZX 4(BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI
	MOVBLZX 5(CX)(AX*1), DI
	MOVBLZX 5(BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI
	MOVBLZX 6(CX)(AX*1), DI
	MOVBLZX 6(BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI
	MOVBLZX 7(CX)(AX*1), DI
	MOVBLZX 7(BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI
	JNE     done
	ADDQ    $0x08, AX
	SUBQ    $0x08, DX
	JMP     cmp8

cmp7:
	CMPQ    DX, $0x07
	JB      cmp6
	MOVBLZX 6(CX)(AX*1), DI
	MOVBLZX 6(BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI

cmp6:
	CMPQ    DX, $0x06
	JB      cmp5
	MOVBLZX 5(CX)(AX*1), DI
	MOVBLZX 5(BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI

cmp5:
	CMPQ    DX, $0x05
	JB      cmp4
	MOVBLZX 4(CX)(AX*1), DI
	MOVBLZX 4(BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI

cmp4:
	CMPQ    DX, $0x04
	JB      cmp3
	MOVBLZX 3(CX)(AX*1), DI
	MOVBLZX 3(BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI

cmp3:
	CMPQ    DX, $0x03
	JB      cmp2
	MOVBLZX 2(CX)(AX*1), DI
	MOVBLZX 2(BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI

cmp2:
	CMPQ    DX, $0x02
	JB      cmp1
	MOVBLZX 1(CX)(AX*1), DI
	MOVBLZX 1(BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI

cmp1:
	CMPQ    DX, $0x01
	JB      success
	MOVBLZX (CX)(AX*1), DI
	MOVBLZX (BX)(AX*1), R8
	MOVB    (R9)(DI*1), DI
	XORB    (R9)(R8*1), DI
	ORB     DI, SI

done:
	SETEQ ret+32(FP)
	RET

success:
	MOVB $0x01, ret+32(FP)
	RET

init_avx:
	MOVB         $0x20, SI
	PINSRB       $0x00, SI, X12
	VPBROADCASTB X12, Y12
	MOVB         $0x1f, SI
	PINSRB       $0x00, SI, X13
	VPBROADCASTB X13, Y13
	MOVB         $0x9a, SI
	PINSRB       $0x00, SI, X14
	VPBROADCASTB X14, Y14
	MOVB         $0x01, SI
	PINSRB       $0x00, SI, X15
	VPBROADCASTB X15, Y15

cmp128:
	CMPQ      DX, $0x80
	JB        cmp64
	VMOVDQU   (CX)(AX*1), Y0
	VMOVDQU   32(CX)(AX*1), Y1
	VMOVDQU   64(CX)(AX*1), Y2
	VMOVDQU   96(CX)(AX*1), Y3
	VMOVDQU   (BX)(AX*1), Y4
	VMOVDQU   32(BX)(AX*1), Y5
	VMOVDQU   64(BX)(AX*1), Y6
	VMOVDQU   96(BX)(AX*1), Y7
	VXORPD    Y0, Y4, Y4
	VPCMPEQB  Y12, Y4, Y8
	VORPD     Y12, Y0, Y0
	VPADDB    Y13, Y0, Y0
	VPCMPGTB  Y0, Y14, Y0
	VPAND     Y8, Y0, Y0
	VPAND     Y15, Y0, Y0
	VPSLLW    $0x05, Y0, Y0
	VPCMPEQB  Y4, Y0, Y0
	VXORPD    Y1, Y5, Y5
	VPCMPEQB  Y12, Y5, Y9
	VORPD     Y12, Y1, Y1
	VPADDB    Y13, Y1, Y1
	VPCMPGTB  Y1, Y14, Y1
	VPAND     Y9, Y1, Y1
	VPAND     Y15, Y1, Y1
	VPSLLW    $0x05, Y1, Y1
	VPCMPEQB  Y5, Y1, Y1
	VXORPD    Y2, Y6, Y6
	VPCMPEQB  Y12, Y6, Y10
	VORPD     Y12, Y2, Y2
	VPADDB    Y13, Y2, Y2
	VPCMPGTB  Y2, Y14, Y2
	VPAND     Y10, Y2, Y2
	VPAND     Y15, Y2, Y2
	VPSLLW    $0x05, Y2, Y2
	VPCMPEQB  Y6, Y2, Y2
	VXORPD    Y3, Y7, Y7
	VPCMPEQB  Y12, Y7, Y11
	VORPD     Y12, Y3, Y3
	VPADDB    Y13, Y3, Y3
	VPCMPGTB  Y3, Y14, Y3
	VPAND     Y11, Y3, Y3
	VPAND     Y15, Y3, Y3
	VPSLLW    $0x05, Y3, Y3
	VPCMPEQB  Y7, Y3, Y3
	VPAND     Y1, Y0, Y0
	VPAND     Y3, Y2, Y2
	VPAND     Y2, Y0, Y0
	ADDQ      $0x80, AX
	SUBQ      $0x80, DX
	VPMOVMSKB Y0, SI
	XORL      $0xffffffff, SI
	JNE       done
	JMP       cmp128

cmp64:
	CMPQ      DX, $0x40
	JB        cmp32
	VMOVDQU   (CX)(AX*1), Y0
	VMOVDQU   32(CX)(AX*1), Y1
	VMOVDQU   (BX)(AX*1), Y2
	VMOVDQU   32(BX)(AX*1), Y3
	VXORPD    Y0, Y2, Y2
	VPCMPEQB  Y12, Y2, Y4
	VORPD     Y12, Y0, Y0
	VPADDB    Y13, Y0, Y0
	VPCMPGTB  Y0, Y14, Y0
	VPAND     Y4, Y0, Y0
	VPAND     Y15, Y0, Y0
	VPSLLW    $0x05, Y0, Y0
	VPCMPEQB  Y2, Y0, Y0
	VXORPD    Y1, Y3, Y3
	VPCMPEQB  Y12, Y3, Y5
	VORPD     Y12, Y1, Y1
	VPADDB    Y13, Y1, Y1
	VPCMPGTB  Y1, Y14, Y1
	VPAND     Y5, Y1, Y1
	VPAND     Y15, Y1, Y1
	VPSLLW    $0x05, Y1, Y1
	VPCMPEQB  Y3, Y1, Y1
	VPAND     Y1, Y0, Y0
	ADDQ      $0x40, AX
	SUBQ      $0x40, DX
	VPMOVMSKB Y0, SI
	XORL      $0xffffffff, SI
	JNE       done

cmp32:
	CMPQ      DX, $0x20
	JB        cmp16
	VMOVDQU   (CX)(AX*1), Y0
	VMOVDQU   (BX)(AX*1), Y1
	VXORPD    Y0, Y1, Y1
	VPCMPEQB  Y12, Y1, Y2
	VORPD     Y12, Y0, Y0
	VPADDB    Y13, Y0, Y0
	VPCMPGTB  Y0, Y14, Y0
	VPAND     Y2, Y0, Y0
	VPAND     Y15, Y0, Y0
	VPSLLW    $0x05, Y0, Y0
	VPCMPEQB  Y1, Y0, Y0
	ADDQ      $0x20, AX
	SUBQ      $0x20, DX
	VPMOVMSKB Y0, SI
	XORL      $0xffffffff, SI
	JNE       done

cmp16:
	CMPQ      DX, $0x10
	JLE       cmp_tail
	VMOVDQU   (CX)(AX*1), X0
	VMOVDQU   (BX)(AX*1), X1
	VXORPD    X0, X1, X1
	VPCMPEQB  X12, X1, X2
	VORPD     X12, X0, X0
	VPADDB    X13, X0, X0
	VPCMPGTB  X0, X14, X0
	VPAND     X2, X0, X0
	VPAND     X15, X0, X0
	VPSLLW    $0x05, X0, X0
	VPCMPEQB  X1, X0, X0
	ADDQ      $0x10, AX
	SUBQ      $0x10, DX
	VPMOVMSKB X0, SI
	XORL      $0x0000ffff, SI
	JNE       done

cmp_tail:
	SUBQ      $0x10, DX
	ADDQ      DX, AX
	VMOVDQU   (CX)(AX*1), X0
	VMOVDQU   (BX)(AX*1), X1
	VXORPD    X0, X1, X1
	VPCMPEQB  X12, X1, X2
	VORPD     X12, X0, X0
	VPADDB    X13, X0, X0
	VPCMPGTB  X0, X14, X0
	VPAND     X2, X0, X0
	VPAND     X15, X0, X0
	VPSLLW    $0x05, X0, X0
	VPCMPEQB  X1, X0, X0
	VPMOVMSKB X0, SI
	XORL      $0x0000ffff, SI
	JMP       done

DATA caseMap<>+0(SB)/256, $"\x00\x01\x02\x03\x04\x05\x06\a\b\t\n\v\f\r\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\"#$%&'()*+,-./0123456789:;<=>?@abcdefghijklmnopqrstuvwxyz[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~\u007f\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f\x90\x91\x92\x93\x94\x95\x96\x97\x98\x99\x9a\x9b\x9c\x9d\x9e\x9f\xa0\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf\xb0\xb1\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xbb\xbc\xbd\xbe\xbf\xc0\xc1\xc2\xc3\xc4\xc5\xc6\xc7\xc8\xc9\xca\xcb\xcc\xcd\xce\xcf\xd0\xd1\xd2\xd3\xd4\xd5\xd6\xd7\xd8\xd9\xda\xdb\xdc\xdd\xde\xdf\xe0\xe1\xe2\xe3\xe4\xe5\xe6\xe7\xe8\xe9\xea\xeb\xec\xed\xee\xef\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff"
GLOBL caseMap<>(SB), RODATA|NOPTR, $256