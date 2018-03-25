package integration

import (
	"testing"
	wyc "github.com/voidshard/wysteria/client"
)

func TestItemVariantSearch(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	setup := []struct{
		Type string
		Variant string
	}{
		{"bob", "me"},
		{"bob", "orme"},
		{"james", "orme"},
		{"alice", "notme"},
	}

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}

	for i, s := range setup {
		_, err := collection.CreateItem(s.Type, s.Variant)
		if err != nil {
			t.Skip(i, err)
		}
	}

	cases := []struct{
		Expect int
		Search *wyc.Search
	} {
		{2, client.Search(wyc.ItemType("bob"))},
		{1, client.Search(wyc.ItemType("bob"), wyc.ItemVariant("me"))},
		{2, client.Search(wyc.ItemVariant("orme"))},
		{3, client.Search(wyc.ItemType("bob")).Or(wyc.ItemVariant("orme"))},
	}

	// act
	for i, tst := range cases {
		found, err := tst.Search.FindItems()
		if err != nil {
			t.Error(i, err)
		}

		// assert
		if len(found) != tst.Expect {
			t.Error(i, "Expected", tst.Expect, "found", len(found))
		}
	}
}


func TestFacetSearch(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	setup := []struct{
		Name string
		Facets map[string]string
	}{
		{randomString(), map[string]string{"find": "me"}},
		{randomString(), map[string]string{"find": "me", "should": "not_matter"}},
		{randomString(), map[string]string{"find": "notme"}},
	}

	for i, tst := range setup {
		_, err := client.CreateCollection(tst.Name, wyc.Facets(tst.Facets))
		if err != nil {
			t.Error(i, err)
		}
	}

	cases := []struct{
		Expect int
		Facets map[string]string
	} {
		{2, map[string]string{"find": "me"}},
		{0, map[string]string{"find": "me", "and": "me"}},
		{0, map[string]string{"nothing": "here"}},
	}

	// act
	for i, tst := range cases {
		found, err := client.Search(wyc.HasFacets(tst.Facets)).FindCollections()
		if err != nil {
			t.Error(i, err)
		}

		// assert
		if len(found) != tst.Expect {
			t.Error(i, "Expected", tst.Expect, "found", len(found))
		}
	}
}