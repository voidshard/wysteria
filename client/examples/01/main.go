package main

import (
	"fmt"
	wysteria "github.com/voidshard/wysteria/client"
)

func main() {
	// Example 01: Creating
	//  We'll use the objects we create here in following examples, so please run this first!
	//
	//  Creating things is pretty straight forward/

	client, err := wysteria.New()
	if err != nil {
		panic(err)
	}

	// Create collection
	col, err := client.CreateCollection("tiles")
	if err != nil {
		panic(err)
	}

	// Create an item in our collection
	oak, err := col.CreateItem("tree", "oak", nil)
	if err != nil {
		panic(err)
	}

	// Create the first version of our oak tree
	oak01, err := oak.CreateVersion()
	if err != nil {
		panic(err)
	}

	// Add some kind of resource(s) to our tree
	err = oak01.AddResource("default", "png", "url://images/oak01.png")
	if err != nil {
		panic(err)
	}
	oak01.AddResource("stats", "xml", "/path/to/file.xml")

	// Throw in some more trees
	elm, _ := col.CreateItem("tree", "elm")
	elm01, _ := elm.CreateVersion(nil)
	elm01.AddResource("default", "png", "/path/to/elm01.png")
	elm01.AddResource("lowres", "jpeg", "/path/lowres/image.jpeg")

	pine, _ := col.CreateItem("tree", "pine")
	pine01, _ := pine.CreateVersion(nil)
	pine01.AddResource("default", "png", "/path/to/pine01.png")
	pine01.AddResource("lowres", "jpeg", "/path/lowres/image.jpeg")
	pine01.AddResource("stats", "json", "url://file.json")

	// Shoot. Our oak tree looks awful. Let's redo it
	oak02, _ := oak.CreateVersion(nil)
	oak02.AddResource("default", "png", "/other/images/oak02.png")

	// Ok sweet. Let's publish the things
	err = oak02.Publish()
	if err != nil {
		panic(err)
	}
	elm01.Publish()
	pine01.Publish()

	// One more oak (note that oak02 is still the "published" version -> important later)
	oak03, _ := oak.CreateVersion()
	oak03.AddResource("default", "png", "/other/images/oak03.png")

	// We'll create some links for later examples too, first, something to link to
	// Just for kicks, we'll create some of these with an extra custom facet(s) too (more on this later)
	customFacets := map[string]string{
		"createdby": "batman",
	}
	maps, _ := client.CreateCollection("maps", customFacets)
	forest, _ := maps.CreateItem("map", "forest", customFacets)
	forest_map_01, _ := forest.CreateVersion(customFacets)
	forest_map_01.AddResource("sherwood", "exterior", "url:/foo/bar.map", customFacets)
	forest_map_01.Publish()

	// Now, let's create links on our forest to it's constituent trees
	err = forest_map_01.LinkTo("oak", oak02, customFacets)
	if err != nil {
		panic(err)
	}
	forest_map_01.LinkTo("elm", elm01)
	forest_map_01.LinkTo("pine", pine01)

	// We can link the other way too
	oak02.LinkTo("usedby", forest_map_01)
	elm01.LinkTo("usedby", forest_map_01)
	pine01.LinkTo("usedby", forest_map_01)

	fmt.Println("Done creating example objects.")
}
