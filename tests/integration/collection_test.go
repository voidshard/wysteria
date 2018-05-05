package integration

import (
	"testing"
	wyc "github.com/voidshard/wysteria/client"
)



func TestCollectionParentAndChild(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	cases := []struct{
		Name string
		Facets map[string]string
	}{
		{randomString(), nil},
		{randomString(), map[string]string{"a": "b"}},
		{randomString(), map[string]string{"a": "b", "c": "1", "q": "dino"}},
	}

	parent, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}

	// add some extra collections to muddy the waters a bit
	fakeparent, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}
	fakeparent.CreateCollection(randomString())
	fakeparent.CreateCollection(randomString())
	fakeparent.CreateCollection(randomString())

	for i, tst := range cases {
		// act & assert
		child, err := parent.CreateCollection(tst.Name, wyc.Facets(tst.Facets))
		if err != nil {
			t.Error(i, err)
		}

		result, err := child.Parent()
		if err != nil {
			t.Error(i, err)
		}

		if result.Id() != parent.Id() {
			t.Error(i, "Expected parent with id", parent.Id(), "got", result.Id())
		}
	}

	children, err := parent.Collections()
	if err != nil {
		t.Error(err)
	}

	if len(children) != len(cases) {
		t.Error("Expected", len(cases), "children but got", len(children))
	}
}

func TestDeleteCollectionDeletes(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	cases := []struct{
		Name string
	}{
		{randomString()},
		{randomString()},
		{randomString()},
	}

	for i, tst := range cases {
		col, err := client.CreateCollection(tst.Name)
		if err != nil {
			t.Error(i, err)
		}

		// act
		err = col.Delete()
		if err != nil {
			t.Error(i, err)
		}

		// assert
		result, err := client.Search(wyc.Id(col.Id())).Or(wyc.Name(col.Name())).FindCollections()
		if err != nil {
			t.Error(i, err)
		}
		if len(result) > 0 {
			t.Error(i, "Expected object deleted, but it was not")
		}
	}
}

func TestCollectionFacetFunction(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	cases := []struct {
		Facets map[string]string
	}{
		{map[string]string{"a": "foo", "c": "/adoo/iad/oaowdo/foo.moo", "x": "what"}},
	}

	for i, tst := range cases {
		collection, err := client.CreateCollection(randomString(), wyc.Facets(tst.Facets))
		if err != nil {
			t.Error(err)
		}

		// act & assert
		for k, v := range tst.Facets {
			result, _ := collection.Facet(k)

			if result != v {
				t.Error(i, "Expected", v, "got", result)
			}
		}
	}
}

func TestUpdateCollectionFacets(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	cases := []struct{
		Facets map[string]string
	} {
		{map[string]string{"a": "foo", "c": "/adoo/iad/oaowdo/foo.moo", "x": "what"}},
		{map[string]string{"a": "bar", "c": "/adoo/iad/oaowdo/foo.moo", "d": "blah"}},
	}

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}

	expected := map[string]string{}
	for k, v := range collection.Facets() { // copy facets
		expected[k] = v
	}

	for i, tst := range cases {
		for k, v := range tst.Facets { // set what we expect
			expected[k] = v
		}

		// act
		err := collection.SetFacets(tst.Facets)
		if err != nil {
			t.Error(err)
		}

		result, err := client.Collection(collection.Id()) // refetch so we're looking at the db copy too
		if err != nil {
			t.Error(err)
		}

		// assert
		if !facetsContain(expected, result.Facets()) {
			t.Error(i, "[Facets: remote copy] Expected", expected , "got", result.Facets())
		}
		if !facetsContain(expected, collection.Facets()) {
			t.Error(i, "[Facets: local copy] Expected", expected , "got", collection.Facets())
		}
	}
}

