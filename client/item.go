package wysteria_client

import (
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
func (i *Item) LinkTo(name string, other *Item) error {
	if i.Id() == other.Id() { // Prevent linking to oneself
		return nil
	}

	lnk := &wyc.Link{
		Name: name,
		Src:  i.data.Id,
		Dst:  other.data.Id,
	}
	_, err := i.conn.middleware.CreateLink(lnk)
	return err
}

// Find and return all linked items for which links exist that name this as the source.
// That is, this first finds all links for which the source Id is this Item's Id, then
// gets all matching Items.
// Since this would cause us to lose the link 'name' we return a map of link name -> []*Item
func (i *Item) Linked(opts ...SearchParam) (map[string][]*Item, error) {
	opts = append(opts, ChildOf(i.Id()))
	links, err := i.conn.Search(opts...).FindLinks()
	if err != nil {
		return nil, err
	}

	item_id_to_link := map[string]*Link{}
	ids := []*wyc.QueryDesc{}
	for _, link := range links {
		id := link.SourceId()

		if link.SourceId() == i.data.Id {
			id = link.DestinationId()
		}

		item_id_to_link[id] = link
		ids = append(ids, &wyc.QueryDesc{Id: id})
	}

	// Apply limit to query just to be safe
	items, err := i.conn.middleware.FindItems(int32(len(ids)), 0, ids)
	if err != nil {
		return nil, err
	}

	result := map[string][]*Item{}
	for _, ver := range items {
		lnk, ok := item_id_to_link[ver.Id]
		if !ok {
			continue
		}

		result_list, ok := result[lnk.Name()]
		if result_list == nil {
			result_list = []*Item{}
		}

		wrapped_item := &Item{
			conn: i.conn,
			data: ver,
		}

		result_list = append(result_list, wrapped_item)
		result[lnk.Name()] = result_list
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

// Get the Id for this Item
func (i *Item) Id() string {
	return i.data.Id
}

// Set all the key:value pairs given on this Item's facets.
// Note that the server will ignore the setting of reserved facets.
func (i *Item) SetFacets(in map[string]string) error {
	return i.conn.middleware.UpdateItemFacets(i.data.Id, in)
}

// Create the next child version of this item, with the given facets (if any).
// That is,
//  - create a version whose parent is this item's id
//  - set the reserved Facets for Collection, ItemType and ItemVariant accordingly
//  - the server will allocate us a Version number
func (i *Item) CreateVersion(facets map[string]string) (*Version, error) {
	all_facets := map[string]string{}
	if all_facets != nil {
		for key, value := range facets {
			all_facets[key] = value
		}
	}

	parentCol, ok := i.data.Facets[wyc.FacetCollection]
	if ok {
		all_facets[wyc.FacetCollection] = parentCol
	}
	all_facets[wyc.FacetItemType] = i.data.ItemType
	all_facets[wyc.FacetItemVariant] = i.data.Variant

	ver := &wyc.Version{
		Parent: i.data.Id,
		Facets: all_facets,
	}

	version_id, version_num, err := i.conn.middleware.CreateVersion(ver)
	if err != nil {
		return nil, err
	}
	ver.Id = version_id
	ver.Number = version_num
	return &Version{
		data: ver,
		conn: i.conn,
	}, nil
}

// Return the Id of the parent of this Item
func (i *Item) ParentId() string {
	return i.data.Parent
}

// Lookup and return the parent Collection of this Item
func (i *Item) Parent() (*Collection, error) {
	return i.conn.Collection(i.data.Parent)
}
