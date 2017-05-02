package wysteria_client

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

type Version struct {
	conn     *wysteriaClient
	data     *wyc.Version
}

func (i *Version) Version() int32 {
	return i.data.Number
}

func (i *Version) Delete() error {
	return i.conn.middleware.DeleteVersion(i.data.Id)
}

func (i *Version) Facet(key string) (string, bool) {
	val, ok := i.data.Facets[key]
	return val, ok
}

func (i *Version) Id() string {
	return i.data.Id
}

func (i *Version) SetFacets(in map[string]string) error {
	return i.conn.middleware.UpdateVersionFacets(i.data.Id, in)
}

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

func (i *Version) GetLinkedByName(name string) ([]*Version, error) {
	found, err := i.getLinkedVersions(name)
	if err != nil {
		return nil, err
	}
	return found[name], nil
}

func (i *Version) GetLinked() (map[string][]*Version, error) {
	return i.getLinkedVersions("")
}

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

func (i *Version) Publish() error {
	return i.conn.middleware.PublishVersion(i.data.Id)
}

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

func (i *Version) GetAllResources() ([]*Resource, error) {
	return i.getResources("", "")
}

func (i *Version) GetResources(name, resource_type string) ([]*Resource, error) {
	return i.getResources(name, resource_type)
}

func (i *Version) GetResourcesByType(resource_type string) ([]*Resource, error) {
	return i.getResources("", resource_type)
}

func (i *Version) GetResourcesByName(name string) ([]*Resource, error) {
	return i.getResources(name, "")
}

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

func (i *Version) Parent() string {
	return i.data.Parent
}

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
