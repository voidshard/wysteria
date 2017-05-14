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
	forest_01, err := items[0].PublishedVersion()
	if err != nil {
		panic(err)
	}

	// Retrieve all linked versions
	linked_versions, err := forest_01.Linked()
	if err != nil {
		panic(err)
	}

	for link_name, found_versions := range linked_versions {
		for _, version := range found_versions {
			fmt.Println(link_name, version.Version())
		}
	}
	//elm 1
	//pine 1
	//oak 2

	// We can also grab links with a specific name
	desired_link_name := "elm"
	linked, err := forest_01.Linked(wysteria.Name(desired_link_name))
	if err != nil {
		panic(err)
	}
	for _, version := range linked[desired_link_name] {
		fmt.Println(desired_link_name, version.Version())
	}
	//elm 1
}
