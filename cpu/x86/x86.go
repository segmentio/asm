package x86

import (
	"github.com/segmentio/asm/cpu/cpuid"
	. "golang.org/x/sys/cpu"
)

const (
	SSE                cpuid.Feature = 1 << iota // SSE functions
	SSE2                                         // P4 SSE functions
	SSE3                                         // Prescott SSE3 functions
	SSE41                                        // Penryn SSE4.1 functions
	SSE42                                        // Nehalem SSE4.2 functions
	SSE4A                                        // AMD Barcelona microarchitecture SSE4a instructions
	SSSE3                                        // Conroe SSSE3 functions
	AVX                                          // AVX functions
	AVX2                                         // AVX2 functions
	AVX512BF16                                   // AVX-512 BFLOAT16 Instructions
	AVX512BITALG                                 // AVX-512 Bit Algorithms
	AVX512BW                                     // AVX-512 Byte and Word Instructions
	AVX512CD                                     // AVX-512 Conflict Detection Instructions
	AVX512DQ                                     // AVX-512 Doubleword and Quadword Instructions
	AVX512ER                                     // AVX-512 Exponential and Reciprocal Instructions
	AVX512F                                      // AVX-512 Foundation
	AVX512IFMA                                   // AVX-512 Integer Fused Multiply-Add Instructions
	AVX512PF                                     // AVX-512 Prefetch Instructions
	AVX512VBMI                                   // AVX-512 Vector Bit Manipulation Instructions
	AVX512VBMI2                                  // AVX-512 Vector Bit Manipulation Instructions, Version 2
	AVX512VL                                     // AVX-512 Vector Length Extensions
	AVX512VNNI                                   // AVX-512 Vector Neural Network Instructions
	AVX512VP2INTERSECT                           // AVX-512 Intersect for D/Q
	AVX512VPOPCNTDQ                              // AVX-512 Vector Population Count Doubleword and Quadword
	CMOV                                         // Conditional move
)

func CPU() cpuid.CPU {
	cpu := cpuid.CPU(0)
	cpu.Set(SSE, true) // TODO: golang.org/x/sys/cpu assumes all CPUs have SEE?
	cpu.Set(SSE2, X86.HasSSE2)
	cpu.Set(SSE3, X86.HasSSE3)
	cpu.Set(SSE41, X86.HasSSE41)
	cpu.Set(SSE42, X86.HasSSE42)
	cpu.Set(SSE4A, false) // TODO: add upstream support in golang.org/x/sys/cpu?
	cpu.Set(SSSE3, X86.HasSSSE3)
	cpu.Set(AVX, X86.HasAVX)
	cpu.Set(AVX2, X86.HasAVX2)
	cpu.Set(AVX512BF16, X86.HasAVX512BF16)
	cpu.Set(AVX512BITALG, X86.HasAVX512BITALG)
	cpu.Set(AVX512BW, X86.HasAVX512BW)
	cpu.Set(AVX512CD, X86.HasAVX512CD)
	cpu.Set(AVX512DQ, X86.HasAVX512DQ)
	cpu.Set(AVX512ER, X86.HasAVX512ER)
	cpu.Set(AVX512F, X86.HasAVX512F)
	cpu.Set(AVX512IFMA, X86.HasAVX512IFMA)
	cpu.Set(AVX512PF, X86.HasAVX512PF)
	cpu.Set(AVX512VBMI, X86.HasAVX512VBMI)
	cpu.Set(AVX512VBMI2, X86.HasAVX512VBMI2)
	cpu.Set(AVX512VL, X86.HasAVX512VL)
	cpu.Set(AVX512VNNI, X86.HasAVX512VNNI)
	cpu.Set(AVX512VP2INTERSECT, false) // TODO: add upstream support in golang.org/x/sys/cpu?
	cpu.Set(AVX512VPOPCNTDQ, X86.HasAVX512VPOPCNTDQ)
	cpu.Set(CMOV, true) // TODO: golang.org/x/sys/cpu assumes all CPUs have CMOV?
	return cpu
}
