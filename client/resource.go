package wysteria_client

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

// Wrapper around wysteria/common Resource obj
type Resource struct {
	conn *wysteriaClient
	data *wyc.Resource
}

// Return the name given for this resource
func (i *Resource) Name() string {
	return i.data.Name
}

// Return the resource type for this resource
func (i *Resource) Type() string {
	return i.data.ResourceType
}

// Delete this resource
func (i *Resource) Delete() error {
	return i.conn.middleware.DeleteResource(i.data.Id)
}

// Return the Id of this resource
func (i *Resource) Id() string {
	return i.data.Id
}

// Return the location string
func (i *Resource) Location() string {
	return i.data.Location
}

// Return the Id of this resources' parent (the Id of a Version object)
func (i *Resource) Parent() string {
	return i.data.Parent
}

// Return the parent (Version) of this resource
func (i *Resource) GetParent() (*Version, error) {
	versions, err := i.conn.middleware.FindVersions(
		1,
		0,
		[]*wyc.QueryDesc{
			{Id: i.data.Parent},
		},
	)
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
