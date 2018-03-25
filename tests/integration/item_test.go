package integration


import (
	"testing"
	wyc "github.com/voidshard/wysteria/client"
)


func TestItemDeletion(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err) // can't perform test
	}

	fromCentreLinkName := "link"
	toCentreLinkName := "centre"

	cases := []struct{
		Type string
		Variant string
		ExtraLinkName string
		LinkFacets map[string]string
	} {
		{"house", "brick", "foo", map[string]string{"a": "b"}},
		{"tree", "oak", "bar", map[string]string{"madeBy": "batman"}},
		{"person", "male", "moo", map[string]string{"a": "b", "foo": "moo"}},
	}

	centre, err := collection.CreateItem("super", "item") // we'll link all to this
	if err != nil {
		t.Skip(err)
	}

	var previous *wyc.Item

	for i, tst := range cases {
		item, err := collection.CreateItem(tst.Type, tst.Variant)
		if err != nil {
			t.Skip(i, err)
		}

		_, err = centre.LinkTo(fromCentreLinkName, item, wyc.Facets(tst.LinkFacets))
		if err != nil {
			t.Skip(i, err)
		}

		_, err = item.LinkTo(toCentreLinkName, centre)
		if err != nil {
			t.Skip(i, err)
		}

		if previous != nil { // random extra link
			_, err = item.LinkTo(tst.ExtraLinkName, previous)
			if err != nil {
				t.Skip(i, err)
			}
		}

		// act
		err = item.Delete()
		if err != nil {
			t.Error(i, err)
		}

		result, err := client.Search(wyc.Id(item.Id())).FindItems()
		if err != nil {
			t.Error(i, err)
		}

		if len(result) != 0 {
			t.Error(i, "Expected not to find item with id", item.Id(), "found", len(result))
		}

		itemLinks, err := client.Search(wyc.LinkSource(item.Id())).Or(wyc.LinkDestination(item.Id())).FindLinks()
		if err != nil {
			t.Error(i, err)
		}

		if len(itemLinks) != 0 {
			t.Error(i, "Expected not to find no links to or from id", item.Id(), "found", len(itemLinks))
		}
	}
}

func TestItemParent(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}

	_, err = client.CreateCollection(randomString())
	if err != nil { // add random extra collection to muddy the waters
		t.Skip(err)
	}

	cases := []struct{
		Type string
		Variant string
	} {
		{"house", "brick"},
		{"person", "male"},
	}

	for i, tst := range cases {
		// act
		item, err := collection.CreateItem(tst.Type, tst.Variant)
		if err != nil {
			t.Skip(i, err)
		}

		parent, err := item.Parent()
		if err != nil {
			t.Error(i, err)
		}

		// assert
		if parent.Id() != collection.Id() {
			t.Error(i, "Expected colletion with id", collection.Id(), "got", parent.Id())
		}
	}
}

func TestItemLinkTo(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err) // can't perform test
	}

	fromCentreLinkName := "link"
	toCentreLinkName := "centre"

	cases := []struct{
		Type string
		Variant string
		ExtraLinkName string
		LinkFacets map[string]string
	} {
		{"house", "brick", "foo", map[string]string{"a": "b"}},
		{"tree", "oak", "bar", map[string]string{"madeBy": "batman"}},
		{"person", "male", "moo", map[string]string{"a": "b", "foo": "moo"}},
	}

	centre, err := collection.CreateItem("super", "item") // we'll link all to this
	if err != nil {
		t.Skip(err)
	}

	var previous *wyc.Item

	for i, tst := range cases {
		item, err := collection.CreateItem(tst.Type, tst.Variant)
		if err != nil {
			t.Skip(i, err)
		}

		link, err := centre.LinkTo(fromCentreLinkName, item, wyc.Facets(tst.LinkFacets))
		if err != nil {
			t.Error(i, err)
		}

		_, err = item.LinkTo(toCentreLinkName, centre)
		if err != nil {
			t.Error(i, err)
		}

		if previous != nil { // random extra link
			_, err = item.LinkTo(tst.ExtraLinkName, previous)
			if err != nil {
				t.Error(i, err)
			}
		}

		// assert
		linked, err := centre.Linked()
		if err != nil {
			t.Error(i, err)
		}

		linkedItems, _ := linked[fromCentreLinkName]
		if len(linkedItems) != i + 1 {
			if err != nil {
				t.Error(i, "Expected", i + 1, "links with name", fromCentreLinkName, "got", len(linkedItems))
			}
		}

		linked, err = item.Linked()
		if err != nil {
			t.Error(i, err)
		}

		linkedItems, _ = linked[toCentreLinkName]
		if len(linkedItems) != 1 {
			t.Error(i, "Expected 1 link to centre item with name", toCentreLinkName)
		}

		if previous != nil {
			linkedItems, _ = linked[tst.ExtraLinkName]
			if len(linkedItems) != 1 {
				t.Error(i, "Expected 1 link to previous item with name", tst.ExtraLinkName)
			}
		}

		if !facetsContain(tst.LinkFacets, link.Facets()) {
			t.Error(i, "Expected link facets", link.Facets(), "to contain", tst.LinkFacets)
		}

		previous = item
	}

}

