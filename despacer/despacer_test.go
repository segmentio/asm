package despacer

import "testing"

func TestDespace(t *testing.T) {
	s := `// So here essentially what we're doing is populating pairs
	// of vector registers with 256 bits of integer data, so as an example
	// of vector registers with 256 bits of integer data, so as an example
	// of vector registers with 256 bits of integer data, so as an example
	// of vector registers with 256 bits of integer data, so as an example
	// of vector registers with 256 bits of integer data, so as an example
	// of vector registers with 256 bits of integer data, so as an example
	// of vector registers with 256 bits of integer data, so as an example
	// of vector registers with 256 bits of integer data, so as an example
	// of vector registers with 256 bits of integer data, so as an example
	// of vector registers with 256 bits of integer data, so as an example
	// of vector registers with 256 bits of integer data, so as an example
	// of vector registers with 256 bits of integer data, so as an example
	// of vector registers with 256 bits of integer data, so as an example
	// for uint64s, it would look like...`
	Despace([]byte(s))
}
