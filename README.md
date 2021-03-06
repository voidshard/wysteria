# Wysteria

An open source asset tracking, versioning & publishing system written in Go. 

* [Before You Start](#before-you-start)
* [Building and Requirements](#building-and-requirements)
* [Available Clients](#clients)
* [Go Client Example](#connecting)
* [Types](#collections)
* [Notes](#notes)
* [Contributing](#contributing)
* [Todo List](#todo-list)

## Before You Start

There are three components that a wysteria server uses in conjunction to work, they are:
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

You can also start wysteria with no config at all in which case it will spin up it's own embedded Nats server
and use BoltDB & Bleve to write to an OS temporary folder.

## Building and Requirements

Project can either be cloned or use `go get github.com/voidshard/wysteria/server`

Dependencies are managed via [glide](http://glide.sh/)

Build the server (requires [Go runtime](https://golang.org/dl/) to be installed):

```
cd wysteria
go build -o wysteria ./server
```


## Clients

* Go: https://godoc.org/github.com/voidshard/wysteria/client
* Python: https://github.com/voidshard/pywysteria

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
collection, _ := client.Collection("spriteSets")
sameCollection, _ := client.Collection(collection.Id())
```
And all child items of a collection via
```Go
items, _ := collection.Items()
```

You may also create child collections to help you organise your data better like so
```Go
foo, _ := client.CreateCollection("foo")
mapsOfFoo, _ := foo.CreateCollection("maps")
```
And fetch children of a collection.
```Go
children, _ := foo.Collections()
```
All collections of a given parent are still required to have unique names.

## Items
A collection can have any number of items, with the constraint that there is at most one item of each 'item type' and 'variant'.
The 'item type' and 'variant' are simply strings that are passed in when an item is created. 
```Go
item1, _ := collection.CreateItem("2dSprite", "alice")
item2, _ := collection.CreateItem("2dSprite", "bob")
item3, _ := collection.CreateItem("spriteSheet", "batman")
```
They also have facets so that one can add custom searchable metadata to them for easy finding later.
```Go
item1.SetFacets(map[string]string{"colour": "white", "publisher": "batman"})
```
Wysteria automatically attaches some facets to created items & versions to make your life easier, such as the name of the parent collection. In the case of versions you also get the item's item type and variant.

Part of the usefulness of items is their ability to be linked together, you might say for example
```Go
tilesets, _ := client.Collection("tilesets")
maps, _ := client.Collection("maps")

yew_tree_tiles, _ := tilesets.CreateItem("forest", "yew01")
forest_scene, _ := maps.CreateItem("exterior", "sherwoodForest")

forest_scene.LinkTo("input", yew_tree_tiles)
```
That is, the forest_scene item has a link called "input" that connects to our yew_tree_tiles item.

You can link any number of items in this way, and the names are not required to be unique. You can use these links to walk from one object to another
```Go
items, _ := forest_scene.Linked()
items, _ := forest_scene.Linked(wysteria.Name("input"))
```

You can also traverse 'up' and 'down' via 
```Go
myitem.Parent()            # fetch the parent of this (a collection)
myitem.Published()         # fetch the published version for this
```

Which brings us to ..

## Versions
Wysteria is about asset tracking and versioning. So far we have a 'sherwoodForest' item which is an 'exterior' belonging to a collection of 'maps' which we might think of as an asset. 
Now we consider different iterations of the asset, which we call versions. Each item has any number of versions which are numbered automatically, starting at 1. Naturally, there is only one version of each number belonging to a given item.

```Go
version1, _ := forest_scene.CreateVersion()
version2, _ := forest_scene.CreateVersion()
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
```
Like with all wysteria structs, you can add your own facets at creation time too if you want to
```Go
customFacet := wysteria.Facets(map[string]string{"creator": "batman"})
myVersion.AddResource("batmanSettings", "url", "http://cdn.mystuff/batman.json", customFacet)
myVersion.AddResource("somethingElse", "other", "/path/to/foo.bar", customFacet)
```

As you might expect, you can easily retrieve a versions' resources, with optional extra parameters
```Go
all_resources, err := ver.Resources()
batman_image_resources, err := ver.Resources(wysteria.Name("sprite"), wysteria.ResourceType("image"))
```

## Searching
At this point, you could start with a given collection, then walk up, down and/or sideways through the hierarchy to find what you're looking for. But more likely you're looking for an easier way. You can search on almost any field, facet, name, type, location or id, or any combination of the above.
A search is it's own object that keeps track of your parameters. You can keep tacking them on to refine the search as much as you want,
```Go
search := client.Search(
    wysteria.ItemType("forest"),
    wysteria.ItemVariant("yew01"),
    wysteria.HasFacets(map[string]string{"foo": "bar"}),
)
```
When you're ready to actually get the results you can call
```
items, _ := search.FindItems()  # return all matching 'item' objects
```
Equally, you can request any of the other wysteria types we've mentioned above, 
```
versions, _ := search.FindVersions()         # return all matching 'version' objects
resources, _ := search.FindResources()   # return all matching 'resource' objects
```
.. etc. Note that each of the terms added this way are joined as if by a logical "and". Also note that if the object you request doesn't have a specified search field then that field is ignored for the purposes of considering whether to return the given object. For example the 'location' field exists only on resources, but you could specify a location search parameter and then request all matching 'version' objects.

Now we have the ability to search via arbitrary "I'd like an item(s) with X and Y and Z" but what about an or statement? Wysteria supports this too and it's pretty similar to the above,
```Go
a_or_b_items, _ := client.Search(
   wysteria.ItemType("a"),
   wysteria.Id("abc123"),
).Or(
    wysteria.ItemType("b"),
).FindItems()
```
That is, this search will return all items from any collection(s) that have either
- id of "abc123" and item type of "a"
- item type of "b"


## Notes

- Unless you are searching by Id, there will be a delay between when something is created in wysteria, and when it is returnable via a search. This delay should be measured in seconds at most, depending on the 'searchbase' being used.
    - There is a searchbase config 'ReindexOnWrite' to mitigate this but this is expensive to do: Using this should be considered very carefully.
- "Moving" and/or "renaming" are foreign concepts to wysteria, if you need to move something you should delete and recreate it (or enforce some kind of naming scheme so you don't screw up names..). That is, most objects and fields are immutable after creation.
- Wysteria makes no attempt to understand the links you set or sanitize strings given. The meaning behind the resource 'location' for example is left entirely to the user -- it could be a filepath, a url, an id into another system or something else entirely.
- There is no maximum number of facets, links or resources that one can attach that is enforced by the system. Practical limits of the underlying systems will begin to apply at some point however.

## Contributing

All contributors are welcome. If you write a new implementation for any of the interfaces or make any improvements please do make a pull request. 

Also, if (when) you find bugs, let me know.


## ToDo List
- admin console 
  - extend the wysteria server chan to allow realtime management of live server(s)
    - allow / disallow certain client requests
    - introduce some kind of configurable auto load-balancing 
    - allow changing of configuration option(s) 
    - add alerts for certain events or server statuses
    - allow temporary rerouting of client requests
- viewer

## v1.1
- Elasticsearch v2 support dropped from master
- Elasticsearch v6 (latest) support added
- Searching for an object(s) by Id(s) bypasses the searchbase, meaning ID searches are faster & have no delay.
- Created IDs are now deterministic, and are the same regardless of the database backend used.
- monitor that writes to stdout now writes time taken in call in millis to be more human readable.
- Docker & docker compose support for
    - local builds (compiled using the local wysteria src)
    - master builds (compiled by cloning github repo)
    - multi container (nats-elastic-mongo) & all in one (gRPC-bleve-bolt) builds
- Integration test suite added using all of the above docker builds
- gRPC fixes & timeout improvements
- SSL support added to gRPC middleware
- added some sugar to set a 'FacetLinkType' on a Link obj when created - it's either FacetVersionLink or FacetItemLink
- clientside now sets facets on local objects on a SetFacets() call(s)
- Removed random extra SetFacets() call when a collection is created
- Fixed a few misc nil pointer fixes that could occur on the server
- Removal of cascading deletes -> only links & resources are deleted when their parent goes
    - The server should now block deletes if it has children (links & resources do not apply to this)
- Delete calls no longer spawn extra routines
- Renamed lots of funcs & vars so they fit the Go style
- more doc strings
- links are now returned when LinkTo is called. The return signature of item and version .LinkTo is now (*Link, error)
- resources are now returned when AddResource is called. Return signature of version.AddResource is (*Resource, error)
- increased monitoring. There is now monitoring around middleware (enter / exit), database and searchbase calls
- resource uniqueness is enforced for a given location, name, type & parent
- Added config var 'ReindexOnWrite' that ensures that written data is immediately searchable. This is useful in tests
  where we want to immediately fetch things and the performance hit isn't considered a problem.
  Setting this to true for any prod system should be considered carefully as it's expected to be very costly.
