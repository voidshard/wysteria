package main

import (
	wysteria "github.com/voidshard/wysteria/client"
	"log"
)

func main () {
	// Example 09: Deleting things
	//  Deleting something will daisy-chain delete everything downstream too - all the way down to links.
	//  But the things you link to wont be touched.
	//
	//  Deleting is slow & isn't to be considered atomic - it isn't intended that this operation be used
	//  frequently, nor on realms that other people are using.
	//
	//  Whether something is hard or soft deleted is left to the implementations & configuration of the
	//  data and searchbase layers.

	client, err := wysteria.New()
	if err != nil {
		panic(err)
	}

	collection, err := client.GetCollection("maps")
	if err != nil {
		panic(err)
	}
	log.Println("Delete maps:", collection.Delete())

	collection, err = client.GetCollection("tiles")
	if err != nil {
		panic(err)
	}
	log.Println("Delete tiles:", collection.Delete())
}
