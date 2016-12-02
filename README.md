v0.9: consider working but alpha until addition of test suite
#Wysteria

An open source asset tracking & versioning system written in Go. 


##Before You Start

There are three extra components that a wysteria server depends on, these are
- A database for long term storage
- A 'searchbase' for indexing and running queries against
- A middleware for transporting data back and forth between client(s) and server(s)

Each of these is behind an interface so it's easy to write implementations for your favourite solutions. At the moment there are implementations for 
- MongoDB (database)
- ElasticSearch (searchbase)
- Nats.io (middleware)

Thus you'll need to have set up each of these services for wysteria to use.

##Connecting
Once you're set up, opening a connection to the server is reasonably simple
```Go
import (       
  wyc "wysteria/wysteria_client"
)

func main() {
  client, err := wyc.New().Connect()
  if err != nil {       
    panic(err)
  }
  defer client.Close()
}
```
By default wysteria will check for an ini file in the current directory, then it'll poll the env variable WYSTERIA_CLIENT_INI to see if it can find a config file there. Failing that, it'll assume a set of default values.

##Collections
At the highest level are collections. Collections are straight forward enough - each has a unique name and can be created via

```Go
client.CreateCollection("spriteSets")
```

You can easily fetch a collection via either the name or id
```Go
collection, _ := client.GetCollection("spriteSets")
sameCollection, _ := client.GetCollection(collection.Id())
```

##Items
A collection can have any number of items, with the constraint that there is at most one item of each 'item type' and 'variant'.
The 'item type' and 'variant' are simply strings that are passed in when an item is created. 
```Go
item1, _ := collection.CreateItem("2dSprite", "alice")
item2, _ := collection.CreateItem("2dSprite", "bob")
item3, _ := collection.CreateItem("spriteSheet", "batman")
```
They also have facets so that one can add custom searchable metadata to them for easy finding later. 
```Go
item1.SetFacet("colour", "white")
```
Wysteria automatically attaches some facets to created items & versions to make your life easier, such as the name of the parent collection.

Part of the usefulness of items is their ability to be linked together, you might say for example
```Go
tilesets, _ := client.GetCollection("tilesets")
maps, _ := client.GetCollection("maps")

yew_tree_tiles, _ := tilesets.CreateItem("forest", "yew01")
forest_scene, _ := maps.CreateItem("exterior", "sherwoodForest")

forest_scene.LinkTo("input", yew_tree_tiles)
```
That is, the forest_scene item has a link called "input" that connects to our yew_tree_tiles item.

You can link any number of items in this way, and the names are not required to be unique. You can use these links to walk from one object to another 
```Go
items, _ := forest_scene.GetLinkedItems() 
input_items, _ := forest_scene.GetLinkedItemsByName("input")
```

You can also traverse 'up' and 'down' via 
```Go
myitem.GetParent()            # fetch the parent of this (a collection)
myitem.GetHighestVersion()    # fetch the highest numbered version of this item
```

Which brings us to ..
##Versions
Wysteria is about asset tracking and versioning. So far we have a 'sherwoodForest' item which is an 'exterior' belonging to a collection of 'maps' which we might think of as an asset. 
Now we consider different iterations of the asset, which we call versions. Each item has any number of versions which are numbered automatically, starting at 1. Naturally, there is only one version of each number belonging to a given item.

```Go
version1, _ := forest_scene.CreateNextVersion()
version2, _ := forest_scene.CreateNextVersion()
```
Versions are linkable and carry facets, exactly like items.

```Go
tree_version.LinkTo("input", sherwood_version)
sherwood_version.LinkTo("output", tree_version)
```

##FileResources
File resources each have a name, type and location, all strings. Any number of them can be attached to a version. 
```Go
myVersion.AddResource("floorA", "image", "/path/to/image.0001.png")
ver.AddResource("floorB", "image", "/path/to/image.0002.png")
ver.AddResource("statsFile", "xml", "/path/to/floor.3.xml")
ver.AddResource("batmanSettings", "url", "http://cdn.mystuff/batman.json")
ver.AddResource("somethingElse", "other", "/path/to/foo.bar")
```

As you might expect, you can easily retrieve a versions' resources in multiple ways
```Go
all_resources, err := ver.GetAllResources()
image_resouces, err = ver.GetResourcesByType("image")
batmans_resources, err = ver.GetResourcesByName("batmanSettings")
```
Note that wysteria makes no attempt to understand the given data, so you're free to use ids to another database, a url, a filename or some other thing.

##Searching
At this point, you could start with a given collection, then walk up, down and/or sideways through the heirarchy to find what you're looking for. But more likely you're looking for an easier way. You can search on almost any field, facet, name, type, location or id, or any combonation of the above.
A search is it's own object that keeps track of your parameters. You can keep tacking them on to refine the search as much as you want,
```Go
searchObj := client.Search()
searchObj.ItemType("forest").ItemVariant("yew01").HasFacets(map[string]string{"foo": "bar"})
```
When you're ready to actually get the results you can call
```
items, _ := search.Items()
```
Equally, you can request any of the other wysteria types we've mentioned above, 
```
versions, _ := search.Versions()
resources, _ := search.FileResources()
```
.. etc. Note that each of the terms added this way are joined as if by a logical "and". Also note that if the object you requested doesn't have a specified search field (for example 'location' exists only on file resources) then that field is ignored for the purposes of considering whether to return the given object.

That works for arbitrary "I'd like an item with X and Y and Z" but what about an or statement? Well, it's pretty close
```Go
a_or_b_items, _ := client.Search().ItemType("a").Id("abc123").Or().ItemType("b").Items()
```
That is, this search will return all items from any collection(s) that have either
- id of "abc123" and item type of "a"
- item type of "b"


##Notes

- There will be a delay between when something is created in wysteria, and when it is returnable via a search. This delay should be measured in seconds at most, depending on the 'searchbase' being used. It's possible that different searchbase implementations will overcome this in the future
- "Moving" and/or "renaming" are foreign concepts to wysteria, if you need to move something you should delete and recreate it. Part of this is because of the complications it introduces, and part of this is to be able to support deterministic ids (in a later version)


##Contributing

All contributors are welcome. If you write a new implementation for any og the interfaces or make any improvements please do make a pull request. 

Also, if (when) you find bugs, let me know.


##ToDo List

- unittests for business logic
  - mock in-memory searchbase implementation
  - mock in-memory database implementation
- logging & live statistics gathering functionality
  - Open to suggestions on implementation
- admin console 
  - extend the wysteria server chan (currently subscribed to but unused) to allow realtime management of live server(s)
    - allow / disallow certain client requests
    - introduce some kind of configurable auto load-balancing 
    - allow changing of configuration option(s) (like the database we're reading / writing to etc)
    - add alerts for certain events or server statuses
    - allow temporary routing of all client requests to another server(s) (that we can shutdown gracefully)
- move encoding scheme to something more time/memory efficient

