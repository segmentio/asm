// Package cpuid provides generic types used to represent CPU features supported
// by the architecture.
package cpuid

// CPU is a bitset of feature flags representing the capabilities of various CPU
// architeectures that this package provides optimized assembly routines for.
//
// The intent is to provide a stable ABI between the Go code that generate the
// assembly, and the program that uses the library functions.
type CPU uint64

// Feature represents a single CPU feature.
type Feature uint64

const (
	// None is a Feature value that has no CPU features enabled.
	None Feature = 0
	// All is a Feature value that has all CPU features enabled.
	All Feature = 0xFFFFFFFFFFFFFFFF
)

func (c CPU) Has(f Feature) bool {
	return (Feature(c) & f) == f
}

func (c *CPU) Set(f Feature, on bool) {
	if on {
		*c |= CPU(f)
	} else {
		*c &= ^CPU(f)
	}
}
