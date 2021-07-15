// +build !amd64

package slices

func sumUint64(x []uint64, y []uint64) {
	sumUint64Generic(x, y)
}

func sumUint32(x []uint32, y []uint32) {
	sumUint32Generic(x, y)
}
