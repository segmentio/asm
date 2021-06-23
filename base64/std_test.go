package base64

import (
	"bytes"
	libBase64 "encoding/base64"
	"math/rand"
	"testing"
)

func TestStdEncode(t *testing.T) {
	for i := 1; i < 1024; i++ {
		src := make([]byte, i)
		rand.Read(src)

		n := libBase64.StdEncoding.EncodedLen(i)
		expect := make([]byte, n)
		actual := make([]byte, n)

		libBase64.StdEncoding.Encode(expect, src)
		StdEncode(actual, src)

		if !bytes.Equal(expect, actual) {
			t.Errorf("failed with %d bytes:\n\texpect = %v\n\tactual = %v", i, expect, actual)
		}
	}
}

func BenchmarkStdEncode(b *testing.B) {
	src := make([]byte, 4096)
	dst := make([]byte, libBase64.StdEncoding.EncodedLen(len(src)))
	rand.Read(src)

	b.Run("asm", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			StdEncode(dst, src)
		}
		b.SetBytes(int64(len(src)))
	})

	b.Run("go", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			libBase64.StdEncoding.Encode(dst, src)
		}
		b.SetBytes(int64(len(src)))
	})
}
