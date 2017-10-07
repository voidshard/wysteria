package wysteria_client

import (
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
func (c *Collection) ParentId() string {
	return c.data.Parent
}

// Return the parent collection of this collection (if any)
func (c *Collection) Parent() (*Collection, error) {
	return c.conn.Collection(c.ParentId())
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
func (c *Collection) CreateItem(itemtype, variant string, opts ...CreateOption) (*Item, error) {
	cmn_item := &wyc.Item{
		Parent:   c.data.Id,
		ItemType: itemtype,
		Variant:  variant,
		Facets:   map[string]string{},
	}
	child := &Item{
		conn: c.conn,
		data: cmn_item,
	}

	for _, opt := range opts {
		opt(c, child)
	}

	cmn_item.Facets[wyc.FacetCollection] = c.data.Name

	item_id, err := c.conn.middleware.CreateItem(cmn_item)
	if err != nil {
		return nil, err
	}
	cmn_item.Id = item_id

	return child, nil
}

// Create a child collection of this collection
func (c *Collection) CreateCollection(name string, opts ...CreateOption) (*Collection, error) {
	return c.conn.createCollection(name, c, opts...)
}

// Create a new collection with the given name & parent id (if any)
func (w *Client) createCollection(name string, parent *Collection, opts ...CreateOption) (*Collection, error) {
	col := &wyc.Collection{Id: "", Name: name, Parent: "", Facets: map[string]string{}}
	child := &Collection{
		conn: w,
		data: col,
	}

	for _, opt := range opts {
		opt(parent, child)
	}

	if parent == nil { // Nb this will overwrite facets set by users that we're going to use -> this is intentional
		col.Facets[wyc.FacetCollection] = wyc.FacetRootCollection
	} else {
		col.Parent = parent.Id()
		col.Facets[wyc.FacetCollection] = parent.Name()
	}

	collection_id, err := w.middleware.CreateCollection(col)
	if err != nil {
		return nil, err
	}
	col.Id = collection_id
	return child, child.SetFacets(child.Facets())
}

// Create a new collection & return it (that is, a collection with no parent)
//  - The collection name is required to be unique among all collections
func (w *Client) CreateCollection(name string, opts ...CreateOption) (*Collection, error) {
	return w.createCollection(name,  nil, opts...)
}

// Set initial user defined facets
func (c *Collection) initUserFacets(in map[string]string) {
	c.data.Facets = in
}