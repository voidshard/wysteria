package main

import (
	"fmt"
	wysteria "github.com/voidshard/wysteria/client"
)

func main() {
	// Example 05: Adding metadata: the sequel
	//  Alternatively, creating items / versions with metadata out of the box is more efficient
	//  as we don't have to make two round trips - handy if we know all the facets we want up front!

	client, err := wysteria.New()
	if err != nil {
		panic(err)
	}

	// Grab our collection
	tiles, err := client.Collection("tiles")
	if err != nil {
		panic(err)
	}

	// Create an item with some metadata
	facets := map[string]string{
		"publisher": "batman",
	}
	redwood, err := tiles.CreateItem(
		"tree",
		"redwood",
		facets,
	)
	if err != nil {
		panic(err)
	}

	// Add a version
	_, err = redwood.CreateVersion(facets)
	if err != nil {
		panic(err)
	}

	fmt.Println("Created with metadata")
}
