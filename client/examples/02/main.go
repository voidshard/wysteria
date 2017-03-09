package main

import (
	wysteria "github.com/voidshard/wysteria/client"
	"fmt"
)

func main () {
	// Example02: Getting things
	//   Generates:
	//-- Collection --
	//tiles
	//-- Items --
	//tree oak
	//tree elm
	//tree pine
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

	// Find items
	items, err := col.GetItems()
	if err != nil {
		panic(err)
	}
	fmt.Println("-- Items --")
	for _, i := range items {
		fmt.Println(i.Type(), i.Variant())
	}

	// Look up our published versions
	fmt.Println("-- Published Versions --")
	for _, i := range items {
		published, err := i.GetPublishedVersion()
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
}

