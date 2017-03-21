# Wysteria

An open source asset tracking, versioning & publishing system written in Go. 


## Before You Start

There are three components that a wysteria server uses in conjunction on to work, they are:
- A database for long term storage
- A 'searchbase' for indexing and running queries against
- A middleware for ferrying data back and forth between server(s) and client(s)

Each of these components is behind an interface so it's easy to write implementations for your favourite solutions. Currently there are implementations for

Databases
 - MongoDb
 - BoltDB (embedded)

Search
 - ElasticSearch
 - Bleve (embedded)

Middleware
 - Gorpc (embedded)
 - Nats.io (either)


## Connecting
Once you're set up, opening a connection to the server is reasonably simple
```Go
import (       
  wyc "wysteria/wysteria_client"
)

func main() {
  client, err := wyc.New()
  if err != nil {       
    panic(err)
  }
  defer client.Close()
}
```
By default wysteria will check for an ini file in the current directory, then it'll poll the env variable WYSTERIA_CLIENT_INI to see if it can find a config file there. Failing that, it'll assume a set of default values.

## Collections
At the highest level are collections. Collections are straight forward enough - each has a unique name and can be created via

```Go
client.CreateCollection("spriteSets")
```

You can easily fetch a collection via either the name or id
```Go
collection, _ := client.GetCollection("spriteSets")
sameCollection, _ := client.GetCollection(collection.Id())
```
And all child items of a collection via
```Go
items, _ := collection.GetItems()
```

## Items
A collection can have any number of items, with the constraint that there is at most one item of each 'item type' and 'variant'.
The 'item type' and 'variant' are simply strings that are passed in when an item is created. 
```Go
item1, _ := collection.CreateItem("2dSprite", "alice", nil)
item2, _ := collection.CreateItem("2dSprite", "bob", nil)
item3, _ := collection.CreateItem("spriteSheet", "batman", nil)
```
They also have facets so that one can add custom searchable metadata to them for easy finding later. 
```Go
item1.SetFacets(map[string]string{"colour": "white", "publisher": "batman"})
```
Wysteria automatically attaches some facets to created items & versions to make your life easier, such as the name of the parent collection. In the case of versions you also get the item's item type and variant.

Part of the usefulness of items is their ability to be linked together, you might say for example
```Go
tilesets, _ := client.GetCollection("tilesets")
maps, _ := client.GetCollection("maps")

yew_tree_tiles, _ := tilesets.CreateItem("forest", "yew01", nil)
forest_scene, _ := maps.CreateItem("exterior", "sherwoodForest", nil)

forest_scene.LinkTo("input", yew_tree_tiles)
```
That is, the forest_scene item has a link called "input" that connects to our yew_tree_tiles item.

You can link any number of items in this way, and the names are not required to be unique. You can use these links to walk from one object to another 
```Go
items, _ := forest_scene.GetLinked()
input_items, _ := forest_scene.GetLinkedByName("input")
```

You can also traverse 'up' and 'down' via 
```Go
myitem.GetParent()            # fetch the parent of this (a collection)
myitem.GetPublished()         # fetch the published version for this
```

Which brings us to ..

## Versions
Wysteria is about asset tracking and versioning. So far we have a 'sherwoodForest' item which is an 'exterior' belonging to a collection of 'maps' which we might think of as an asset. 
Now we consider different iterations of the asset, which we call versions. Each item has any number of versions which are numbered automatically, starting at 1. Naturally, there is only one version of each number belonging to a given item.

```Go
version1, _ := forest_scene.CreateVersion(nil)
version2, _ := forest_scene.CreateVersion(nil)
```
Versions are linkable and carry facets, exactly like items. They also have a few other unique properties.

Firstly Versions are numbered starting from 1.
```Go
version1.Version() // returns 1 (an int32)
```

Secondly a Version may be marked as "published" via
```Go
version1.Publish()
```
each Item can have at most one child Version marked as Published.

