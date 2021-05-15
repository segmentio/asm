package sortedset

import (
	"reflect"
	"testing"
)

func TestUnion(t *testing.T) {
	for _, test := range []struct {
		name   string
		a      []byte
		b      []byte
		size   int
		expect []byte
	}{
		{
			name: "empty",
			size: 1,
		},
		{
			name:   "size 1, empty a",
			a:      nil,
			b:      []byte{1, 2, 3, 4, 5},
			size:   1,
			expect: []byte{1, 2, 3, 4, 5},
		},
		{
			name:   "size 1, empty b",
			a:      []byte{1, 2, 3, 4, 5},
			b:      nil,
			size:   1,
			expect: []byte{1, 2, 3, 4, 5},
		},
		{
			name:   "size 1, a == b",
			a:      []byte{1, 2, 3, 4, 5},
			b:      []byte{1, 2, 3, 4, 5},
			size:   1,
			expect: []byte{1, 2, 3, 4, 5},
		},
		{
			name:   "size 1, a < b",
			a:      []byte{1, 2, 3},
			b:      []byte{4, 5, 6},
			size:   1,
			expect: []byte{1, 2, 3, 4, 5, 6},
		},
		{
			name:   "size 1, b < a",
			a:      []byte{4, 5, 6},
			b:      []byte{1, 2, 3},
			size:   1,
			expect: []byte{1, 2, 3, 4, 5, 6},
		},
		{
			name:   "size 1, a <= b",
			a:      []byte{1, 2, 3},
			b:      []byte{3, 4, 5},
			size:   1,
			expect: []byte{1, 2, 3, 4, 5},
		},
		{
			name:   "size 1, b < a",
			a:      []byte{3, 4, 5},
			b:      []byte{1, 2, 3},
			size:   1,
			expect: []byte{1, 2, 3, 4, 5},
		},
		{
			name:   "size 1, interleaved 1",
			a:      []byte{1, 3, 5},
			b:      []byte{2, 4, 6},
			size:   1,
			expect: []byte{1, 2, 3, 4, 5, 6},
		},
		{
			name:   "size 1, interleaved 2",
			a:      []byte{2, 4, 6},
			b:      []byte{1, 3, 5},
			size:   1,
			expect: []byte{1, 2, 3, 4, 5, 6},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			dst := make([]byte, 0, len(test.a)+len(test.b))

			actual := Union(dst, test.a, test.b, test.size)
			if (len(test.expect) == 0 && len(actual) != 0) ||
				(len(test.expect) > 0 && !reflect.DeepEqual(actual, test.expect)) {
				t.Fatalf("not equal: %v vs expected %v", actual, test.expect)
			}
		})
	}
}
