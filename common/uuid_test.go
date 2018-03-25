package wysteria_common

import (
	"testing"
)

func TestNewUUID(t *testing.T) {
	// arrange
	cases := []struct{
		Args []interface{}
	} {
		{[]interface{}{1, 2, 3, 4, 5}},
		{[]interface{}{0, 2, 3, 4, 5}},
		{[]interface{}{1, "kljdiad", 3, 4, 5}},
		{[]interface{}{1, 2, 3, "iadiawduiawiduhwad", 5}},
		{[]interface{}{1, 2, 3, 4, 5}},
		{[]interface{}{1, 2, true, 4, 5}},
		{[]interface{}{1, 2, "bananana", 4, 5}},
		{[]interface{}{1, 2, 3, 4, "foo"}},
		{[]interface{}{1, 2, 3, 4, "Foo"}},
		{[]interface{}{1, false, 3, true, "foo"}},
		{[]interface{}{1, false, 3, true, "foo"}},
		{[]interface{}{1, "blah", 3, "bar", "foo"}},
		{[]interface{}{"collection", "foo", "dcaa5941-7670-4fc4-b338-a085881664a1"}},
		{[]interface{}{"collection", "bar", nil}},
		{[]interface{}{"collection", "foo", "dcacc5941-7670-4fc4-b338-a085881664a1"}},
		{[]interface{}{"collection", "foo", "dcaaahww-7670-4fc4-b338-a085881664a1"}},
		{[]interface{}{"collection", "foo", "dcaa5941-7dad0-4fc4-b338-a085881664a1"}},
		{[]interface{}{"resource", "foo", "bar", "dcaa5941-7dad0-4fc4-b338-a085881664a1"}},
		{[]interface{}{"resource", "foo", "z", "dcaa5941-7dad0-4fc4-b338-a085881664a1"}},
		{[]interface{}{"resource", "foo", "thumbnail", "/akjdjw/amdwm/aw"}},
		{[]interface{}{"resource", "foo", "png", "/bar/bar/foo/moo"}},
		{[]interface{}{"resource", "foo", "jpeg", "dcaa5941-7dad0-4fc4-b338-a085881664a1"}},
	}

	soFar := map[string]bool{}

	for i, tst := range cases {
		// act
		a := newId(tst.Args...)
		b := newId(tst.Args...)

		_, ok := soFar[a]

		// assert
		if ok {
			t.Error(i, "Repeated id formed from", tst.Args)
		}

		if a != b {
			t.Error(i, "Expected", a, b, "to be equal given", tst.Args)
		}
	}
}