Lastly they are also able to have Resources attached to them. Which brings us to ..

## Resources
Resources each have a name, type and location, all strings. Any number of them can be attached to a version.
```Go
myVersion.AddResource("floorA", "image", "/path/to/image.0001.png")
myVersion.AddResource("floorB", "image", "/path/to/image.0002.png")
myVersion.AddResource("statsFile", "xml", "/path/to/floor.3.xml")
myVersion.AddResource("batmanSettings", "url", "http://cdn.mystuff/batman.json")
myVersion.AddResource("somethingElse", "other", "/path/to/foo.bar")
```

As you might expect, you can easily retrieve a versions' resources in multiple ways
```Go
all_resources, err := ver.GetAllResources()
batman_image_resources, err := ver.GetResources("sprite", "image")
image_resouces, err = ver.GetResourcesByType("image")
batman_resources, err = ver.GetResourcesByName("sprite")
```

## Searching
At this point, you could start with a given collection, then walk up, down and/or sideways through the hierarchy to find what you're looking for. But more likely you're looking for an easier way. You can search on almost any field, facet, name, type, location or id, or any combination of the above.
A search is it's own object that keeps track of your parameters. You can keep tacking them on to refine the search as much as you want,
```Go
searchObj := client.Search()
searchObj.ItemType("forest").ItemVariant("yew01").HasFacets(map[string]string{"foo": "bar"})
```
When you're ready to actually get the results you can call
```
items, _ := search.Items()  # return all matching 'item' objects
```
Equally, you can request any of the other wysteria types we've mentioned above, 
```
versions, _ := search.Versions()         # return all matching 'version' objects
resources, _ := search.Resources()   # return all matching 'resource' objects
```
.. etc. Note that each of the terms added this way are joined as if by a logical "and". Also note that if the object you request doesn't have a specified search field then that field is ignored for the purposes of considering whether to return the given object. For example the 'location' field exists only on resources, but you could specify a location search parameter and then request all matching 'version' objects.

Now we have the ability to search via arbitrary "I'd like an item(s) with X and Y and Z" but what about an or statement? Wysteria supports this too and it's pretty similar to the above,
```Go
a_or_b_items, _ := client.Search().ItemType("a").Id("abc123").Or().ItemType("b").Items()
```
That is, this search will return all items from any collection(s) that have either
- id of "abc123" and item type of "a"
- item type of "b"


## Notes

- There will be a delay between when something is created in wysteria, and when it is returnable via a search. This delay should be measured in seconds at most, depending on the 'searchbase' being used. It's possible that different searchbase implementations will overcome this in the future
- "Moving" and/or "renaming" are foreign concepts to wysteria, if you need to move something you should delete and recreate it. Part of this is because of the complications it introduces, and part of this is to be able to support deterministic ids (in a later version). That is, most objects and fields are immutable after creation.
- Wysteria makes no attempt to understand the links you set or sanitize strings given. The meaning behind the resource 'location' for example is left entirely to the user -- it could be a filepath, a url, an id into another system or something else entirely.
- There is no maximum number of facets, links or resources that one can attach that is enforced by the system. Practical limits of the underlying systems will begin to apply at some point however.

## Contributing

All contributors are welcome. If you write a new implementation for any of the interfaces or make any improvements please do make a pull request. 

Also, if (when) you find bugs, let me know.


## ToDo List
- unittests for business logic
- logging & live statistics gathering functionality
- admin console 
  - extend the wysteria server chan (currently subscribed to but unused) to allow realtime management of live server(s)
    - allow / disallow certain client requests
    - introduce some kind of configurable auto load-balancing 
    - allow changing of configuration option(s) 
    - add alerts for certain events or server statuses
    - allow temporary rerouting of client requests
- implement system for deterministic ids (?)
- check if ffjson using the Marshalable interface is slowing it down - suspect it might be after reading the docs a bit more
