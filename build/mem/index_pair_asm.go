// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	//. "github.com/mmcloughlin/avo/operand"
	//. "github.com/mmcloughlin/avo/reg"
)

func main() {
	TEXT("indexPair1", NOSPLIT, "func(b []byte) int")
	Doc("indexPair1 is the x86 specialization of mem.IndexPair for items of size 1")

	RET()

	Generate()
}
