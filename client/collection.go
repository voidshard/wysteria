package wysteria_client

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

// Wrapper class for wysteria/common Collection
type Collection struct {
	conn *Client
	data *wyc.Collection
}

// Return the name of this collection
func (c *Collection) Name() string {
	return c.data.Name
}

// Return the Id of this collection
func (c *Collection) Id() string {
	return c.data.Id
}

// Return the Id of this collection's parent (if any)
func (c *Collection) Parent() string {
	return c.data.Parent
}

// Get the facet value and a bool indicating if the value exists for the given key.
func (i *Collection) Facet(key string) (string, bool) {
	val, ok := i.data.Facets[key]
	return val, ok
}

// Get all facets
func (i *Collection) Facets() map[string]string {
	return i.data.Facets
}

// Set all the key:value pairs given on this Collection's facets.
func (i *Collection) SetFacets(in map[string]string) error {
	return i.conn.middleware.UpdateCollectionFacets(i.data.Id, in)
}

// Delete this collection.
// Warning: Any & all child Items and their children will be deleted too.
func (c *Collection) Delete() error {
	return c.conn.middleware.DeleteCollection(c.data.Id)
}

// Get all Item objects that name this collection as "parent"
func (c *Collection) Items(opts ...SearchParam) ([]*Item, error) {
	opts = append(opts, ChildOf(c.data.Id))
	return c.conn.Search(opts...).FindItems()
}

// Get all Collection objects that name this collection as "parent"
func (c *Collection) Collections(opts ...SearchParam) ([]*Collection, error) {
	opts = append(opts, ChildOf(c.data.Id))
	return c.conn.Search(opts...).FindCollections()
}

// Create a new Item with the given fields & return it.
//  - The item type & variant fields must be unique within a given collection.
//  - The reserved facet FacetCollection is set as a facet automatically.
func (c *Collection) CreateItem(itemtype, variant string, facets map[string]string) (*Item, error) {
	all_facets := map[string]string{}
	if facets != nil {
		for key, value := range facets {
			all_facets[key] = value
		}
	}

	all_facets[wyc.FacetCollection] = c.data.Name

	cmn_item := &wyc.Item{
		Parent:   c.data.Id,
		ItemType: itemtype,
		Variant:  variant,
		Facets:   all_facets,
	}

	item_id, err := c.conn.middleware.CreateItem(cmn_item)
	if err != nil {
		return nil, err
	}
	cmn_item.Id = item_id

	return &Item{
		conn: c.conn,
		data: cmn_item,
	}, nil
}

// Create a child collection of this collection
func (c *Collection) CreateCollection(name string, facets map[string]string) (*Collection, error) {
	if facets == nil {
		facets = map[string]string{}
	}
	facets[wyc.FacetCollection] = c.Name()
	return c.conn.createCollection(name, c.Id(), facets)
}

// Create a new collection with the given name & parent id (if any)
func (w *Client) createCollection(name, parent string, facets map[string]string) (*Collection, error) {
	col := &wyc.Collection{Id: "", Name: name, Parent: parent}
	collection_id, err := w.middleware.CreateCollection(col)
	if err != nil {
		return nil, err
	}

	col.Id = collection_id
	return &Collection{
		conn: w,
		data: col,
	}, nil
}

// Create a new collection & return it (that is, a collection with no parent)
//  - The collection name is required to be unique among all collections
func (w *Client) CreateCollection(name string, facets map[string]string) (*Collection, error) {
	return w.createCollection(name, "", facets)
}

// Collection is a helpful wrapper that looks for a single collection
// with either the name or Id of the given 'identifier' and returns it if found
func (w *Client) Collection(identifier string) (*Collection, error) {
	results, err := w.Search(Id(identifier)).Or(Name(identifier)).FindCollections()
	if err != nil {
		return nil, err
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return nil, errors.New(fmt.Sprintf("Expected 1 result, got %d", len(results)))
}
