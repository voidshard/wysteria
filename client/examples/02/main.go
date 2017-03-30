package main

import (
	"fmt"
	wysteria "github.com/voidshard/wysteria/client"
)

func main() {
	// Example 02: Getting things
	//  Before we get into searching, let's try just grabbing the children of each object
	//  on the way down.

	client, err := wysteria.New()
	if err != nil {
		panic(err)
	}

	// Find our collection
	col, err := client.GetCollection("tiles")
	if err != nil {
		panic(err)
	}
	fmt.Println("-- Collection --")
	fmt.Println(col.Name())
	//-- Collection --
	//tiles

	// Get all child items
	items, err := col.GetItems()
	if err != nil {
		panic(err)
	}
	fmt.Println("-- Items --")
	for _, i := range items {
		fmt.Println(i.Type(), i.Variant())
	}
	//-- Items --
	//tree oak
	//tree elm
	//tree pine

	// Look up all published versions of our items
	fmt.Println("-- Published Versions --")
	for _, i := range items {
		published, err := i.GetPublished()
		if err != nil {
			panic(err)
		}
		fmt.Println(published.Version())

		// Get & print attached resources for our version
		resources, err := published.GetAllResources()
		if err != nil {
			panic(err)
		}
		for _, r := range resources {
			fmt.Println("  [resource]", r.Name(), r.Type(), r.Location())
		}
	}

	//-- Published Versions --
	//2
	//  [resource] default png /other/images/oak02.png
	//1
	//  [resource] default png /path/to/elm01.png
	//  [resource] lowres jpeg /path/lowres/image.jpeg
	//1
	//  [resource] default png /path/to/pine01.png
	//  [resource] lowres jpeg /path/lowres/image.jpeg
	//  [resource] stats json url://file.json
}
