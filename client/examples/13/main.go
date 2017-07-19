package main

import (
	wysteria "github.com/voidshard/wysteria/client"
	"fmt"
)

func main() {
	// Example 12: Getting sub collections

	client, err := wysteria.New()
	if err != nil {
		panic(err)
	}

	projectCollection, err := client.Collection("theFooProject")
	if err != nil {
		panic(err)
	}

	children, _ := projectCollection.Collections()
	for _, child := range children {
		fmt.Println(child.Name())
	}
	// maps
}
