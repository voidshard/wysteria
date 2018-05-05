package integration

import (
	wyc "github.com/voidshard/wysteria/client"
	"testing"
)

/*
Item & Version test files test the creation & fetching of links.
*/

func TestLinkSetFacets(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err) // can't perform test
	}

	item1, err := collection.CreateItem("super1", "item")
	if err != nil {
		t.Skip(err)
	}

	item2, err := collection.CreateItem("super2", "item")
	if err != nil {
		t.Skip(err)
	}

	cases := []struct{
		Facets map[string]string
	} {
		{map[string]string{"boo": "moo", "hithere": "/omadw/ad", "71dd": "jq9197ee9bweyb8 fy8f8wgf--dqiubwdiyaud-+"}},
		{map[string]string{"ar": "213", "hithere": "+/", "qwd": "()_+_+_+--1203719837639481(&* &!^@#(*!%@(yaud-+"}},
		{map[string]string{"boo": "sha256:33fb5550ce42935faaf86d03284e26219 150c28b 0755 a6f9d e24cc 054e6eb40e"}},
		{map[string]string{"jq9197ee9bweyb8fy8f8wgf--dqiubwdiy aud-+": "6bf7a22e-0b8e-47c0-bc78-c4fb2219dd15"}},
	}

	link, err := item1.LinkTo(randomString(), item2)
	if err != nil {
		t.Skip(err)
	}

	for i, tst := range cases {
		// act
		err := link.SetFacets(tst.Facets)
		if err != nil {
			t.Error(i, err)
		}

		tmp, err := client.Search(wyc.Id(link.Id())).FindLinks()
		if err != nil {
			t.Error(i, err)
		}
		if len(tmp) != 1 {
			t.Error(i, "Expected to find 1 link with id", link.Id(), "found", len(tmp))
		}
		remote := tmp[0]

		// assert
		if !facetsContain(tst.Facets, link.Facets()) {
			t.Error(i, "Expected link facets", link.Facets(), "to contain", tst.Facets)
		}
		if !facetsContain(tst.Facets, remote.Facets()) {
			t.Error(i, "[remote] Expected link facets", remote.Facets(), "to contain", tst.Facets)
		}
	}
}

func TestClientLink(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}

	item, err := collection.CreateItem("TestClientLink", "TestClientLink")
	if err != nil {
		t.Skip(err)
	}

	version, err := item.CreateVersion()
	if err != nil {
		t.Skip(err)
	}

	version2, err := item.CreateVersion()
	if err != nil {
		t.Skip(err)
	}

	result, err := version.LinkTo("TestClientLink", version2)
	if err != nil {
		t.Error(err)
	}

	// act
	remote, err := client.Link(result.Uri())
	if err != nil {
		t.Error(err)
	}

	// assert
	if remote == nil {
		t.Error("did not find desired object by uri")
	}

	if remote.Uri() != result.Uri() || remote.Id() != result.Id() {
		t.Error("expected", result, "got", remote)
	}
}

func TestLinkCreationViaItems(t *testing.T) {
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

	for i, tst := range cases {
		item, err := collection.CreateItem(tst.Type, tst.Variant)
		if err != nil {
			t.Skip(i, err)
		}

		link, err := centre.LinkTo(tst.ExtraLinkName, item, wyc.Facets(tst.LinkFacets))
		if err != nil {
			t.Error(i, err)
		}

		tmp, err := client.Search(wyc.Id(link.Id())).FindLinks()
		if err != nil {
			t.Error(i, err)
		}
		if len(tmp) != 1 {
			t.Error(i, "Expected to find 1 link with id", link.Id(), "found", len(tmp))
		}
		remote := tmp[0]

		// assert
		if link.Name() != tst.ExtraLinkName {
			t.Error(i, "Expected Name", tst.ExtraLinkName, "got", link.Name())
		}
		if link.Id() == "" {
			t.Error(i, "Expected non null id")
		}
		if link.Uri() == "" {
			t.Error(i, "Expected non empty Uri field")
		}
		if link.SourceId() != centre.Id() {
			t.Error(i, "Expected Src", centre.Id(), "got", link.SourceId())
		}
		if link.DestinationId() != item.Id() {
			t.Error(i, "Expected Dst", item.Id(), "got", link.DestinationId())
		}
		if !facetsContain(tst.LinkFacets, link.Facets()) {
			t.Error(i, "Expected link facets", link.Facets(), "to contain", tst.LinkFacets)
		}

		if remote.Name() != tst.ExtraLinkName {
			t.Error(i, "[remote] Expected Name", tst.ExtraLinkName, "got", remote.Name())
		}
		if remote.Uri() == "" {
			t.Error(i, "Expected non empty Uri field")
		}
		if remote.Id() == "" {
			t.Error(i, "[remote] Expected non null id")
		}
		if remote.SourceId() != centre.Id() {
			t.Error(i, "[remote] Expected Src", centre.Id(), "got", remote.SourceId())
		}
		if remote.DestinationId() != item.Id() {
			t.Error(i, "[remote] Expected Dst", item.Id(), "got", remote.DestinationId())
		}
		if !facetsContain(tst.LinkFacets, remote.Facets()) {
			t.Error(i, "[remote] Expected link facets", remote.Facets(), "to contain", tst.LinkFacets)
		}
	}

}
