package wysteria_client

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

type collection struct {
	conn *wysteriaClient
	data *wyc.Collection
}

func (c *collection) Name() string {
	return c.data.Name
}

func (c *collection) Id() string {
	return c.data.Id
}

func (c *collection) Delete() error {
	return c.conn.middleware.DeleteCollection(c.data.Id)
}

func (c *collection) GetItems() ([]*item, error) {
	return c.conn.Search().ChildOf(c.data.Id).Items()
}


func (c *collection) CreateItem(itemtype, variant string, facets map[string]string) (*item, error) {
	all_facets := map[string]string{}
	if facets != nil {
		for key, value := range facets {
			all_facets[key] = value
		}
	}

	all_facets["collection"] = c.data.Name

	cmn_item := &wyc.Item{
		Parent:   c.data.Id,
		ItemType: itemtype,
		Variant:  variant,
		Facets: all_facets,
	}

	item_id, err := c.conn.middleware.CreateItem(cmn_item)
	if err != nil {
		return nil, err
	}
	cmn_item.Id = item_id

	return &item{
		conn: c.conn,
		data: cmn_item,
	}, nil
}

func (w *wysteriaClient) CreateCollection(name string) (*collection, error) {
	collection_id, err := w.middleware.CreateCollection(name)
	if err != nil {
		return nil, err
	}

	return &collection{
		conn: w,
		data: &wyc.Collection{
			Id: collection_id,
			Name: name,
		},
	}, nil
}

func (w *wysteriaClient) GetCollection(identifier string) (*collection, error) {
	results, err := w.Search().Id(identifier).Or().Name(identifier).Collections()
	if err != nil {
		return nil, err
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return nil, errors.New(fmt.Sprintf("Expected 1 result, got %d", len(results)))
}
