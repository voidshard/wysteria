package wysteria_client

import (
	wyc "wysteria/wysteria_common"
	"fmt"
	"errors"
)

type collection struct {
	conn *wysteriaClient
	data wyc.Collection
}

func (c *collection) Name() string {
	return c.data.Name
}

func (c *collection) Id() string {
	return c.data.Id
}

func (c *collection) Delete() error {
	return c.conn.requestData(wyc.MSG_DELETE_COLLECTION, &c.data, nil)
}

func (c *collection) GetItems() ([]*item, error) {
	cdata := []wyc.Item{}
	query := []wyc.QueryDesc{{Parent: c.data.Id}}
	err := c.conn.requestData(wyc.MSG_FIND_ITEM, &query, &cdata)
	if err != nil {
		return nil, err
	}

	items := []*item{}
	for _, idata := range cdata {
		items = append(items, &item{
			conn: c.conn,
			data: idata,
		})
	}
	return items, nil
}


func (c *collection) CreateItem(itemtype, variant string) (*item, error) {
	i := item{
		conn: c.conn,
		data: wyc.Item{
			Parent: c.data.Id,
			ItemType: itemtype,
			Variant: variant,
			Facets: map[string]string {
				"collection": c.data.Name,
			},
		},
	}

	err := c.conn.requestData(wyc.MSG_CREATE_ITEM, &i.data, &i.data)
	if err != nil {
		return nil, err
	}

	return &i, nil
}

func (w *wysteriaClient) CreateCollection(name string) (*collection, error) {
	c := collection{
		conn: w,
		data: wyc.Collection{Name: name},
	}

	err := w.requestData(wyc.MSG_CREATE_COLLECTION, &c.data, &c.data)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (w *wysteriaClient) GetCollection(identifier string) (*collection, error) {
	cdata := []wyc.Collection{}
	query := []wyc.QueryDesc{
		{Id: identifier},
		{Name: identifier},
	}

	err := w.requestData(wyc.MSG_FIND_COLLECTION, &query, &cdata)
	if err != nil {
		return nil, err
	}

	if len(cdata) == 1 {
		return &collection{conn: w, data: cdata[0]}, nil
	}
	return nil, errors.New(fmt.Sprintf("Expected 1 result, got %s", len(cdata)))
}

