package ascii

import "github.com/segmentio/asm/cpu"

func ValidString(s string) bool {
	return validString(s, uint64(cpu.X86))
}

func ValidPrintString(s string) bool {
	return validPrintString(s, uint64(cpu.X86))
}

func EqualFoldString(a, b string) bool {
	return equalFoldString(a, b, uint64(cpu.X86))
}
