package main

import (
	"fmt"
	wysteria "github.com/voidshard/wysteria/client"
)

func main() {
	// Example 07: Using links
	//  Links are an arbitrary way of describing that things are related.
	//  Both items and versions can be linked, though versions only to versions,
	//  and items only to items.

	client, err := wysteria.New()
	if err != nil {
		panic(err)
	}

	// Look up our forest map item
	items, err := client.Search(wysteria.ItemType("map"), wysteria.ItemVariant("forest")).FindItems()
	if err != nil {
		panic(err)
	}
	if len(items) != 1 {
		panic(fmt.Sprintf("There should be one item found, got %d", len(items)))
	}

	// Retrieve the current published version (we linked the Version, but we could have linked
	// items in the same manner
	forest01, err := items[0].PublishedVersion()
	if err != nil {
		panic(err)
	}

	// Retrieve all linked versions
	linkedVersions, err := forest01.Linked()
	if err != nil {
		panic(err)
	}

	for linkName, foundVersions := range linkedVersions {
		for _, version := range foundVersions {
			fmt.Println(linkName, version.Version())
		}
	}
	//elm 1
	//pine 1
	//oak 2

	// We can also grab links with a specific name
	desiredLinkName := "elm"
	linked, err := forest01.Linked(wysteria.Name(desiredLinkName))
	if err != nil {
		panic(err)
	}
	for _, version := range linked[desiredLinkName] {
		fmt.Println(desiredLinkName, version.Version())
	}
	//elm 1
}
