package wysteria_client

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

type resource struct {
	conn *wysteriaClient
	data wyc.Resource
}

func (i *resource) Name() string {
	return i.data.Name
}

func (i *resource) Type() string {
	return i.data.ResourceType
}

func (c *resource) Delete() error {
	return c.conn.requestData(wyc.MSG_DELETE_RESOURCE, &c.data, nil)
}

func (i *resource) Id() string {
	return i.data.Id
}

func (i *resource) Location() string {
	return i.data.Location
}

func (i *resource) Parent() string {
	return i.data.Parent
}

func (i *resource) GetParent() (*version, error) {
	qry := []*wyc.QueryDesc{
		{Id: i.data.Parent},
	}

	results := []wyc.Version{}
	err := i.conn.requestData(wyc.MSG_FIND_VERSION, &qry, &results)
	if err != nil {
		return nil, err
	}

	if len(results) < 1 {
		return nil, errors.New(fmt.Sprintf("Version with Id %s not found", i.data.Parent))
	}
	return &version{
		conn: i.conn,
		data: results[0],
	}, nil
}
