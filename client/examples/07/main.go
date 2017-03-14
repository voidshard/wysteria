package main

import (
	wysteria "github.com/voidshard/wysteria/client"
	"fmt"
)

func main () {
	// Example 07: Using links

	client, err := wysteria.New()
	if err != nil {
		panic(err)
	}

	// Look up our forest map item
	items, err := client.Search().ItemType("map").ItemVariant("forest").Items()
	if err != nil {
		panic(err)
	}
	if len(items) != 1 {
		panic(fmt.Sprintf("There should be one item found, got %d", len(items)))
	}

	// Retrieve the current published version (we linked the Version, but we could have linked
	// items in the same manner
	forest_01, err := items[0].GetPublished()
	if err != nil {
		panic(err)
	}

	// Retrieve all linked versions
	linked_versions, err := forest_01.GetLinked()
	if err != nil {
		panic(err)
	}

	for _, version := range linked_versions {
		fmt.Println(version.Link().Name(), version.Version())
	}
	//elm 1
	//pine 1
	//oak 2

	// We can also grab links with a specific name
	linked_elms, err := forest_01.GetLinkedByName("elm")
	if err != nil {
		panic(err)
	}
	for _, version := range linked_elms {
		fmt.Println(version.Link().Name(), version.Version())
	}
	//elm 1
}
