package main

import (
	"fmt"
	wysteria "github.com/voidshard/wysteria/client"
)

func main() {
	// Example 12: Sub collections

	client, err := wysteria.New()
	if err != nil {
		panic(err)
	}

	// Collections may have sub collections to help you organise things
	projectCollection, err := client.CreateCollection("theFooProject", nil)
	if err != nil {
		panic(err)
	}

	mapsOfFoo, err := projectCollection.CreateCollection("maps", nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(projectCollection.Id(), projectCollection.Name())
	fmt.Println(mapsOfFoo.Name(), "child of", mapsOfFoo.Parent())

	// Two collections with the same parent still may not have the same name
	_, err = projectCollection.CreateCollection("maps", nil)
	if err == nil {
		panic("We shouldn't be able to get here!")
	}
}