func TestCreateNestedCollection(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	cases := []struct{
		Name string
		Facets map[string]string
	}{
		{"tcncFoobar", nil},
		{"tcncFoomoo", map[string]string{"a": "b"}},
		{"tcncBarfoo", map[string]string{"a": "b", "c": "9187263-198273812-198263912-123"}},
		{"tcncBarfoo", map[string]string{"a": "b", "c": "1", "q": "dino"}},
		{"tcncBarfoo", map[string]string{}},
	}

	// root level parent
	result, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}

	previous := result

	for i, tst := range cases {
		// act
		result, err := previous.CreateCollection(tst.Name, wyc.Facets(tst.Facets))
		if err != nil {
			t.Error(err)
		}

		remote, err := client.Collection(result.Id())
		if err != nil {
			t.Error(err)
		}

		// assert
		parentName := result.Facets()[wyc.FacetCollection]

		if result.Id() == "" {
			t.Error(i, "Expected non empty Id field")
		}
		if result.ParentId() != previous.Id() {
			t.Error(i, "[ParentId] Expected", previous.Id() , "got", result.ParentId())
		}
		if parentName != previous.Name() {
			t.Error(i, "[FacetCollection] Expected", previous.Name() , "got", parentName)
		}
		if tst.Facets != nil {
			if !facetsContain(tst.Facets, result.Facets()) {
				t.Error(i, "[Facets] Expected", tst.Facets , "got", result.Facets())
			}
		}
		if result.Name() != tst.Name {
			t.Error(i, "[Name] Expected", tst.Name , "got", result.Name())
		}

		if remote.Id() == "" {
			t.Error(i, "Expected non empty Id field")
		}
		if remote.ParentId() != previous.Id() {
			t.Error(i, "[ParentId: remote] Expected", previous.Id() , "got", remote.ParentId())
		}
		if remote.Facets()[wyc.FacetCollection] != previous.Name() {
			t.Error(i, "[FacetCollection: remote] Expected", previous.Name() , "got", remote.Facets()[wyc.FacetCollection])
		}
		if tst.Facets != nil {
			if !facetsContain(tst.Facets, remote.Facets()) {
				t.Error(i, "[Facets: remote] Expected", tst.Facets , "got", remote.Facets())
			}
		}
		if remote.Name() != tst.Name {
			t.Error(i, "[Name: remote] Expected", tst.Name , "got", remote.Name())
		}

		previous = result
	}
}

func TestCreateCollectionSameNameAndParentInvalid(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	cases := []struct{
		Name string
	}{
		// collections with the same name & parent aren't allowed
		{randomString()},
	}

	for i, tst := range cases {
		// act
		_, err := client.CreateCollection(tst.Name)
		if err != nil {
			t.Error(err)
		}

		result, err := client.CreateCollection(tst.Name)

		// assert
		if err == nil || result != nil {
			t.Error(i, "Expected err but got collection:", result, "test:", tst)
		}
	}
}

func TestCreateCollectionSameNameDifferentParent(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collectionOne, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}
	collectionTwo, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}

	cases := []struct{
		Name string
		ParentA *wyc.Collection
		ParentB *wyc.Collection
	} {
		// since these have different parents, this should be fine
		{"tccifoo", collectionOne, collectionTwo},
	}

	for i, tst := range cases {
		// act
		_, err := tst.ParentA.CreateCollection(tst.Name)
		if err != nil {
			t.Error(i, err)
		}

		_, err = tst.ParentB.CreateCollection(tst.Name)

		// assert
		if err != nil {
			t.Error(i, err)
		}
	}
}

func TestCreateCollection(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	root := wyc.FacetRootCollection

	cases := []struct{
		Name string
		Facets map[string]string
	}{
		{randomString(), nil},
		{randomString(), map[string]string{}},
		{randomString(), map[string]string{"a": "b"}},
		{randomString(), map[string]string{"a": "b", "c": "9187263-198273812-198263912-123"}},
	}

	for i, tst := range cases {
		// act
		result, err := client.CreateCollection(tst.Name, wyc.Facets(tst.Facets))
		if err != nil {
			t.Error(err)
		}

		remote, err := client.Collection(result.Id())
		if err != nil {
			t.Error(err)
		}
		
		// assert
		if result.Id() == "" {
			t.Error(i, "Expected non empty Id field")
		}
		if result.Uri() == "" {
			t.Error(i, "Expected non empty Uri field")
		}
		if result.ParentId() != "" {
			t.Error(i, "[ParentId] Expected [empty string] got", result.ParentId())
		}
		if result.Facets()[wyc.FacetCollection] != root {
			t.Error(i, "[FacetCollection] Expected", root , "got", result.ParentId())
		}
		if tst.Facets != nil {
			if !facetsContain(tst.Facets, result.Facets()) {
				t.Error(i, "[Facets] Expected", tst.Facets , "got", result.Facets())
			}
		}
		if result.Name() != tst.Name {
			t.Error(i, "[Name] Expected", tst.Name , "got", result.Name())
		}


		if remote.Id() == "" {
			t.Error(i, "[remote] Expected non empty Id field")
		}
		if result.Uri() == "" {
			t.Error(i, "Expected non empty Uri field")
		}
		if remote.ParentId() != "" {
			t.Error(i, "[ParentId: remote] Expected [empty string] got", remote.ParentId())
		}
		if remote.Facets()[wyc.FacetCollection] != root {
			t.Error(i, "[FacetCollection: remote] Expected", root , "got", remote.ParentId())
		}
		if tst.Facets != nil {
			if !facetsContain(tst.Facets, remote.Facets()) {
				t.Error(i, "[Facets: remote] Expected", tst.Facets , "got", remote.Facets())
			}
		}
		if remote.Name() != tst.Name {
			t.Error(i, "[Name: remote] Expected", tst.Name , "got", remote.Name())
		}
	}
}


