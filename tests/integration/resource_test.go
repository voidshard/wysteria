package integration

import (
	"testing"
	wyc "github.com/voidshard/wysteria/client"
)

func TestDeleteResource(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}

	item, err := collection.CreateItem("foo", "bar")
	if err != nil {
		t.Skip(err)
	}

	version, err := item.CreateVersion()
	if err != nil {
		t.Skip(err)
	}

	resource, err := version.AddResource("a","b", "c")
	if err != nil {
		t.Error(err)
	}

	// act
	err = resource.Delete()
	if err != nil {
		t.Error(err)
	}

	found, err := client.Search(wyc.Id(resource.Id())).FindResources()
	if err != nil {
		t.Error(err)
	}

	// assert
	if len(found) != 0 {
		t.Error("Expected not to find anything, found", len(found), "resource(s)")
	}

}

func TestResourceGetParent(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}

	item, err := collection.CreateItem("foo", "bar")
	if err != nil {
		t.Skip(err)
	}

	version, err := item.CreateVersion()
	if err != nil {
		t.Skip(err)
	}

	resource, err := version.AddResource("a", "b", "c")
	if err != nil {
		t.Error(err)
	}

	// act
	parent, err := resource.Parent()
	if err != nil {
		t.Error(err)
	}
	
	// assert
	if parent.Id() != version.Id() {
		t.Error("Expected version with id", version.Id(), "got", parent.Id())
	}
}

func TestResourceSetFacets(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}

	item, err := collection.CreateItem("foo", "bar")
	if err != nil {
		t.Skip(err)
	}

	version, err := item.CreateVersion()
	if err != nil {
		t.Skip(err)
	}

	resource, err := version.AddResource("a", "b", "c")
	if err != nil {
		t.Error(err)
	}

	cases := []struct{
		Facets map[string]string
	} {
		{map[string]string{"a": "b"}},
		{ map[string]string{"a": "b", "q": "iuhuifawd"}},
		{ map[string]string{"foobar": "/oiaow./awd/adw"}},
		{ map[string]string{"foobar": "/daw./awd/adw", "zee": "moo"}},
	}

	expected := map[string]string{}

	// act
	for i, tst := range cases {
		for k, v := range tst.Facets {
			expected[k] = v
		}

		err = resource.SetFacets(tst.Facets)
		if err != nil {
			t.Error(i, err)
		}

		tmp, err := client.Search(wyc.Id(resource.Id())).FindResources(wyc.Limit(2))
		if err != nil {
			t.Error(i, err)
		}
		if len(tmp) != 1 {
			t.Error(i, "Expected 1 resource with id", resource.Id(), "got", len(tmp))
		}
		remote := tmp[0]

		//  assert
		if !facetsContain(tst.Facets, resource.Facets()) {
			t.Error(i, "Expected", resource.Facets(), "to contain", tst.Facets)
		}

		if !facetsContain(tst.Facets, remote.Facets()) {
			t.Error(i, "[remote] Expected", remote.Facets(), "to contain", tst.Facets)
		}
	}
}

func TestCreateResourceFailsWithDuplicateSettings(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}

	item, err := collection.CreateItem("foo", "bar")
	if err != nil {
		t.Skip(err)
	}

	version1, err := item.CreateVersion()
	if err != nil {
		t.Skip(err)
	}

	version2, err := item.CreateVersion()
	if err != nil {
		t.Skip(err)
	}

	rname := "name"
	rtype := "type"
	rloc := "location"

	// act & assert
	_, err = version1.AddResource(rname, rtype, rloc)
	if err != nil {
		t.Error(err)
	}

	_, err = version1.AddResource(rname, rtype, rloc)
	if err == nil {
		t.Error("Expected duplicate resource fail", version1.Id(), rname, rtype, rloc)
	}

	_, err = version2.AddResource(rname, rtype, rloc)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateResource(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}

	item, err := collection.CreateItem("foo", "bar")
	if err != nil {
		t.Skip(err)
	}

	version, err := item.CreateVersion()
	if err != nil {
		t.Skip(err)
	}

	cases := []struct{
		Name string
		Type string
		Location string
		Facets map[string]string
	} {
		{"foo", "img", "/path/to/image", map[string]string{"a": "b"}},
		{"bar", "xml", "/path/to/xml.xml", map[string]string{"a": "b", "q": "iuhuifawd"}},
		{"baz", "uuid", "8964829374hhludhoi3whr8w7r", map[string]string{"foobar": "/oiaow./awd/adw"}},
		{"moo", "xxx", ">//?o@^#)(^@(#^^%!@$/", map[string]string{}},
		{"moo", "boo", ">dawd@$/", nil},
	}

	for i, tst := range cases {
		// act
		resource, err := version.AddResource(tst.Name, tst.Type, tst.Location, wyc.Facets(tst.Facets))
		if err != nil {
			t.Error(i, err)
		}

		found, err := version.Resources(wyc.Id(resource.Id()))
		if err != nil {
			t.Error(i, err)
		}
		if len(found) != 1 {
			t.Error(i, "Expected to find 1 resource, found", len(found))
		}
		remote := found[0]


		allResources, err := version.Resources()
		if err != nil {
			t.Error(i, err)
		}

		// assert
		if len(allResources) != i + 1 {
			t.Error(i, "Expected", i + 1, "get", len(allResources))
		}

		if resource.Id() == "" {
			t.Error(i, "Expected Id to be set")
		}
		if resource.ParentId() != version.Id() {
			t.Error(i, "Expected ParentId to be", version.Id(), "got", resource.ParentId())
		}
		if resource.Name() != tst.Name {
			t.Error(i, "Expected", tst.Name, "got", resource.Name())
		}
		if resource.Type() != tst.Type {
			t.Error(i, "Expected", tst.Type, "got", resource.Type())
		}
		if resource.Location() != tst.Location {
			t.Error(i, "Expected", tst.Location, "got", resource.Location())
		}
		if !facetsContain(tst.Facets, resource.Facets()) {
			t.Error(i, "Expected", tst.Facets, "got", resource.Facets())
		}

		if remote.Id() == "" {
			t.Error(i, "[remote] Expected Id to be set")
		}
		if remote.ParentId() != version.Id() {
			t.Error(i, "[remote] Expected ParentId to be", version.Id(), "got", remote.ParentId())
		}
		if remote.Name() != tst.Name {
			t.Error(i, "[remote] Expected", tst.Name, "got", remote.Name())
		}
		if remote.Type() != tst.Type {
			t.Error(i, "[remote] Expected", tst.Type, "got", remote.Type())
		}
		if remote.Location() != tst.Location {
			t.Error(i, "[remote] Expected", tst.Location, "got", remote.Location())
		}
		if !facetsContain(tst.Facets, remote.Facets()) {
			t.Error(i, "[remote] Expected", tst.Facets, "got", remote.Facets())
		}
	}
}