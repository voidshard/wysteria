package wysteria_client

import (
	wyc "wysteria/wysteria_common"
	"errors"
	"fmt"
)

type fileResource struct {
	conn *wysteriaClient
	data wyc.FileResource
}

func (i *fileResource) Name() string {
	return i.data.Name
}

func (i *fileResource) Type() string {
	return i.data.ResourceType
}

func (c *fileResource) Delete() error {
	return c.conn.requestData(wyc.MSG_DELETE_FILERESOURCE, &c.data, nil)
}

func (i *fileResource) Id() string {
	return i.data.Id
}

func (i *fileResource) Location() string {
	return i.data.Location
}

func (i *fileResource) Parent() string {
	return i.data.Parent
}

func (i *fileResource) GetParent() (*version, error) {
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

