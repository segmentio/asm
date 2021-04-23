package cpu

import (
	"testing"
)

var x86 = map[string]X86Feature{
	"SSE":                SSE,
	"SSE2":               SSE2,
	"SSE3":               SSE3,
	"SSE4":               SSE4,
	"SSE42":              SSE42,
	"SSE4A":              SSE4A,
	"SSSE3":              SSSE3,
	"AVX":                AVX,
	"AVX2":               AVX2,
	"AVX512BF16":         AVX512BF16,
	"AVX512BITALG":       AVX512BITALG,
	"AVX512BW":           AVX512BW,
	"AVX512CD":           AVX512CD,
	"AVX512DQ":           AVX512DQ,
	"AVX512ER":           AVX512ER,
	"AVX512F":            AVX512F,
	"AVX512IFMA":         AVX512IFMA,
	"AVX512PF":           AVX512PF,
	"AVX512VBMI":         AVX512VBMI,
	"AVX512VBMI2":        AVX512VBMI2,
	"AVX512VL":           AVX512VL,
	"AVX512VNNI":         AVX512VNNI,
	"AVX512VP2INTERSECT": AVX512VP2INTERSECT,
	"AVX512VPOPCNTDQ":    AVX512VPOPCNTDQ,
}

var arm = map[string]ARMFeature{
	"ASIMD":    ASIMD,
	"ASIMDDP":  ASIMDDP,
	"ASIMDHP":  ASIMDHP,
	"ASIMDRDM": ASIMDRDM,
}

func TestX86None(t *testing.T) {
	var c X86CPU

	for name, feature := range x86 {
		if c.Has(feature) {
			t.Errorf("Should not have %s feature enabled", name)
		}
	}
}

func TestX86All(t *testing.T) {
	var c X86CPU
	for _, feature := range x86 {
		c.Set(feature, true)
	}

	for name, feature := range x86 {
		if !c.Has(feature) {
			t.Errorf("Should have %s feature enabled", name)
		}
	}
}

func TestX86Single(t *testing.T) {
	for name, feature := range x86 {
		var c X86CPU
		c.Set(feature, true)
		t.Run(name, func(t *testing.T) {
			for n, f := range x86 {
				if n == name {
					if !c.Has(f) {
						t.Errorf("Should have %s feature enabled", n)
					}
				} else {
					if c.Has(f) {
						t.Errorf("Should not have %s feature enabled", n)
					}
				}
			}
		})
	}
}

func TestARMNone(t *testing.T) {
	var c ARMCPU

	for name, feature := range arm {
		if c.Has(feature) {
			t.Errorf("Should not have %s feature enabled", name)
		}
	}
}

func TestARMAll(t *testing.T) {
	var c ARMCPU
	for _, feature := range arm {
		c.Set(feature, true)
	}

	for name, feature := range arm {
		if !c.Has(feature) {
			t.Errorf("Should have %s feature enabled", name)
		}
	}
}

func TestARMSingle(t *testing.T) {
	for name, feature := range arm {
		var c ARMCPU
		c.Set(feature, true)
		t.Run(name, func(t *testing.T) {
			for n, f := range arm {
				if n == name {
					if !c.Has(f) {
						t.Errorf("Should have %s feature enabled", n)
					}
				} else {
					if c.Has(f) {
						t.Errorf("Should not have %s feature enabled", n)
					}
				}
			}
		})
	}
}
