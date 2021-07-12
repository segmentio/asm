// +build !amd64

package base64

import "encoding/base64"

// An Encoding is a radix 64 encoding/decoding scheme, defined by a
// 64-character alphabet.
type Encoding = base64.Encoding

func newEncoding(encoder string) *Encoding {
	return base64.NewEncoding(encoder)
}
