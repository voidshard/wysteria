package wysteria_client

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

type resource struct {
	conn *wysteriaClient
	data *wyc.Resource
}

func (i *resource) Name() string {
	return i.data.Name
}

func (i *resource) Type() string {
	return i.data.ResourceType
}

func (i *resource) Delete() error {
	return i.conn.middleware.DeleteResource(i.data.Id)
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
	versions, err := i.conn.middleware.FindVersions([]*wyc.QueryDesc{
		{Id: i.data.Parent},
	})
	if err != nil {
		return nil, err
	}
	if len(versions) < 1 {
		return nil, errors.New(fmt.Sprintf("Version with Id %s not found", i.data.Parent))
	}

	return &version{
		conn: i.conn,
		data: versions[0],
	}, nil
}
