package wysteria_client

import (
	wyc "github.com/voidshard/wysteria/common"
)

type Item struct {
	conn *wysteriaClient
	data *wyc.Item
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

func (i *Item) getLinkedItems(name string) ([]*Item, error) {
	links, err := i.conn.Search().Src(i.data.Id).Name(name).Links()
	if err != nil {
		return nil, err
	}

	item_id_to_link := map[string]*Link{}
	search := i.conn.Search()
	for _, link := range links {
		id := link.SourceId()

		if link.SourceId() == i.data.Id {
			id = link.DestinationId()
		}

		item_id_to_link[id] = link
		search.Id(id).Or()
	}

	items, err := search.Items()
	if err != nil {
		return nil, err
	}

	result := []*Item{}
	for _, itm := range items {
		lnk, ok := item_id_to_link[itm.Id()]
		if ok {
			itm.fromLink = lnk
		}
	}
	return result, nil
}

func (i *Item) GetLinkedByName(name string) ([]*Item, error) {
	return i.getLinkedItems(name)
}

func (i *Item) GetLinked() ([]*Item, error) {
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
