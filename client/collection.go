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
	items, err := c.conn.middleware.FindItems(
		[]*wyc.QueryDesc{{Parent: c.data.Id}},
	)
	if err != nil {
		return nil, err
	}

	results := []*item{}
	for _, i := range items {
		results = append(results, &item{
			conn: c.conn,
			data: i,
		})
	}
	return results, nil
}

func (c *collection) CreateItem(itemtype, variant string) (*item, error) {
	cmn_item := &wyc.Item{
		Parent:   c.data.Id,
		ItemType: itemtype,
		Variant:  variant,
		Facets: map[string]string{
			"collection": c.data.Name,
		},
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
	collection_id,  err := w.middleware.CreateCollection(name)
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
	results, err := w.middleware.FindCollections(
		[]*wyc.QueryDesc{{Id: identifier}, {Name: identifier}},
	)
	if err != nil {
		return nil, err
	}

	if len(results) == 1 {
		return &collection{ conn: w, data: results[0]}, nil
	}
	return nil, errors.New(fmt.Sprintf("Expected 1 result, got %s", len(results)))
}
