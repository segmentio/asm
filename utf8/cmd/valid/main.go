package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	stdlib "unicode/utf8"

	"github.com/segmentio/asm/ascii"
	"github.com/segmentio/asm/utf8"
)

func main() {
	var data []byte
	if len(os.Args) > 1 {
		s := os.Args[1]
		s, err := strconv.Unquote(`"` + s + `"`)
		if err != nil {
			panic(err)
		}
		data = []byte(s)
	} else {
		var err error
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
	}

	s := string(data)
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "go test fuzz") {
		fmt.Println("Got fuzzer input")
		// TODO: parse with go/parse instead of regexp?
		r := regexp.MustCompile(`^\[\]byte\((.+)\)`)
		results := r.FindStringSubmatch(lines[1])
		s, err := strconv.Unquote(results[1])
		if err != nil {
			panic(err)
		}
		data = []byte(s)
	}

	fmt.Println(string(data))
	fmt.Println(data)
	fmt.Println(len(data), "bytes")

	uref := stdlib.Valid(data)
	aref := ascii.Valid(data)
	fmt.Println("stdlib: utf8:", uref, "ascii:", aref)

	v := utf8.Validate(data)
	fmt.Println("valid:  utf8:", v.IsUTF8(), "ascii:", v.IsASCII(), "v:", v)

	if uref != v.IsUTF8() || aref != v.IsASCII() {
		os.Exit(1)
	}
}
