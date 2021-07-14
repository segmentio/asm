// +build !amd64

package slices

func sumUint64(x []uint64, y []uint64) {
	sumUint64Generic(x, y)
}
