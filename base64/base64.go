package base64

import (
	"encoding/base64"
)

const (
	StdPadding rune = base64.StdPadding
	NoPadding  rune = base64.NoPadding

	encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	encodeURL = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

	letterRange = int8('Z' - 'A' + 1)
	prefixRange = 2*letterRange + 10
)

// StdEncoding is the standard base64 encoding, as defined in RFC 4648.
var StdEncoding = NewEncoding(encodeStd)

// URLEncoding is the alternate base64 encoding defined in RFC 4648.
// It is typically used in URLs and file names.
var URLEncoding = NewEncoding(encodeURL)

// RawStdEncoding is the standard unpadded base64 encoding defined in RFC 4648 section 3.2.
// This is the same as StdEncoding but omits padding characters.
var RawStdEncoding = StdEncoding.WithPadding(NoPadding)

// RawURLEncoding is the unpadded alternate base64 encoding defined in RFC 4648.
// This is the same as URLEncoding but omits padding characters.
var RawURLEncoding = URLEncoding.WithPadding(NoPadding)

// NewEncoding returns a new padded Encoding defined by the given alphabet,
// which must be a 64-byte string that does not contain the padding character
// or CR / LF ('\r', '\n'). Additionally, the alphabet in ranges [0, 26),
// [26, 52), and [53, 63) must be sequential. This additional letter range
// strictness is required for the accelerated codec. This encoder expects the
// alphabet to follow standard base-64 structure:
//     * 26 sequential characters (standard 'A'..'Z')
//     * 26 sequential characters (standard 'a'..'z')
//     * 10 sequential characters (standard '0'..'9')
//     * 2 characters (standard '+','/' or '-','_')
// While the characters of the sequences can be non-standard, it does not
// support using an abitrary alphabet, unlike the standard library.
// The resulting Encoding uses the default padding character ('='), which may
// be changed or disabled via WithPadding.
func NewEncoding(encoder string) *Encoding {
	var prev byte
	for i := int8(0); i < prefixRange; i++ {
		ch := encoder[i]
		if i%letterRange != 0 && ch != prev+1 {
			panic("encoding alphabet ranges [0,26), [26,52), [53,63) must be sequential")
		}
		prev = ch
	}
	return newEncoding(encoder)
}
