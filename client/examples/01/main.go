package main

import (
	wysteria "github.com/voidshard/wysteria/client"
)

func main () {
	// Example01: Creating
	//  (We'll use the resources created here in the next examples)

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
	oak, err := col.CreateItem("tree", "oak")
	if err != nil {
		panic(err)
	}

	// Create the first version of our oak tree
	oak01, err := oak.CreateNextVersion()
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
	elm01, _ := elm.CreateNextVersion()
	elm01.AddResource("default", "png", "/path/to/elm01.png")
	elm01.AddResource("lowres", "jpeg", "/path/lowres/image.jpeg")

	pine, _ := col.CreateItem("tree", "pine")
	pine01, _ := pine.CreateNextVersion()
	pine01.AddResource("default", "png", "/path/to/pine01.png")
	pine01.AddResource("lowres", "jpeg", "/path/lowres/image.jpeg")
	pine01.AddResource("stats", "json", "url://file.json")

	// Shoot. Our oak tree looks awful. Let's redo it
	oak02, _ := oak.CreateNextVersion()
	oak02.AddResource("default", "png", "/other/images/oak02.png")

	// Ok sweet. Let's publish the things
	err = oak02.Publish()
	if err != nil {
		panic(err)
	}
	elm01.Publish()
	pine01.Publish()

	// One more oak (note that oak02 is still the "published" version -> important later)
	oak03, _ := oak.CreateNextVersion()
	oak03.AddResource("default", "png", "/other/images/oak03.png")

	// We'll create some links for later examples too, first, something to link to
	maps, _ := client.CreateCollection("maps")
	forest, _ := maps.CreateItem("map", "forest")
	forest_map_01, _ := forest.CreateNextVersion()

	// Now, let's create links to our forest
	err = oak02.LinkTo("oak", forest_map_01)
	if err != nil {
		panic(err)
	}
	elm01.LinkTo("elm", forest_map_01)
	pine01.LinkTo("pine", forest_map_01)
}
