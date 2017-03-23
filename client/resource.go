package wysteria_client

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

type Resource struct {
	conn *wysteriaClient
	data *wyc.Resource
}

func (i *Resource) Name() string {
	return i.data.Name
}

func (i *Resource) Type() string {
	return i.data.ResourceType
}

func (i *Resource) Delete() error {
	return i.conn.middleware.DeleteResource(i.data.Id)
}

func (i *Resource) Id() string {
	return i.data.Id
}

func (i *Resource) Location() string {
	return i.data.Location
}

func (i *Resource) Parent() string {
	return i.data.Parent
}

func (i *Resource) GetParent() (*Version, error) {
	versions, err := i.conn.middleware.FindVersions([]*wyc.QueryDesc{
		{Id: i.data.Parent},
	})
	if err != nil {
		return nil, err
	}
	if len(versions) < 1 {
		return nil, errors.New(fmt.Sprintf("Version with Id %s not found", i.data.Parent))
	}

	return &Version{
		conn: i.conn,
		data: versions[0],
	}, nil
}
