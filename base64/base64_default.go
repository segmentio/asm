// +build !amd64

package base64

import "encoding/base64"

type Encoding = base64.Encoding

func newEncoding(encoder string) *Encoding {
	return base64.NewEncoding(encoder)
}
