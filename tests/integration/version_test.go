package integration

import (
	wyc "github.com/voidshard/wysteria/client"
	"testing"
)

func TestVersionDeletion(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}

	item, err := collection.CreateItem("super", "item")
	if err != nil {
		t.Skip(err)
	}

	fooitem, err := collection.CreateItem("foo", "item")
	if err != nil {
		t.Skip(err)
	}
	centre, err := fooitem.CreateVersion()
	if err != nil {
		t.Skip(err)
	}

	for i := 0; i < 10; i ++ {
		version, err := item.CreateVersion()
		if err != nil {
			t.Skip(err)
		}

		_, err = version.LinkTo("somelinkname", centre)
		if err != nil {
			t.Error(i, err)
		}

		_, err = version.AddResource("a", "b", "c")
		if err != nil {
			t.Error(i, err)
		}
	}

	// act & assert
	err = centre.Delete()
	if err != nil {
		t.Error(err)
	}

	foundVersions, err := client.Search(wyc.Id(centre.Id())).FindVersions()
	if err != nil {
		t.Error(err)
	}
	if len(foundVersions) != 0 {
		t.Error("Expect not to find version with id", centre.Id(), "found", len(foundVersions))
	}

	foundResources, err := client.Search(wyc.ChildOf(centre.Id())).FindResources()
	if err != nil {
		t.Error(err)
	}
	if len(foundResources) != 0{
		t.Error("Expect not to child resources of", centre.Id(), "found", len(foundResources))
	}

	foundLinks, err := client.Search(wyc.LinkSource(centre.Id()), wyc.LinkDestination(centre.Id())).FindLinks()
	if err != nil {
		t.Error(err)
	}
	if len(foundLinks) != 0{
		t.Error("Expect not to links to/from", centre.Id(), "found", len(foundLinks))
	}
}

func TestVersionLinkTo(t *testing.T) {
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
		ExtraLinkName string
		LinkFacets map[string]string
	} {
		{ "foo", map[string]string{"a": "b"}},
		{"bar", map[string]string{"madeBy": "batman"}},
		{ "moo", map[string]string{"a": "b", "foo": "moo"}},
	}

	item, err := collection.CreateItem("super", "item")
	if err != nil {
		t.Skip(err)
	}

	centre, err := item.CreateVersion() // we'll link all to this
	if err != nil {
		t.Skip(err)
	}

	var previous *wyc.Version

	for i, tst := range cases {
		version, err := item.CreateVersion()
		if err != nil {
			t.Skip(i, err)
		}

		link, err := centre.LinkTo(fromCentreLinkName, version, wyc.Facets(tst.LinkFacets))
		if err != nil {
			t.Error(i, err)
		}

		_, err = version.LinkTo(toCentreLinkName, centre)
		if err != nil {
			t.Error(i, err)
		}

		if previous != nil { // random extra link
			_, err = version.LinkTo(tst.ExtraLinkName, previous)
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

		linked, err = version.Linked()
		if err != nil {
			t.Error(i, err)
		}

		linkedItems, _ = linked[toCentreLinkName]
		if len(linkedItems) != 1 {
			t.Error(i, "Expected 1 link to centre item with name", toCentreLinkName)
		}

		if !facetsContain(tst.LinkFacets, link.Facets()) {
			t.Error(i, "Expected link facets", link.Facets(), "to contain", tst.LinkFacets)
		}

		if previous != nil {
			linkedItems, _ = linked[tst.ExtraLinkName]
			if len(linkedItems) != 1 {
				t.Error(i, "Expected 1 link to previous item with name", tst.ExtraLinkName)
			}
		}

		previous = version
	}

}

func TestVersionSetFacets(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err) // can't perform test
	}

	item, err := collection.CreateItem("foo", "bar")
	if err != nil {
		t.Skip(err) // can't perform test
	}

	cases := []struct{
		Facets map[string]string
	} {
		{map[string]string{"boo": "moo", "hithere": "/omadw/ad", "71dd": "jq9197ee9bweyb8 fy8f8wgf--dqiubwdiyaud-+"}},
		{map[string]string{"ar": "213", "hithere": "+/", "qwd": "()_+_+_+--1203719837639481(&* &!^@#(*!%@(yaud-+"}},
		{map[string]string{"boo": "sha256:33fb5550ce42935faaf86d03284e26219 150c28b 0755 a6f9d e24cc 054e6eb40e"}},
		{map[string]string{"jq9197ee9bweyb8fy8f8wgf--dqiubwdiy aud-+": "6bf7a22e-0b8e-47c0-bc78-c4fb2219dd15"}},
	}

	for i, tst := range cases {
		version, err := item.CreateVersion()
		if err != nil {
			t.Error(i, err)
		}

		// act
		err = version.SetFacets(tst.Facets)

		results, err := client.Search(wyc.Id(version.Id())).FindVersions()
		if err != nil {
			t.Error(i, err)
		}
		if len(results) != 1 {
			t.Error(i, "Expected to find version with id", version.Id(), "found", len(results))
		}
		remote := results[0]

		// assert
		if !facetsContain(tst.Facets, version.Facets()) {
			t.Error(i, "Expected facets", version.Facets(), "to contain all of", tst.Facets)
		}

		if !facetsContain(tst.Facets, remote.Facets()) {
			t.Error(i, "[remote] Expected facets", remote.Facets(), "to contain all of", tst.Facets)
		}
	}
}

