// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("Despace", NOSPLIT, "func(data []byte)")
	Doc("remove spaces (in-place) from string bytes (UTF-8 or ASCII)")

	//spaces := YMM()
	//newline := YMM()
	VPBROADCASTB(Imm(' '), YMM())
	//__m256i spaces = _mm256_set1_epi8(' ');
	//__m256i newline = _mm256_set1_epi8('\n');
	//__m256i carriage = _mm256_set1_epi8('\r');
	Generate()

}
