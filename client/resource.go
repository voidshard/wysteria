package wysteria_client

import (
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

// Wrapper around wysteria/common Resource obj
type Resource struct {
	conn *Client
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

// Get the facet value and a bool indicating if the value exists for the given key.
func (i *Resource) Facet(key string) (string, bool) {
	val, ok := i.data.Facets[key]
	return val, ok
}

// Get all facets
func (i *Resource) Facets() map[string]string {
	return i.data.Facets
}

// Set all the key:value pairs given on this Resource's facets.
func (i *Resource) SetFacets(in map[string]string) error {
	if in == nil {
		return nil
	}
	for k, v := range in {
		i.data.Facets[k] = v
	}
	return i.conn.middleware.UpdateResourceFacets(i.data.Id, in)
}

// Set initial user defined facets
func (i *Resource) initUserFacets(in map[string]string) {
	i.data.Facets = in
}

// Return the location string
func (i *Resource) Location() string {
	return i.data.Location
}

// Return the Id of this resources' parent (the Id of a Version object)
func (i *Resource) ParentId() string {
	return i.data.Parent
}

// Return the parent (Version) of this resource
func (i *Resource) Parent() (*Version, error) {
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
		return nil, fmt.Errorf("version with Id %s not found", i.data.Parent)
	}

	return &Version{
		conn: i.conn,
		data: versions[0],
	}, nil
}
