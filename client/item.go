package wysteria_client

import (
	wyc "github.com/voidshard/wysteria/common"
)

type Item struct {
	conn     *wysteriaClient
	data     *wyc.Item
	fromLink *Link
}

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

func (i *Item) Link() *Link {
	return i.fromLink
}

func (i *Item) Type() string {
	return i.data.ItemType
}

func (i *Item) Delete() error {
	return i.conn.middleware.DeleteItem(i.data.Id)
}

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

func (i *Item) GetLinkedByName(name string) ([]*Item, error) {
	found, err := i.getLinkedItems(name)
	if err != nil {
		return nil, err
	}
	return found[name], nil
}

func (i *Item) GetLinked() (map[string][]*Item, error) {
	return i.getLinkedItems("")
}

func (i *Item) Variant() string {
	return i.data.Variant
}

func (i *Item) Facet(key string) (string, bool) {
	val, ok := i.data.Facets[key]
	return val, ok
}

func (i *Item) Id() string {
	return i.data.Id
}

func (i *Item) SetFacets(in map[string]string) error {
	return i.conn.middleware.UpdateItemFacets(i.data.Id, in)
}

func (i *Item) CreateVersion(facets map[string]string) (*Version, error) {
	all_facets := map[string]string{}
	if all_facets != nil {
		for key, value := range facets {
			all_facets[key] = value
		}
	}

	parentCol, ok := i.data.Facets["collection"]
	if ok {
		all_facets["collection"] = parentCol
	}
	all_facets["itemtype"] = i.data.ItemType
	all_facets["variant"] = i.data.Variant

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

func (i *Item) Parent() string {
	return i.data.Parent
}

func (i *Item) GetParent() (*Collection, error) {
	return i.conn.GetCollection(i.data.Parent)
}
