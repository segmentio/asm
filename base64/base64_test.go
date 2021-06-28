package base64

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/rand"
	"testing"
)

const encodeNonStd = "abcdefghijklmnopqrstuvwxyz" +
	"\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c" +
	"\x8d\x8e\x8f\x90\x91\x92\x93\x94\x95\x96\x97\x98\x99" +
	"#$%&'()*+,-."

var encodings = map[string]struct {
	control   *base64.Encoding
	candidate *Encoding
}{
	"std": {
		control:   base64.StdEncoding,
		candidate: StdEncoding,
	},
	"url": {
		control:   base64.URLEncoding,
		candidate: URLEncoding,
	},
	"raw-std": {
		control:   base64.RawStdEncoding,
		candidate: RawStdEncoding,
	},
	"raw-url": {
		control:   base64.RawURLEncoding,
		candidate: RawURLEncoding,
	},
	"non-std": {
		control:   base64.NewEncoding(encodeNonStd),
		candidate: NewEncoding(encodeNonStd),
	},
}

func TestEncoding(t *testing.T) {
	for name, enc := range encodings {
		for i := 1; i < 1024; i++ {
			ok := t.Run(fmt.Sprintf("%s-%d", name, i), func(t *testing.T) {
				src := make([]byte, i)
				rand.Read(src)

				n := enc.control.EncodedLen(i)
				encExpect := make([]byte, n)
				encActual := make([]byte, n)

				enc.control.Encode(encExpect, src)
				enc.candidate.Encode(encActual, src)

				if !bytes.Equal(encExpect, encActual) {
					t.Errorf("failed encode:\n\texpect = %v\n\tactual = %v", encExpect, encActual)
				}

				n = enc.control.DecodedLen(n)
				decExpect := make([]byte, n)
				decActual := make([]byte, n)

				nControl, errControl := enc.control.Decode(decExpect, encExpect)
				nCandidate, errCandidate := enc.candidate.Decode(decActual, encActual)

				if errControl != nil {
					t.Fatalf("control decode failed: %v", errControl)
				}

				if errCandidate != nil {
					t.Fatalf("candidate decode failed: %v", errCandidate)
				}

				if nControl != nCandidate {
					t.Fatalf("failed decode length: expect = %d, actual = %d", nControl, nCandidate)
				}

				if !bytes.Equal(decExpect, decActual) {
					t.Errorf("failed decode:\n\texpect = %v\n\tactual = %v", decExpect, decActual)
				}
			})

			if !ok {
				break
			}
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	src := make([]byte, 4096)
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	rand.Read(src)

	b.Run("asm", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			StdEncoding.Encode(dst, src)
		}
		b.SetBytes(int64(len(src)))
	})

	b.Run("go", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			base64.StdEncoding.Encode(dst, src)
		}
		b.SetBytes(int64(len(src)))
	})
}
