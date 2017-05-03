package wysteria_client

import (
	wyc "github.com/voidshard/wysteria/common"
)

// Wrapper around wysteria/common Item object
type Item struct {
	conn     *wysteriaClient
	data     *wyc.Item
}

// Get which ever Version object is considered the "published" version of this Item
//  Note that this may not be the latest, or even be set, even if child versions exist.
func (i *Item) GetPublished() (*Version, error) {
	ver, err := i.conn.middleware.GetPublishedVersion(i.data.Id)
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
func (i *Item) getLinkedItems(name string) (map[string][]*Item, error) {
	links, err := i.conn.middleware.FindLinks(
		[]*wyc.QueryDesc{
			{LinkSrc: i.data.Id, Name: name},
		},
	)
	if err != nil {
		return nil, err
	}

	item_id_to_link := map[string]*wyc.Link{}
	ids := []*wyc.QueryDesc{}
	for _, link := range links {
		id := link.Src

		if link.Src == i.data.Id {
			id = link.Dst
		}

		item_id_to_link[id] = link
		ids = append(ids, &wyc.QueryDesc{Id: id})
	}

	items, err := i.conn.middleware.FindItems(ids)
	if err != nil {
		return nil, err
	}

	result := map[string][]*Item{}
	for _, ver := range items {
		lnk, ok := item_id_to_link[ver.Id]
		if !ok {
			continue
		}

		result_list, ok := result[lnk.Name]
		if result_list == nil {
			result_list = []*Item{}
		}

		wrapped_item := &Item{
			conn: i.conn,
			data: ver,
		}

		result_list = append(result_list, wrapped_item)
		result[lnk.Name] = result_list
	}
	return result, nil
}

// Get all linked items (items where links exist that mention this as the source and them as the destination)
// where the link name is the given 'name'.
func (i *Item) GetLinkedByName(name string) ([]*Item, error) {
	found, err := i.getLinkedItems(name)
	if err != nil {
		return nil, err
	}
	return found[name], nil
}

// Find and return all linked items for which links exist that name this as the source.
// That is, this first finds all links for which the source Id is this Item's Id, then
// gets all matching Items.
// Since this would cause us to lose the link 'name' we return a map of link name -> []*Item
func (i *Item) GetLinked() (map[string][]*Item, error) {
	return i.getLinkedItems("")
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

	parentCol, ok := i.data.Facets[FacetCollection]
	if ok {
		all_facets[FacetCollection] = parentCol
	}
	all_facets[FacetItemType] = i.data.ItemType
	all_facets[FacetItemVariant] = i.data.Variant

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
func (i *Item) Parent() string {
	return i.data.Parent
}

// Lookup and return the parent Collection of this Item
func (i *Item) GetParent() (*Collection, error) {
	return i.conn.GetCollection(i.data.Parent)
}
