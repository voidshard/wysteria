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

	// We can also search without any search terms.
	//  Note that internally Wysteria will cap searches that include no query args
	fmt.Println("\n--collections--")
	collections, _ := client.Search().FindCollections()
	for _, c := range collections {
		fmt.Println(c.Name())
	}

	fmt.Println("--items--")
	items, _ = client.Search().FindItems()
	for _, i := range items {
		fmt.Println(i.Type(), i.Variant())
	}

	fmt.Println("--versions--")
	versions, _ := client.Search().FindVersions()
	for _, v := range versions {
		fmt.Println(v.Version())
	}

	fmt.Println("--resources--")
	resources, _ = client.Search().FindResources()
	for _, r := range resources {
		fmt.Println(r.Name(), r.Type(), r.Name())
	}

	fmt.Println("--links--")
	links, _ := client.Search().FindLinks()
	for _, l := range links {
		fmt.Println(l.Name(), l.SourceId(), l.DestinationId())
	}
}
