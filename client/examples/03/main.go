package main

import (
	"fmt"
	wysteria "github.com/voidshard/wysteria/client"
)

func main() {
	// Example 03: Searching for things
	//  Rather than starting at the collection and grabbing everything, we'd like to
	//  be able to find exactly what we're looking for.

	client, err := wysteria.New()
	if err != nil {
		panic(err)
	}

	// The search function allows you to search for various named fields, and arbitrary user set facets
	// Let's find all our default resources

	resources, err := client.Search(wysteria.Name("default")).FindResources()
	if err != nil {
		panic(err)
	}

	fmt.Println("--default resources--")
	for _, r := range resources {
		fmt.Println(r.Name(), r.Type(), r.Location())
	}

	//--Resources--
	//default png url://images/oak01.png
	//default png /path/to/elm01.png
	//default png /path/to/pine01.png
	//default png /other/images/oak02.png
	//default png /other/images/oak03.png

	// A single search works like an "and"
	items, err := client.Search(
		wysteria.ItemType("tree"), wysteria.ItemVariant("oak"),
	).FindItems()
	if err != nil {
		panic(err)
	}

	fmt.Println("--tree/oak items--")
	for _, i := range items {
		fmt.Println(i.Type(), i.Variant())
	}
	//--tree/oak items--
	//tree oak

	// You can search for multiple things at once via an "Or" call like so
	items, err = client.Search(
		wysteria.ItemType("tree"),
		wysteria.ItemVariant("oak"),
	).Or(
		wysteria.ItemType("tree"),
		wysteria.ItemVariant("elm"),
	).FindItems()
	if err != nil {
		panic(err)
	}

	fmt.Println("--tree/oak or tree/elm items--")
	for _, i := range items {
		fmt.Println(i.Type(), i.Variant())
	}
	//--tree/oak or tree/elm items--
	//tree oak
	//tree elm
}
