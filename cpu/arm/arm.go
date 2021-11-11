package arm

import (
	"github.com/segmentio/asm/cpu/cpuid"
	. "golang.org/x/sys/cpu"
)

const (
	SWP      cpuid.Feature = 1 << iota // SWP instruction support
	HALF                               // Half-word load and store support
	THUMB                              // ARM Thumb instruction set
	BIT26                              // Address space limited to 26-bits
	FASTMUL                            // 32-bit operand, 64-bit result multiplication support
	FPA                                // Floating point arithmetic support
	VFP                                // Vector floating point support
	EDSP                               // DSP Extensions support
	JAVA                               // Java instruction set
	IWMMXT                             // Intel Wireless MMX technology support
	CRUNCH                             // MaverickCrunch context switching and handling
	THUMBEE                            // Thumb EE instruction set
	NEON                               // NEON instruction set
	VFPv3                              // Vector floating point version 3 support
	VFPv3D16                           // Vector floating point version 3 D8-D15
	TLS                                // Thread local storage support
	VFPv4                              // Vector floating point version 4 support
	IDIVA                              // Integer divide instruction support in ARM mode
	IDIVT                              // Integer divide instruction support in Thumb mode
	VFPD32                             // Vector floating point version 3 D15-D31
	LPAE                               // Large Physical Address Extensions
	EVTSTRM                            // Event stream support
	AES                                // AES hardware implementation
	PMULL                              // Polynomial multiplication instruction set
	SHA1                               // SHA1 hardware implementation
	SHA2                               // SHA2 hardware implementation
	CRC32                              // CRC32 hardware implementation
)

func CPU() cpuid.CPU {
	cpu := cpuid.CPU(0)
	cpu.Set(SWP, ARM.HasSWP)
	cpu.Set(HALF, ARM.HasHALF)
	cpu.Set(THUMB, ARM.HasTHUMB)
	cpu.Set(BIT26, ARM.Has26BIT)
	cpu.Set(FASTMUL, ARM.HasFASTMUL)
	cpu.Set(FPA, ARM.HasFPA)
	cpu.Set(VFP, ARM.HasVFP)
	cpu.Set(EDSP, ARM.HasEDSP)
	cpu.Set(JAVA, ARM.HasJAVA)
	cpu.Set(IWMMXT, ARM.HasIWMMXT)
	cpu.Set(CRUNCH, ARM.HasCRUNCH)
	cpu.Set(THUMBEE, ARM.HasTHUMBEE)
	cpu.Set(NEON, ARM.HasNEON)
	cpu.Set(VFPv3, ARM.HasVFPv3)
	cpu.Set(VFPv3D16, ARM.HasVFPv3D16)
	cpu.Set(TLS, ARM.HasTLS)
	cpu.Set(VFPv4, ARM.HasVFPv4)
	cpu.Set(IDIVA, ARM.HasIDIVA)
	cpu.Set(IDIVT, ARM.HasIDIVT)
	cpu.Set(VFPD32, ARM.HasVFPD32)
	cpu.Set(LPAE, ARM.HasLPAE)
	cpu.Set(EVTSTRM, ARM.HasEVTSTRM)
	cpu.Set(AES, ARM.HasAES)
	cpu.Set(PMULL, ARM.HasPMULL)
	cpu.Set(SHA1, ARM.HasSHA1)
	cpu.Set(SHA2, ARM.HasSHA2)
	cpu.Set(CRC32, ARM.HasCRC32)
	return cpu
}