func TestVersionParent(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Error(err)
	}

	item, err := collection.CreateItem("batman", "moo")
	if err != nil { // add random extra collection to muddy the waters
		t.Skip(err)
	}

	fakeCol, err := client.CreateCollection(randomString())
	if err != nil {
		t.Error(err)
	}
	fakeCol.CreateItem("foo", "bar")

	for i := 0 ; i < 5 ; i ++ {
		version, err := item.CreateVersion()
		if err != nil {
			t.Error(i, err)
		}

		// act
		result, err := version.Parent()
		if err != nil {
			t.Error(i, err)
		}

		// assert
		if result.Id() != item.Id() {
			t.Error(i, "Expected parent with id", item.Id(), "got", result.Id())
		}
	}
}

func TestClientVersion(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err)
	}

	item, err := collection.CreateItem("TestClientVersion", "TestClientVersion")
	if err != nil {
		t.Skip(err)
	}

	result, err := item.CreateVersion()
	if err != nil {
		t.Skip(err)
	}

	// act
	remote, err := client.Version(result.Uri())
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

func TestCreateVersion(t *testing.T) {
	// arrange
	client := newClient(t)
	defer client.Close()

	collection, err := client.CreateCollection(randomString())
	if err != nil {
		t.Skip(err) // can't perform test
	}

	item, err := collection.CreateItem("foo", "bar")
	if err != nil {
		t.Skip(err) // can't perform test
	}

	cases := []struct{
		Facets map[string]string
	} {
		{map[string]string{"boo": "moo", "hithere": "/omadw/ad", "71dd": "jq9197ee9bweyb8 fy8f8wgf--dqiubwdiyaud-+"}},
		{map[string]string{"ar": "213", "hithere": "+/", "qwd": "()_+_+_+--1203719837639481(&* &!^@#(*!%@(yaud-+"}},
		{map[string]string{"boo": "sha256:33fb5550ce42935faaf86d03284e26219 150c28b 0755 a6f9d e24cc 054e6eb40e"}},
		{map[string]string{"jq9197ee9bweyb8fy8f8wgf--dqiubwdiy aud-+": "6bf7a22e-0b8e-47c0-bc78-c4fb2219dd15"}},
	}

	var previous *wyc.Version

	for i, tst := range cases {
		// act
		version, err := item.CreateVersion(wyc.Facets(tst.Facets))
		if err != nil {
			t.Error(i, err)
		}

		found, err := client.Search(wyc.Id(version.Id()), wyc.VersionNumber(version.Version())).FindVersions()
		if err != nil {
			t.Error(i, err)
		}
		if len(found) != 1 {
			t.Error(i, "Expected 1 version with id", version.Id(), version.Version())
		}
		remote := found[0]

		err = version.Publish()
		if err != nil {
			t.Error(i, err)
		}

		published, err := item.PublishedVersion()
		if err != nil {
			t.Error(i, err)
		}

		// assert
		if published.Id() != version.Id() {
			t.Error(i, "Expected published version to be", version.Id(), "got", published.Id())
		}

		if version.Version() != int32(i + 1) {
			t.Error(i, "Expected version number", i+1, "got", version.Version())
		}
		if version.Uri() == "" {
			t.Error(i, "Expected non empty Uri field")
		}
		if version.Id() == "" {
			t.Error(i, "Expected version Id to be set")
		}
		if !facetsContain(tst.Facets, version.Facets()) {
			t.Error(i, "Expected facets", tst.Facets, "got", version.Facets())
		}

		if remote.Version() != int32(i + 1) {
			t.Error(i, "[remote] Expected version number", i + 1, "got", remote.Version())
		}
		if remote.Id() == "" {
			t.Error(i, "[remote] Expected version Id to be set")
		}
		if remote.Uri() == "" {
			t.Error(i, "Expected non empty Uri field")
		}
		if !facetsContain(tst.Facets, remote.Facets()) {
			t.Error(i, "[remote] Expected facets", tst.Facets, "got", remote.Facets())
		}

		if previous != nil {
			if version.Version() <= previous.Version() {
				t.Error(i, "Expected version number", version.Version(), "greater than previous", previous.Version())
			}

			err = previous.Publish()
			if err != nil {
				t.Error(i, err)
			}

			published, err = item.PublishedVersion()
			if err != nil {
				t.Error(i, err)
			}

			if published.Id() != previous.Id() {
				t.Error(i, "Expected published version to be", previous.Id(), "got", published.Id())
			}
		}

		previous = version
	}

}
