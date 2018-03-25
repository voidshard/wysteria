package main

import (
	"testing"
	"github.com/fgrid/uuid"
	wyc "github.com/voidshard/wysteria/client"
)

// Return random uuid4 (as a string)
//
func randomString() string {
	return uuid.NewV4().String()
}

//
func newClient(t *testing.T) *wyc.Client {
	client, err := wyc.New(
		wyc.Host("localhost:31000"), wyc.Driver("grpc"),
	)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	return client
}

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