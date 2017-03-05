package wysteria_client

import (
	wyc "github.com/voidshard/wysteria/common"
)

type item struct {
	conn *wysteriaClient
	data *wyc.Item
	fromLink *link
}

func (i *item) Link() *link {
	return i.fromLink
}

func (i *item) Type() string {
	return i.data.ItemType
}

func (c *item) Delete() error {
	return c.conn.middleware.DeleteItem(c.data.Id)
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

	result := []*item{}
	for _, itm := range items {
		wrapped_item := &item{
			conn: i.conn,
			data: itm,
		}

		lnk, ok := item_id_to_link[itm.Id]
		if ok {
			wrapped_item.fromLink = &link{conn: i.conn, data: lnk}
		}
		result = append(result, wrapped_item)
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
