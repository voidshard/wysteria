package main

import (
	wysteria "github.com/voidshard/wysteria/client"
	"fmt"
)

func main() {
	// Example 06: Searching for custom fields

	client, err := wysteria.New()
	if err != nil {
		panic(err)
	}

	// the fields we'll want to search for
	facets := map[string]string{
		"publisher": "batman",
	}

	// set up our search
	batman_search := client.Search().HasFacets(facets)

	// grab matching items
	items, err := batman_search.Items()
	if err != nil {
		panic(err)
	}

	fmt.Println("Found items:")
	for _, item := range items {
		fmt.Println(item.Type(), item.Variant())
	}
	//Found items:
	//tree oak
	//tree redwood

	// grab matching versions
	versions, err := batman_search.Versions()

	fmt.Println("Found versions:")
	if err != nil {
		panic(err)
	}
	for _, version := range versions {
		fmt.Println(version.Version())
	}
	//Found versions:
	//2
	//1
}
