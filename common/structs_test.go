package wysteria_common

import (
	"testing"
)

// ffjson docs warn that using interfaces can cause Go to revert to using default Marshal func
// which can be slower.
//
// Times here seem to be the same so I've left us using the Marshalable interface to make the
// code cleaner.

func marshal(m Marshalable) ([]byte, error) {
	return m.MarshalJSON()
}

func BenchmarkMarshalInterface(b *testing.B) {
	item := &Item{
		Parent:   "oahwdiouawdhaaDuwdiUWUIDHAIHDdawdow",
		ItemType: "foo",
		Variant:  "bar",
		Facets: map[string]string{
			"baz":        "test1",
			"anotherbaz": "2test",
		},
	}

	for n := 0; n < b.N; n++ {
		data, _ := marshal(item)
		item2 := &Item{}
		item2.UnmarshalJSON(data)
	}
}

func BenchmarkMarshal(b *testing.B) {
	item := &Item{
		Parent:   "oahwdiouawdhaaDuwdiUWUIDHAIHDdawdow",
		ItemType: "foo",
		Variant:  "bar",
		Facets: map[string]string{
			"baz":        "test1",
			"anotherbaz": "2test",
		},
	}

	for n := 0; n < b.N; n++ {
		data, _ := item.MarshalJSON()
		item2 := &Item{}
		item2.UnmarshalJSON(data)
	}
}
