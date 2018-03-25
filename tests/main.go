
package main

import (
	wyc "github.com/voidshard/wysteria/client"
	"github.com/fgrid/uuid"
	"fmt"
)


func nClient() *wyc.Client {
	client, err := wyc.New(wyc.Host("localhost:31000"), wyc.Driver("grpc"))
	if err != nil {
		panic(err)
	}
	return client
}

func main() {
	client := nClient()
	defer client.Close()

	// Grab our collection
	tiles, err := client.CreateCollection(uuid.NewV4().String())
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
		wyc.Facets(facets),
	)
	if err != nil {
		panic(err)
	}

	found, err := client.Search(wyc.Id(redwood.Id())).FindItems()
	if err != nil {
		panic(err)
	}
	fmt.Println("Found", len(found), found)

	//// Add a version
	//_, err = redwood.CreateVersion(wyc.Facets(facets))
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println("Created with metadata")
	//
	//time.Sleep(5 * time.Second)



}