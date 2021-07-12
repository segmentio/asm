// +build amd64

package base64

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestEncodeAVX2(t *testing.T) {
	for _, enc := range encodings {
		for i := minEncodeLen; i < 1024; i++ {
			ok := t.Run(fmt.Sprintf("%s-%d", enc.name, i), func(t *testing.T) {
				src := make([]byte, i)
				dst := make([]byte, enc.candidate.EncodedLen(i))

				rand.Read(src)

				_, ns := encodeAVX2(dst, src, enc.candidate.enc)

				if i-ns >= 32 {
					t.Errorf("encodeAVX2 remain should be less than 32, but is %d", i-ns)
				}
			})

			if !ok {
				return
			}
		}
	}
}

func TestDecodeAVX2(t *testing.T) {
	for _, enc := range encodings {
		for i := 34; i < 1024; i++ {
			ok := t.Run(fmt.Sprintf("%s-%d", enc.name, i), func(t *testing.T) {
				raw := make([]byte, i)
				src := make([]byte, enc.candidate.EncodedLen(i))
				dst := make([]byte, i)

				rand.Read(raw)
				enc.candidate.Encode(src, raw)

				_, ns := decodeAVX2(dst, src, enc.candidate.dec)

				if i-ns >= 45 {
					t.Errorf("decodeAVX2 remain should be less than 45, but is %d", i-ns)
				}
			})

			if !ok {
				return
			}
		}
	}
}