func TestItemUpdateFacets(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err) // can't perform test
	}

	cases := []struct{
		Type string
		Variant string
		Facets map[string]string
	} {
		{"house", "brick", map[string]string{"a": "x"}},
		{"tree", "oak", map[string]string{"a": "b", "d": "/aweaw/aweaw/ewea/01"}},
		{"person", "male", map[string]string{"a": "b", "c": "9187263-198273812-198263912-123"}},
	}

	for i, tst := range cases {
		expected := map[string]string{
			wyc.FacetCollection: collection.Name(),
			wyc.FacetItemType: tst.Type,
			wyc.FacetItemVariant: tst.Variant,
		}
		for k, v := range tst.Facets {
			expected[k] = v
		}

		item, err := collection.CreateItem(tst.Type, tst.Variant)
		if err != nil {
			t.Skip(i, err)
		}

		// act
		err = item.SetFacets(tst.Facets)
		if err != nil {
			t.Error(i, err)
		}

		remote, err := client.Search(wyc.Id(item.Id())).FindItems(wyc.Limit(1))
		if err != nil {
			t.Error(i, err)
		}
		if len(remote) != 1 {
			t.Error(i, "Did not find item with id", item.Id())
		}

		// assert
		if !facetsContain(item.Facets(), expected) {
			t.Error(i, "[Facets] Expected", expected, "got", item.Facets())
		}
		if !facetsContain(remote[0].Facets(), expected) {
			t.Error(i, "[Facets] Expected", expected, "got", remote[0].Facets())
		}
	}
}

func TestCreateItemFailsOnDuplicate(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err) // can't perform test
	}

	cases := []struct{
		Type string
		Variant string
		ShouldFail bool
	}{
		{"tree", "elm", false},
		{"tree", "oak", false},
		{"tree", "oak", true},
		{"professor", "oak",false},
	}

	for i, tst := range cases {
		_, err := collection.CreateItem(tst.Type, tst.Variant)

		if tst.ShouldFail && err == nil {
			t.Error(i, "Expected failure when creating", tst)
		} else if err != nil && !tst.ShouldFail {
			t.Error(i, "Expected success when creating", tst, "got:", err)
		}
	}
}

func TestCreateItemDuplicatesInvalid(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err) // can't perform test
	}

	collection2, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err) // can't perform test
	}

	itype := "foo"
	ivar := "bar"

	// act & assert
	_, err = collection.CreateItem(itype, ivar)
	if err != nil {
		t.Error(err)
	}

	_, err = collection.CreateItem(itype, ivar)
	if err == nil {
		t.Error("Expected creation of second item ", itype, ivar, "as child of", collection.Id(), collection.Name(), "to fail")
	}

	_, err = collection2.CreateItem(itype, ivar)
	if err != nil {
		t.Error("Did not expect creationg of item", itype, ivar, "as child of", collection2.Id(), collection2.Name(), "to fail, got:", err)
	}
}

func TestCreateItem(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err) // can't perform test
	}

	cases := []struct{
		Type string
		Variant string
		Facets map[string]string
	}{
		{"house", "brick", nil},
		{"tree", "oak", map[string]string{"a": "b"}},
		{"person", "male", map[string]string{"a": "b", "c": "9187263-198273812-198263912-123"}},
	}

	for i, tst := range cases {
		// act
		result, err := collection.CreateItem(tst.Type, tst.Variant, wyc.Facets(tst.Facets))
		if err != nil {
			t.Error(err)
		}

		tmp, err := collection.Items(wyc.Id(result.Id()))
		if err != nil {
			t.Error(err)
		}
		if len(tmp) != 1 {
			t.Error("Unable to find newly created item by id")
		}
		remote := tmp[0]

		parent, err := result.Parent()

		// assert
		if result.Id() == "" {
			t.Error(i, "Expected non empty Id field")
		}
		if tst.Facets != nil {
			if !facetsContain(tst.Facets, result.Facets()) {
				t.Error(i, "[Facets] Expected", tst.Facets , "got", result.Facets())
			}
		}
		if result.Facets()[wyc.FacetCollection] != collection.Name() {
			t.Error(i, "[ParentId] FacetCollection", collection.Name() , "got", result.ParentId())
		}
		if result.Type() != tst.Type {
			t.Error(i, "[Type] Expected", tst.Type , "got", result.Type())
		}
		if result.Variant() != tst.Variant {
			t.Error(i, "[Variant] Expected", tst.Variant , "got", result.Variant())
		}
		if result.ParentId() != collection.Id() {
			t.Error(i, "[ParentId] Expected", tst.Facets , "got", result.Facets())
		}

		if remote.Id() == "" {
			t.Error(i, "Expected non empty Id field")
		}
		if tst.Facets != nil {
			if !facetsContain(tst.Facets, remote.Facets()) {
				t.Error(i, "[Facets] Expected", tst.Facets , "got", remote.Facets())
			}
		}
		if remote.Facets()[wyc.FacetCollection] != collection.Name() {
			t.Error(i, "[ParentId] FacetCollection", collection.Name() , "got", remote.ParentId())
		}
		if remote.Type() != tst.Type {
			t.Error(i, "[Type] Expected", tst.Type , "got", remote.Type())
		}
		if remote.Variant() != tst.Variant {
			t.Error(i, "[Variant] Expected", tst.Variant , "got", remote.Variant())
		}
		if remote.ParentId() != collection.Id() {
			t.Error(i, "[ParentId] Expected", tst.Facets , "got", remote.Facets())
		}	

		if parent == nil {
			t.Error(i, "[Parent] Expected parent obj got error:", err)
		} else if parent.Id() != collection.Id() {
			t.Error(i, "[Parent] Expected", collection.Id() , "got", parent.Id())
		}
	}
}
