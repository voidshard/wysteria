package test

import (
	"testing"
	wyc "github.com/voidshard/wysteria/common"
)

func marshal(m wyc.Marshalable) ([]byte, error){
	return m.MarshalJSON()
}

func BenchmarkMarshalInterface(b *testing.B) {
	item := &wyc.Item{
		Parent:"oahwdiouawdhaaDuwdiUWUIDHAIHDdawdow",
		ItemType: "foo",
		Variant: "bar",
		Facets: map[string]string{
			"baz": "test1",
			"anotherbaz": "2test",
		},
	}

	for n := 0; n < b.N; n++ {
		marshal(item)
	}
}


func BenchmarkMarshal(b *testing.B) {
	item := &wyc.Item{
		Parent:"oahwdiouawdhaaDuwdiUWUIDHAIHDdawdow",
		ItemType: "foo",
		Variant: "bar",
		Facets: map[string]string{
			"baz": "test1",
			"anotherbaz": "2test",
		},
	}

	for n := 0; n < b.N; n++ {
		item.MarshalJSON()
	}
}
