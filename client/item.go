package wysteria_client

import (
	"errors"
	wyc "github.com/voidshard/wysteria/common"
)

// Wrapper around wysteria/common Item object
type Item struct {
	conn *Client
	data *wyc.Item
}

// Get which ever Version object is considered the "published" version of this Item
//  Note that this may not be the latest, or even be set, even if child versions exist.
func (i *Item) PublishedVersion() (*Version, error) {
	ver, err := i.conn.middleware.PublishedVersion(i.data.Id)
	if err != nil {
		return nil, err
	}
	return &Version{
		conn: i.conn,
		data: ver,
	}, nil
}

// Return the ItemType of this item
func (i *Item) Type() string {
	return i.data.ItemType
}

// Delete this Item. Note all child Versions & their children will be deleted too.
func (i *Item) Delete() error {
	return i.conn.middleware.DeleteItem(i.data.Id)
}

// Link this item with a link described by 'name' to some other item.
func (i *Item) LinkTo(name string, other *Item, opts ...CreateOption) (*Link, error) {
	if i.Id() == other.Id() { // Prevent linking to oneself
		return nil, errors.New("link src and dst IDs cannot be equal")
	}

	lnk := &wyc.Link{
		Name:   name,
		Src:    i.data.Id,
		Dst:    other.data.Id,
		Facets: map[string]string{},
	}
	child := &Link{conn: i.conn, data: lnk}

	for _, opt := range opts {
		opt(i, child)
	}
	lnk.Facets[FacetLinkType] = FacetItemLink

	id, err := i.conn.middleware.CreateLink(lnk)
	lnk.Id = id
	return child, err
}

// Find and return all linked items for which links exist that name this as the source.
// That is, this first finds all links for which the source Id is this Item's Id, then
// gets all matching Items.
// Since this would cause us to lose the link 'name' we return a map of link name -> []*Item
func (i *Item) Linked(opts ...SearchParam) (map[string][]*Item, error) {
	opts = append(opts, LinkSource(i.Id()))
	links, err := i.conn.Search(opts...).FindLinks()
	if err != nil {
		return nil, err
	}

	itemIdToLinks := map[string]*Link{}
	ids := []*wyc.QueryDesc{}
	for _, link := range links {
		id := link.SourceId()

		if link.SourceId() == i.data.Id {
			id = link.DestinationId()
		}

		itemIdToLinks[id] = link
		ids = append(ids, &wyc.QueryDesc{Id: id})
	}

	// Apply limit to query just to be safe
	items, err := i.conn.middleware.FindItems(int32(len(ids)), 0, ids)
	if err != nil {
		return nil, err
	}

	result := map[string][]*Item{}
	for _, ver := range items {
		lnk, ok := itemIdToLinks[ver.Id]
		if !ok {
			continue
		}

		resultLists, ok := result[lnk.Name()]
		if resultLists == nil {
			resultLists = []*Item{}
		}

		wrappedItem := &Item{
			conn: i.conn,
			data: ver,
		}

		resultLists = append(resultLists, wrappedItem)
		result[lnk.Name()] = resultLists
	}
	return result, nil
}

// Return the variant string associated with this Item
func (i *Item) Variant() string {
	return i.data.Variant
}

// Get the facet value and a bool indicating if the value exists for the given key.
func (i *Item) Facet(key string) (string, bool) {
	val, ok := i.data.Facets[key]
	return val, ok
}

// Get all facets
func (i *Item) Facets() map[string]string {
	return i.data.Facets
}

// Get the Id for this Item
func (i *Item) Id() string {
	return i.data.Id
}

// Get the URI for this Item
func (i *Item) Uri() string {
	if i.data.Uri == "" {
		results, err := i.conn.Search(Id(i.Id())).FindItems(Limit(1))
		if len(results) == 1 && err == nil {
			i.data.Uri = results[0].Uri()
		}
	}
	return i.data.Uri
}

// Set all the key:value pairs given on this Item's facets.
// Note that the server will ignore the setting of reserved facets.
func (i *Item) SetFacets(in map[string]string) error {
	if in == nil {
		return nil
	}
	for k, v := range in {
		i.data.Facets[k] = v
	}
	return i.conn.middleware.UpdateItemFacets(i.data.Id, in)
}

// Create the next child version of this item, with the given facets (if any).
// That is,
//  - create a version whose parent is this item's id
//  - set the reserved Facets for Collection, ItemType and ItemVariant accordingly
//  - the server will allocate us a Version number
func (i *Item) CreateVersion(opts ...CreateOption) (*Version, error) {
	ver := &wyc.Version{
		Parent: i.data.Id,
		Facets: map[string]string{},
	}
	child := &Version{
		data: ver,
		conn: i.conn,
	}

	for _, opt := range opts {
		opt(i, child)
	}

	parentCol, ok := i.data.Facets[wyc.FacetCollection]
	if ok {
		ver.Facets[wyc.FacetCollection] = parentCol
	}
	ver.Facets[wyc.FacetItemType] = i.data.ItemType
	ver.Facets[wyc.FacetItemVariant] = i.data.Variant

	version_id, version_num, err := i.conn.middleware.CreateVersion(ver)
	if err != nil {
		return nil, err
	}

	ver.Id = version_id
	ver.Number = version_num
	return child, nil
}

// Return the Id of the parent of this Item
func (i *Item) ParentId() string {
	return i.data.Parent
}

// Lookup and return the parent Collection of this Item
func (i *Item) Parent() (*Collection, error) {
	return i.conn.Collection(i.data.Parent)
}

// Set initial user defined facets
func (i *Item) initUserFacets(in map[string]string) {
	i.data.Facets = in
}
