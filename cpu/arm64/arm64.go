package arm64

import (
	"github.com/segmentio/asm/cpu/cpuid"
	. "golang.org/x/sys/cpu"
)

const (
	FP       cpuid.Feature = 1 << iota // Floating-point instruction set (always available)
	ASIMD                              // Advanced SIMD (always available)
	EVTSTRM                            // Event stream support
	AES                                // AES hardware implementation
	PMULL                              // Polynomial multiplication instruction set
	SHA1                               // SHA1 hardware implementation
	SHA2                               // SHA2 hardware implementation
	CRC32                              // CRC32 hardware implementation
	ATOMICS                            // Atomic memory operation instruction set
	FPHP                               // Half precision floating-point instruction set
	ASIMDHP                            // Advanced SIMD half precision instruction set
	CPUID                              // CPUID identification scheme registers
	ASIMDRDM                           // Rounding double multiply add/subtract instruction set
	JSCVT                              // Javascript conversion from floating-point to integer
	FCMA                               // Floating-point multiplication and addition of complex numbers
	LRCPC                              // Release Consistent processor consistent support
	DCPOP                              // Persistent memory support
	SHA3                               // SHA3 hardware implementation
	SM3                                // SM3 hardware implementation
	SM4                                // SM4 hardware implementation
	ASIMDDP                            // Advanced SIMD double precision instruction set
	SHA512                             // SHA512 hardware implementation
	SVE                                // Scalable Vector Extensions
	ASIMDFHM                           // Advanced SIMD multiplication FP16 to FP32
)

func CPU() cpuid.CPU {
	cpu := cpuid.CPU(0)
	cpu.Set(FP, ARM64.HasFP)
	cpu.Set(ASIMD, ARM64.HasASIMD)
	cpu.Set(EVTSTRM, ARM64.HasEVTSTRM)
	cpu.Set(AES, ARM64.HasAES)
	cpu.Set(PMULL, ARM64.HasPMULL)
	cpu.Set(SHA1, ARM64.HasSHA1)
	cpu.Set(SHA2, ARM64.HasSHA2)
	cpu.Set(CRC32, ARM64.HasCRC32)
	cpu.Set(ATOMICS, ARM64.HasATOMICS)
	cpu.Set(FPHP, ARM64.HasFPHP)
	cpu.Set(ASIMDHP, ARM64.HasASIMDHP)
	cpu.Set(CPUID, ARM64.HasCPUID)
	cpu.Set(ASIMDRDM, ARM64.HasASIMDRDM)
	cpu.Set(JSCVT, ARM64.HasJSCVT)
	cpu.Set(FCMA, ARM64.HasFCMA)
	cpu.Set(LRCPC, ARM64.HasLRCPC)
	cpu.Set(DCPOP, ARM64.HasDCPOP)
	cpu.Set(SHA3, ARM64.HasSHA3)
	cpu.Set(SM3, ARM64.HasSM3)
	cpu.Set(SM4, ARM64.HasSM4)
	cpu.Set(ASIMDDP, ARM64.HasASIMDDP)
	cpu.Set(SHA512, ARM64.HasSHA512)
	cpu.Set(SVE, ARM64.HasSVE)
	cpu.Set(ASIMDFHM, ARM64.HasASIMDFHM)
	return cpu
}
