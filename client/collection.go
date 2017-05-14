package wysteria_client

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

// Wrapper class for wysteria/common Collection
type Collection struct {
	conn *wysteriaClient
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

// Create a new collection & return it
//  - The collection name is required to be unique among collections
func (w *wysteriaClient) CreateCollection(name string) (*Collection, error) {
	collection_id, err := w.middleware.CreateCollection(name)
	if err != nil {
		return nil, err
	}

	return &Collection{
		conn: w,
		data: &wyc.Collection{
			Id:   collection_id,
			Name: name,
		},
	}, nil
}

// Collection is a helpful wrapper that looks for a single collection
// with either the name or Id of the given 'identifier' and returns it if found
func (w *wysteriaClient) Collection(identifier string) (*Collection, error) {
	results, err := w.Search(Id(identifier)).Or(Name(identifier)).FindCollections()
	if err != nil {
		return nil, err
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return nil, errors.New(fmt.Sprintf("Expected 1 result, got %d", len(results)))
}
