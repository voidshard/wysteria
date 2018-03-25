package main

import (
	"fmt"
	wysteria "github.com/voidshard/wysteria/client"
)

func main() {
	// Example 08: Resources

	client, err := wysteria.New()
	if err != nil {
		panic(err)
	}

	// Let's get our published pine
	items, _ := client.Search(wysteria.ItemType("tree"), wysteria.ItemVariant("pine")).FindItems()
	publishedVersion, _ := items[0].PublishedVersion()

	// We can grab resources by Name
	defaultResources, err := publishedVersion.Resources(wysteria.Name("default"))
	if err != nil {
		panic(err)
	}
	fmt.Println("--default resources--")
	for _, resource := range defaultResources {
		fmt.Println(resource.Name(), resource.Type(), resource.Location())
	}
	//--default resources--
	//default png /path/to/pine01.png

	// by Type
	jsonResources, _ := publishedVersion.Resources(wysteria.ResourceType("json"))
	fmt.Println("--json resources--")
	for _, resource := range jsonResources {
		fmt.Println(resource.Name(), resource.Type(), resource.Location())
	}
	//--json resources--
	//stats json url://file.json

	// some combination of the two
	lowresJpgs, _ := publishedVersion.Resources(wysteria.Name("lowres"), wysteria.ResourceType("jpeg"))
	fmt.Println("--lowres jpeg resources--")
	for _, resource := range lowresJpgs {
		fmt.Println(resource.Name(), resource.Type(), resource.Location())
	}
	//--lowres jpeg resources--
	//lowres jpeg /path/lowres/image.jpeg

	// Or grab all of them
	allResources, _ := publishedVersion.Resources()
	fmt.Println("--resources--")
	for _, resource := range allResources {
		fmt.Println(resource.Name(), resource.Type(), resource.Location())
	}
	//--resources--
	//default png /path/to/pine01.png
	//lowres jpeg /path/lowres/image.jpeg
	//stats json url://file.json
}
