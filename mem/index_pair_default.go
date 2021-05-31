// +build !amd64

package mem

func indexPair1(b []byte) int {
	return indexPairGeneric(b, 1)
}
