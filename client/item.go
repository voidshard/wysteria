package wysteria_client

import (
	wyc "github.com/voidshard/wysteria/common"
)

type item struct {
	conn *wysteriaClient
	data *wyc.Item
	fromLink *link
}

func (i *item) GetPublishedVersion() (*version, error) {
	ver, err := i.conn.middleware.GetPublishedVersion(i.data.Id)
	if err != nil {
		return nil, err
	}
	return &version{
		conn: i.conn,
		data: ver,
	}, nil
}

func (i *item) Link() *link {
	return i.fromLink
}

func (i *item) Type() string {
	return i.data.ItemType
}

func (i *item) Delete() error {
	return i.conn.middleware.DeleteItem(i.data.Id)
}

func (i *item) LinkTo(name string, other *item) error {
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

func (i *item) getLinkedItems(name string) ([]*item, error) {
	links, err := i.conn.Search().Src(i.data.Id).Name(name).Links()
	if err != nil {
		return nil, err
	}

	item_id_to_link := map[string]*link{}
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

	result := []*item{}
	for _, itm := range items {
		lnk, ok := item_id_to_link[itm.Id()]
		if ok {
			itm.fromLink = lnk
		}
	}
	return result, nil
}

func (i *item) GetLinkedItemsByName(name string) ([]*item, error) {
	return i.getLinkedItems(name)
}

func (i *item) GetLinkedItems() ([]*item, error) {
	return i.getLinkedItems("")
}

func (i *item) Variant() string {
	return i.data.Variant
}

func (i *item) Facet(key string) (string, bool) {
	val, ok := i.data.Facets[key]
	return val, ok
}

func (i *item) Id() string {
	return i.data.Id
}

func (i *item) SetFacets(in map[string]string) error {
	return i.conn.middleware.UpdateItemFacets(i.data.Id, in)
}

func (i *item) CreateNextVersion() (*version, error) {
	facets := map[string]string{}
	parentCol, ok := i.data.Facets["collection"]
	if ok {
		facets["collection"] = parentCol
	}
	facets["itemtype"] = i.data.ItemType
	facets["variant"] = i.data.Variant

	ver := &wyc.Version{
		Parent: i.data.Id,
		Facets: facets,
	}

	version_id, version_num, err := i.conn.middleware.CreateVersion(ver)
	if err != nil {
		return nil, err
	}
	ver.Id = version_id
	ver.Number = version_num
	return &version{
		data: ver,
		conn: i.conn,
	}, nil
}

func (i *item) Parent() string {
	return i.data.Parent
}

func (i *item) GetParent() (*collection, error) {
	return i.conn.GetCollection(i.data.Parent)
}
