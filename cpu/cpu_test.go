package cpu_test

import (
	"testing"

	"github.com/segmentio/asm/cpu/arm64"
	"github.com/segmentio/asm/cpu/cpuid"
	"github.com/segmentio/asm/cpu/x86"
)

var x86Tests = map[string]cpuid.Feature{
	"SSE":                x86.SSE,
	"SSE2":               x86.SSE2,
	"SSE3":               x86.SSE3,
	"SSE41":              x86.SSE41,
	"SSE42":              x86.SSE42,
	"SSE4A":              x86.SSE4A,
	"SSSE3":              x86.SSSE3,
	"AVX":                x86.AVX,
	"AVX2":               x86.AVX2,
	"AVX512BF16":         x86.AVX512BF16,
	"AVX512BITALG":       x86.AVX512BITALG,
	"AVX512BW":           x86.AVX512BW,
	"AVX512CD":           x86.AVX512CD,
	"AVX512DQ":           x86.AVX512DQ,
	"AVX512ER":           x86.AVX512ER,
	"AVX512F":            x86.AVX512F,
	"AVX512IFMA":         x86.AVX512IFMA,
	"AVX512PF":           x86.AVX512PF,
	"AVX512VBMI":         x86.AVX512VBMI,
	"AVX512VBMI2":        x86.AVX512VBMI2,
	"AVX512VL":           x86.AVX512VL,
	"AVX512VNNI":         x86.AVX512VNNI,
	"AVX512VP2INTERSECT": x86.AVX512VP2INTERSECT,
	"AVX512VPOPCNTDQ":    x86.AVX512VPOPCNTDQ,
}

var arm64Tests = map[string]cpuid.Feature{
	"ASIMD":    arm64.ASIMD,
	"ASIMDDP":  arm64.ASIMDDP,
	"ASIMDHP":  arm64.ASIMDHP,
	"ASIMDRDM": arm64.ASIMDRDM,
}

func TestCPU(t *testing.T) {
	for _, test := range []struct {
		arch string
		feat map[string]cpuid.Feature
	}{
		{arch: "x86", feat: x86Tests},
		{arch: "arm64", feat: arm64Tests},
	} {
		t.Run("none", func(t *testing.T) {
			c := cpuid.CPU(cpuid.None)

			for name, feature := range test.feat {
				t.Run(name, func(t *testing.T) {
					if c.Has(feature) {
						t.Error("cpuid.None must not have any features enabled")
					}
				})
			}
		})

		t.Run("all", func(t *testing.T) {
			c := cpuid.CPU(cpuid.All)

			for name, feature := range test.feat {
				t.Run(name, func(t *testing.T) {
					if !c.Has(feature) {
						t.Errorf("missing a feature that should have been enabled by cpuid.All")
					}
				})
			}
		})

		t.Run("single", func(t *testing.T) {
			for name, feature := range test.feat {
				t.Run(name, func(t *testing.T) {
					c := cpuid.CPU(0)
					c.Set(feature, true)

					for n, f := range test.feat {
						if n == name {
							if !c.Has(f) {
								t.Errorf("expected feature not set on CPU: %s", n)
							}
						} else {
							if c.Has(f) {
								t.Errorf("unexpected feature set on CPU: %s", n)
							}
						}
					}
				})
			}
		})
	}
}
