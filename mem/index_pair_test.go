package mem

import (
	"fmt"
	"testing"
)

var (
	indexPairSizes = [...]int{
		1, 2, 4, 8, 10, 16, 32,
	}
)

func TestIndexPair(t *testing.T) {
	for _, size := range indexPairSizes {
		makeInput := func(values ...byte) []byte {
			input := make([]byte, size*len(values))
			for i := range values {
				input[i*size] = values[i]
			}
			return input
		}

		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			tests := []struct {
				scenario string
				input    []byte
				index    int
			}{
				{
					scenario: "empty input",
					input:    nil,
					index:    0,
				},

				{
					scenario: "input with only one item",
					input:    makeInput(1),
					index:    1,
				},

				{
					scenario: "input with two non-equal items",
					input:    makeInput(1, 2),
					index:    2,
				},

				{
					scenario: "input with two equal items",
					input:    makeInput(1, 1),
					index:    0,
				},

				{
					scenario: "input with two equal items at the end",
					input:    makeInput(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 9),
					index:    9,
				},
			}

			for _, test := range tests {
				t.Run(test.scenario, func(t *testing.T) {
					i := test.index * size
					j := IndexPair(test.input, size)

					if i != j {
						t.Errorf("expected=%d found=%d", i, j)
					}
				})
			}
		})
	}
}

func BenchmarkIndexPair(b *testing.B) {
	for _, size := range indexPairSizes {
		input := make([]byte, 256*size)
		for i := 0; i < 256; i++ {
			input[i*size] = byte(i)
		}

		b.Run(fmt.Sprintf("N=%d", size), func(b *testing.B) {
			b.SetBytes(int64(len(input)))

			for i := 0; i < b.N; i++ {
				_ = IndexPair(input, size)
			}
		})
	}
}
