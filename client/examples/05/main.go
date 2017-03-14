package main

import (
	wysteria "github.com/voidshard/wysteria/client"
)

func main() {
	// Example 05: Adding metadata: the sequel
	//  Alternatively, creating items / versions with metadata out of the box is more efficient

	client, err := wysteria.New()
	if err != nil {
		panic(err)
	}

	// Grab our collection
	tiles, err := client.GetCollection("tiles")
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
}
