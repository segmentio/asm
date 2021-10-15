package cpu

import (
	cpuid "github.com/klauspost/cpuid/v2"
)

type X86CPU uint64
type X86Feature uint64

func (c X86CPU) Has(f X86Feature) bool {
	return (uint64(c) & uint64(f)) == uint64(f)
}

func (c *X86CPU) Set(f X86Feature, on bool) {
	*c = X86CPU(set(uint64(*c), uint64(f), on))
}

type ARMCPU uint64
type ARMFeature uint64

func (c ARMCPU) Has(f ARMFeature) bool {
	return (uint64(c) & uint64(f)) == uint64(f)
}

func (c *ARMCPU) Set(f ARMFeature, on bool) {
	*c = ARMCPU(set(uint64(*c), uint64(f), on))
}

var (
	X86 X86CPU
	ARM ARMCPU
)

const (
	SSE                X86Feature = 1 << iota // SSE functions
	SSE2                                      // P4 SSE functions
	SSE3                                      // Prescott SSE3 functions
	SSE4                                      // Penryn SSE4.1 functions
	SSE42                                     // Nehalem SSE4.2 functions
	SSE4A                                     // AMD Barcelona microarchitecture SSE4a instructions
	SSSE3                                     // Conroe SSSE3 functions
	AVX                                       // AVX functions
	AVX2                                      // AVX2 functions
	AVX512BF16                                // AVX-512 BFLOAT16 Instructions
	AVX512BITALG                              // AVX-512 Bit Algorithms
	AVX512BW                                  // AVX-512 Byte and Word Instructions
	AVX512CD                                  // AVX-512 Conflict Detection Instructions
	AVX512DQ                                  // AVX-512 Doubleword and Quadword Instructions
	AVX512ER                                  // AVX-512 Exponential and Reciprocal Instructions
	AVX512F                                   // AVX-512 Foundation
	AVX512FP16                                // AVX-512 FP16 Instructions
	AVX512IFMA                                // AVX-512 Integer Fused Multiply-Add Instructions
	AVX512PF                                  // AVX-512 Prefetch Instructions
	AVX512VBMI                                // AVX-512 Vector Bit Manipulation Instructions
	AVX512VBMI2                               // AVX-512 Vector Bit Manipulation Instructions, Version 2
	AVX512VL                                  // AVX-512 Vector Length Extensions
	AVX512VNNI                                // AVX-512 Vector Neural Network Instructions
	AVX512VP2INTERSECT                        // AVX-512 Intersect for D/Q
	AVX512VPOPCNTDQ                           // AVX-512 Vector Population Count Doubleword and Quadword
	CMOV                                      // Conditional move
)

const (
	ASIMD    ARMFeature = 1 << iota // Advanced SIMD
	ASIMDDP                         // SIMD Dot Product
	ASIMDHP                         // Advanced SIMD half-precision floating point
	ASIMDRDM                        // Rounding Double Multiply Accumulate/Subtract (SQRDMLAH/SQRDMLSH)
)

func init() {
	X86.Set(SSE, cpuid.CPU.Has(cpuid.SSE))
	X86.Set(SSE2, cpuid.CPU.Has(cpuid.SSE2))
	X86.Set(SSE3, cpuid.CPU.Has(cpuid.SSE3))
	X86.Set(SSE4, cpuid.CPU.Has(cpuid.SSE4))
	X86.Set(SSE42, cpuid.CPU.Has(cpuid.SSE42))
	X86.Set(SSE4A, cpuid.CPU.Has(cpuid.SSE4A))
	X86.Set(SSSE3, cpuid.CPU.Has(cpuid.SSSE3))
	X86.Set(AVX, cpuid.CPU.Has(cpuid.AVX))
	X86.Set(AVX2, cpuid.CPU.Has(cpuid.AVX2))
	X86.Set(AVX512BF16, cpuid.CPU.Has(cpuid.AVX512BF16))
	X86.Set(AVX512BITALG, cpuid.CPU.Has(cpuid.AVX512BITALG))
	X86.Set(AVX512BW, cpuid.CPU.Has(cpuid.AVX512BW))
	X86.Set(AVX512CD, cpuid.CPU.Has(cpuid.AVX512CD))
	X86.Set(AVX512DQ, cpuid.CPU.Has(cpuid.AVX512DQ))
	X86.Set(AVX512ER, cpuid.CPU.Has(cpuid.AVX512ER))
	X86.Set(AVX512F, cpuid.CPU.Has(cpuid.AVX512F))
	X86.Set(AVX512FP16, cpuid.CPU.Has(cpuid.AVX512FP16))
	X86.Set(AVX512IFMA, cpuid.CPU.Has(cpuid.AVX512IFMA))
	X86.Set(AVX512PF, cpuid.CPU.Has(cpuid.AVX512PF))
	X86.Set(AVX512VBMI, cpuid.CPU.Has(cpuid.AVX512VBMI))
	X86.Set(AVX512VBMI2, cpuid.CPU.Has(cpuid.AVX512VBMI2))
	X86.Set(AVX512VL, cpuid.CPU.Has(cpuid.AVX512VL))
	X86.Set(AVX512VNNI, cpuid.CPU.Has(cpuid.AVX512VNNI))
	X86.Set(AVX512VP2INTERSECT, cpuid.CPU.Has(cpuid.AVX512VP2INTERSECT))
	X86.Set(AVX512VPOPCNTDQ, cpuid.CPU.Has(cpuid.AVX512VPOPCNTDQ))
	X86.Set(CMOV, cpuid.CPU.Has(cpuid.CMOV))

	ARM.Set(ASIMD, cpuid.CPU.Has(cpuid.ASIMD))
	ARM.Set(ASIMDDP, cpuid.CPU.Has(cpuid.ASIMDDP))
	ARM.Set(ASIMDHP, cpuid.CPU.Has(cpuid.ASIMDHP))
	ARM.Set(ASIMDRDM, cpuid.CPU.Has(cpuid.ASIMDRDM))
}

func set(c, f uint64, on bool) uint64 {
	if on {
		return c | f
	} else {
		return c & ^f
	}
}
