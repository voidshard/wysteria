/*
Common structs

Offers common structs used throughout wysteria as the common medium. These are the structs we pass to the
server proper, the database & searchbase modules and transcode to and from when sending over the middleware.

Middleware implementations may turn these into other objects, append more fields or otherwise encode them for
transport over the wire, but they are then required to then change them back into these standard formats for passing
to clients.
*/

package wysteria_common

// A collection is the highest level of object in wysteria.
// Each collection has a unique name and is used mostly to form logical groupings
// and help divide the search space for items into hopefully even-ish chunks.
type Collection struct {
	Parent string            `json:"Parent"`
	Name   string            `json:"Name"`
	Id     string            `json:"Id"`
	Uri    string            `json:"Uri"`
	Facets map[string]string `json:"Facets"`
}

// Items are the second tier of object in wysteria. Each has a parent collection
// denoted by the 'parent' field (the Id of a Collection).
// An 'Item' represents an abstract resource of a specific variant.
// For example different varieties of trees may all have type 'tree' and variants of 'oak' 'pine' etc.
// An Item doesn't refer to any particular version of something, rather it's the concept OF the thing
// independent of the specifics.
type Item struct {
	Parent   string            `json:"Parent"`
	Id       string            `json:"Id"`
	Uri      string            `json:"Uri"`
	ItemType string            `json:"ItemType"`
	Variant  string            `json:"Variant"`
	Facets   map[string]string `json:"Facets"`
}

// Versions represent a specific version or iteration of an Item.
// That is, assuming you were designing a model for an oak tree (an Item) your first model
// would be attached to Version #1 of item Oak Tree.
// If you then made a better model, that would constitute Version #2 of item Oak Tree.
type Version struct {
	Parent string            `json:"Parent"`
	Id     string            `json:"Id"`
	Uri    string            `json:"Uri"`
	Number int32             `json:"Number"`
	Facets map[string]string `json:"Facets"`
}

// A Resource is a path to specific named URI with some type.
// The name string is intended to convey what the resource is or represents.
// The resource type string is intended to convey how the resource should be understood or used.
// Location is the actual URI to the resource
// For example, an image resource might have name:thumbnail type:url location:www.foo.com/bar.jpg
type Resource struct {
	Parent       string            `json:"Parent"`
	Name         string            `json:"Name"`
	ResourceType string            `json:"ResourceType"`
	Id           string            `json:"Id"`
	Uri          string            `json:"Uri"`
	Location     string            `json:"Location"`
	Facets       map[string]string `json:"Facets"`
}

// A Link is an abstract named link "the thing of Id Src relates to the thing of Id Dst"
// The name is intended to convey the nature of the relationship.
type Link struct {
	Name   string            `json:"Name"`
	Id     string            `json:"Id"`
	Uri    string            `json:"Uri"`
	Src    string            `json:"Src"`
	Dst    string            `json:"Dst"`
	Facets map[string]string `json:"Facets"`
}

// A QueryDesc is a generic way to describe what one is searching for.
// Each field of a given QueryDesc is understood as an AND.
// Multiple QueryDesc's together are understood as an OR.
// That is, given a list of QueryDesc objects, results should be returned that match ALL of the fields
// (relevant to what you're search for) from at least one of the QueryDesc objects.
type QueryDesc struct {
	Parent        string
	Id            string
	Uri           string
	VersionNumber int32
	ItemType      string
	Variant       string
	Facets        map[string]string
	Name          string
	ResourceType  string
	Location      string
	LinkSrc       string
	LinkDst       string
}

// Interface for turning the struct into some byte representation.
// By default we're using ffjson to encode & later decode these but middleware needn't use this if
// they've got different requirements.
type Marshalable interface {
	MarshalJSON() ([]byte, error)
}
