package wysteria_client

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

// Wrapper around wysteria/common Version object
type Version struct {
	conn     *wysteriaClient
	data     *wyc.Version
}

// Return the version number associated with this version.
// Each version of a given item is numbered starting from 1
func (i *Version) Version() int32 {
	return i.data.Number
}

// Delete this version, and any & all children
func (i *Version) Delete() error {
	return i.conn.middleware.DeleteVersion(i.data.Id)
}

// Get the facet value and a bool indicating if the value exists for the given key.
func (i *Version) Facet(key string) (string, bool) {
	val, ok := i.data.Facets[key]
	return val, ok
}

// Get this version's Id
func (i *Version) Id() string {
	return i.data.Id
}

// Set all the key:value pairs given on this Item's facets.
// Note that the server will ignore the setting of reserved facets.
func (i *Version) SetFacets(in map[string]string) error {
	return i.conn.middleware.UpdateVersionFacets(i.data.Id, in)
}

// Find and return all linked Versions for which links exist that name this as the source.
// That is, this first finds all links for which the source Id is this Version's Id, then
// gets all matching Versions.
// Since this would cause us to lose the link 'name' we return a map of link name -> []*Version
func (i *Version) getLinkedVersions(name string) (map[string][]*Version, error) {
	links, err := i.conn.middleware.FindLinks(
		[]*wyc.QueryDesc{
			{LinkSrc: i.data.Id, Name: name},
		},
	)
	if err != nil {
		return nil, err
	}

	version_id_to_link := map[string]*wyc.Link{}
	ids := []*wyc.QueryDesc{}
	for _, link := range links {
		id := link.Src

		if link.Src == i.data.Id {
			id = link.Dst
		}

		version_id_to_link[id] = link
		ids = append(ids, &wyc.QueryDesc{Id: id})
	}

	items, err := i.conn.middleware.FindVersions(ids)
	if err != nil {
		return nil, err
	}

	result := map[string][]*Version{}
	for _, ver := range items {
		lnk, ok := version_id_to_link[ver.Id]
		if !ok {
			continue
		}

		result_list, ok := result[lnk.Name]
		if result_list == nil {
			result_list = []*Version{}
		}

		wrapped_item := &Version{
			conn: i.conn,
			data: ver,
		}

		result_list = append(result_list, wrapped_item)
		result[lnk.Name] = result_list
	}
	return result, nil
}

// Get all linked Versions (Versions where links exist that mention this as the source and them as the destination)
// where the link name is the given 'name'.
func (i *Version) GetLinkedByName(name string) ([]*Version, error) {
	found, err := i.getLinkedVersions(name)
	if err != nil {
		return nil, err
	}
	return found[name], nil
}


// Find and return all linked Versions for which links exist that name this as the source.
// That is, this first finds all links for which the source Id is this Version's Id, then
// gets all matching Versions.
// Since this would cause us to lose the link 'name' we return a map of link name -> []*Version
func (i *Version) GetLinked() (map[string][]*Version, error) {
	return i.getLinkedVersions("")
}

// Link this Version with a link described by 'name' to some other Version.
func (i *Version) LinkTo(name string, other *Version) error {
	if i.Id() == other.Id() { // Prevent linking to oneself
		return nil
	}

	lnk := &wyc.Link{
		Name: name,
		Src:  i.Id(),
		Dst:  other.Id(),
	}
	_, err := i.conn.middleware.CreateLink(lnk)
	return err
}

// Mark this Version as the published version.
// An item can only have one 'publised' version at a time.
func (i *Version) Publish() error {
	return i.conn.middleware.PublishVersion(i.data.Id)
}

// Add a resource with the given name, type and location to this version.
func (i *Version) AddResource(name, rtype, location string) error {
	res := &wyc.Resource{
		Parent:       i.data.Id,
		Name:         name,
		ResourceType: rtype,
		Location:     location,
	}

	id, err := i.conn.middleware.CreateResource(res)
	if err != nil {
		return err
	}
	res.Id = id
	return nil
}

// Retrieve all resources whose parent is this Version
func (i *Version) GetAllResources() ([]*Resource, error) {
	return i.getResources("", "")
}

// Retrieve all child resources of this Version with the given name & resource type
func (i *Version) GetResources(name, resource_type string) ([]*Resource, error) {
	return i.getResources(name, resource_type)
}

// Retrieve all child resources of this Version with the given resource type
func (i *Version) GetResourcesByType(resource_type string) ([]*Resource, error) {
	return i.getResources("", resource_type)
}

// Retrieve all child resources of this Version with the given name
func (i *Version) GetResourcesByName(name string) ([]*Resource, error) {
	return i.getResources(name, "")
}

// Retrieve all child resources of this Version with the given name & resource type
func (i *Version) getResources(name, resource_type string) ([]*Resource, error) {
	results, err := i.conn.middleware.FindResources(
		[]*wyc.QueryDesc{{Parent: i.data.Id, Name: name, ResourceType: resource_type}},
	)
	if err != nil {
		return nil, err
	}

	items := []*Resource{}
	for _, data := range results {
		items = append(items, &Resource{
			conn: i.conn,
			data: data,
		})
	}
	return items, nil
}

// Return the Id of this Version's parent
func (i *Version) Parent() string {
	return i.data.Parent
}

// Get the parent Item of this Version
func (i *Version) GetParent() (*Item, error) {
	items, err := i.conn.middleware.FindItems(
		[]*wyc.QueryDesc{{Id: i.data.Parent}},
	)
	if err != nil {
		return nil, err
	}
	if len(items) < 1 {
		return nil, errors.New(fmt.Sprintf("Expected 1 result, got %s", len(items)))
	}
	return &Item{
		conn: i.conn,
		data: items[0],
	}, nil
}
