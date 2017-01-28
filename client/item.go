package wysteria_client

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

type item struct {
	conn *wysteriaClient
	data wyc.Item
}

func (i *item) Type() string {
	return i.data.ItemType
}

func (c *item) Delete() error {
	return c.conn.requestData(wyc.MSG_DELETE_ITEM, &c.data, nil)
}

func (i *item) LinkTo(name string, other *item) error {
	if i.Id() == other.Id() { // Prevent linking to oneself
		return nil
	}
	for _, lid := range i.data.Links { // Prevent duplicate links
		if lid == other.Id() {
			return nil
		}
	}

	lnk := wyc.Link{
		Name: name,
		Src:  i.data.Id,
		Dst:  other.data.Id,
	}
	err := i.conn.requestData(wyc.MSG_CREATE_LINK, lnk, &lnk)
	if err != nil {
		return err
	}
	i.data.Links = append(i.data.Links, lnk.Id)
	return i.update()
}

func (i *item) getLinked(name string) ([]*item, map[string]string, error) {
	query := []wyc.QueryDesc{}
	for _, lnk_id := range i.data.Links {
		q := wyc.QueryDesc{Id: lnk_id, LinkSrc: i.data.Id, Name: name}
		query = append(query, q)
	}

	lnks := []wyc.Link{}
	err := i.conn.requestData(wyc.MSG_FIND_LINK, &query, &lnks)
	if err != nil {
		return nil, nil, err
	}

	lnk_name_map := map[string]string{}
	iquery := []wyc.QueryDesc{}
	for _, lnk := range lnks {
		lnk_name_map[lnk.Dst] = lnk.Name
		iquery = append(iquery, wyc.QueryDesc{Id: lnk.Dst})
	}

	data := []wyc.Item{}
	err = i.conn.requestData(wyc.MSG_FIND_ITEM, &iquery, &data)
	if err != nil {
		return nil, nil, err
	}

	items := []*item{}
	for _, itemdata := range data {
		items = append(items, &item{
			conn: i.conn,
			data: itemdata,
		})
	}
	return items, lnk_name_map, nil
}

func (i *item) GetLinkedItemsByName(name string) ([]*item, error) {
	items, _, err := i.getLinked(name)
	return items, err
}

func (i *item) GetLinkedItems() (map[string][]*item, error) {
	items, name_map, err := i.getLinked("")
	if err != nil {
		return nil, err
	}

	results := map[string][]*item{}
	for _, itm := range items {
		name, ok := name_map[itm.Id()]
		if !ok {
			continue
		}

		ls, ok := results[name]
		if !ok {
			ls = []*item{itm}
		} else {
			ls = append(ls, itm)
		}
		results[name] = ls
	}
	return results, nil
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

func (i *item) SetFacet(key, value string) error {
	i.data.Facets[key] = value
	return i.update()
}

func (i *item) update() error {
	// Save the current state
	return i.conn.requestData(wyc.MSG_UPDATE_ITEM, i.data, &i.data)
}

func (i *item) GetHighestVersion() (*version, error) {
	ver := version{
		conn: i.conn,
		data: wyc.Version{
			Parent: i.data.Id,
		},
	}
	err := i.conn.requestData(wyc.MSG_FIND_HIGHEST_VERSION, &ver.data, &ver.data)
	if err != nil {
		return nil, err
	}

	return &ver, nil
}

func (i *item) CreateNextVersion() (*version, error) {
	facets := map[string]string{}
	parentCol, ok := i.data.Facets["collection"]
	if ok {
		facets["collection"] = parentCol
	}
	facets["itemtype"] = i.data.ItemType
	facets["variant"] = i.data.Variant

	ver := version{
		conn: i.conn,
		data: wyc.Version{
			Parent: i.data.Id,
			Facets: facets,
		},
	}

	err := i.conn.requestData(wyc.MSG_CREATE_VERSION, &ver.data, &ver.data)
	if err != nil {
		return nil, err
	}

	return &ver, nil
}

func (i *item) Parent() string {
	return i.data.Parent
}

func (i *item) GetParent() (*collection, error) {
	qry := []*wyc.QueryDesc{
		{Id: i.data.Parent},
	}

	results := []wyc.Collection{}
	err := i.conn.requestData(wyc.MSG_FIND_COLLECTION, &qry, &results)
	if err != nil {
		return nil, err
	}

	if len(results) < 1 {
		return nil, errors.New(fmt.Sprintf("Item with Id %s not found", i.data.Parent))
	}
	return &collection{
		conn: i.conn,
		data: results[0],
	}, nil
}
