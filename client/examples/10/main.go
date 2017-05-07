package main

import (
	wysteria "github.com/voidshard/wysteria/client"
	"log"
)

func main() {
	// Example 10: Limit & offset
	// Limiting query sizes & pagination are also possible and are achieved in the same manner you might expect.
	// By default a new Search has a limit set to a default value (500 at the time of writing) with an offset
	// of 0.

	client, err := wysteria.New()
	if err != nil {
		panic(err)
	}

	for i := 0; i < 3; i++ {
		resources, err := client.Search().Name("default").Limit(1).Offset(i).FindResources()
		if err != nil {
			panic(err)
		}
		log.Println("Found", len(resources), "=>", resources[0].Name(), resources[0].Location())
	}
	// Found 1 => default url://images/oak01.png
	// Found 1 => default /path/to/elm01.png
	// Found 1 => default /path/to/pine01.png
}
