package wysteria_client

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

type version struct {
	conn *wysteriaClient
	data *wyc.Version
	fromLink *link
}

func (i *version) Version() int32 {
	return i.data.Number
}

func (i *version) Link() *link {
	return i.fromLink
}

func (i *version) Delete() error {
	return i.conn.middleware.DeleteVersion(i.data.Id)
}

func (i *version) Facet(key string) (string, bool) {
	val, ok := i.data.Facets[key]
	return val, ok
}

func (i *version) Id() string {
	return i.data.Id
}

func (i *version) SetFacets(in map[string]string) error {
	return i.conn.middleware.UpdateVersionFacets(i.data.Id, in)
}

func (i *version) getLinkedVersions(name string) ([]*version, error) {
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

	result := []*version{}
	for _, ver := range items {
		wrapped_item := &version{
			conn: i.conn,
			data: ver,
		}

		lnk, ok := version_id_to_link[ver.Id]
		if ok {
			wrapped_item.fromLink = &link{conn: i.conn, data: lnk}
		}
		result = append(result, wrapped_item)
	}
	return result, nil
}


func (i *version) GetLinkedVersionsByName(name string) ([]*version, error) {
	return i.getLinkedVersions(name)
}

func (i *version) GetLinkedVersions() ([]*version, error) {
	return i.getLinkedVersions("")
}

func (i *version) LinkTo(name string, other *version) error {
	if i.Id() == other.Id() { // Prevent linking to oneself
		return nil
	}

	lnk := &wyc.Link{
		Name: name,
		Src:  i.data.Id,
		Dst:  other.data.Id,
	}
	_, err := i.conn.middleware.CreateLink(lnk)
	return err
}

func (i *version) AddResource(name, rtype, location string) (string, error) {
	res := &wyc.Resource{
		Parent:       i.data.Id,
		Name:         name,
		ResourceType: rtype,
		Location:     location,
	}

	return i.conn.middleware.CreateResource(res)
}

func (i *version) GetAllResources() ([]*resource, error) {
	return i.getResources("", "")
}

func (i *version) GetResources(name, resource_type string) ([]*resource, error) {
	return i.getResources(name, resource_type)
}

func (i *version) GetResourcesByType(resource_type string) ([]*resource, error) {
	return i.getResources("", resource_type)
}

func (i *version) GetResourcesByName(name string) ([]*resource, error) {
	return i.getResources(name, "")
}

func (i *version) getResources(name, resource_type string) ([]*resource, error) {
	results, err := i.conn.middleware.FindResources(
		[]*wyc.QueryDesc{{Parent: i.data.Id, Name: name, ResourceType: resource_type}},
	)
	if err != nil {
		return nil, err
	}

	items := []*resource{}
	for _, data := range results {
		items = append(items, &resource{
			conn: i.conn,
			data: data,
		})
	}
	return items, nil
}

func (i *version) Parent() string {
	return i.data.Parent
}

func (i *version) GetParent() (*item, error) {
	items, err := i.conn.middleware.FindItems(
		[]*wyc.QueryDesc{{Id: i.data.Parent}},
	)
	if err != nil {
		return nil, err
	}
	if len(items) < 1 {
		return nil, errors.New(fmt.Sprintf("Expected 1 result, got %s", len(items)))
	}
	return &item{
		conn: i.conn,
		data: items[0],
	}, nil
}
