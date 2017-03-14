package main

import (
	wysteria "github.com/voidshard/wysteria/client"
	"fmt"
)

func main() {
	// Example 04: Adding metadata
	//  Default searching can get you a long way, but let's assume you've got custom
	//  stuff that you want to be able to search for.
	//  Here's how you'd set that on existing items / versions.

	client, err := wysteria.New()
	if err != nil {
		panic(err)
	}

	// Let's find the latest published version for our oak tile, lookup the parent Item first
	items, err := client.Search().ItemType("tree").ItemVariant("oak").Items()
	if err != nil {
		panic(err)
	}
	if len(items) != 1 {
		panic(fmt.Sprintf("There should be one item found, got %d", len(items)))
	}

	// Ok cool, now we can grab the latest published version
	version, err := items[0].GetPublished()
	if err != nil {
		panic(err)
	}

	// Set some metadata on the item
	err = items[0].SetFacets(map[string]string{
		"publisher": "batman",
	})
	if err != nil {
		panic(err)
	}

	// Set some metadata on the version
	err = version.SetFacets(map[string]string{
		"publisher": "batman",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Metadata added")
}
