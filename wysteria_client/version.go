package wysteria_client

import (
	wyc "wysteria/wysteria_common"
	"fmt"
	"errors"
)

type version struct {
	conn *wysteriaClient
	data wyc.Version
}

func (i *version) Version() int {
	return i.data.Number
}

func (c *version) Delete() error {
	return c.conn.requestData(wyc.MSG_DELETE_VERSION, &c.data, nil)
}

func (i *version) Facet(key string) (string, bool) {
	val, ok := i.data.Facets[key]
	return val, ok
}

func (i *version) Id() string {
	return i.data.Id
}

func (i *version) SetFacet(key, value string) error {
	i.data.Facets[key] = value
	return i.update()
}

func (i *version) update() error {
	return i.conn.requestData(wyc.MSG_UPDATE_VERSION, i.data, &i.data)
}

func (i *version) AddResource(name, rtype, location string) (string, error) {
	res := wyc.FileResource{
		Parent: i.data.Id,
		Name: name,
		ResourceType: rtype,
		Location: location,
	}
	err := i.conn.requestData(wyc.MSG_CREATE_FILERESOURCE, res, &res)
	if err != nil {
		return "", err
	}
	i.data.Resources = append(i.data.Resources, res.Id)
	return res.Id, i.update()
}

func (i *version) GetAllResources() ([]*fileResource, error) {
	return i.getResources("", "")
}

func (i *version) GetResources(name, resource_type string) ([]*fileResource, error) {
	return i.getResources(name, resource_type)
}

func (i *version) GetResourcesByType(resource_type string) ([]*fileResource, error) {
	return i.getResources("", resource_type)
}

func (i *version) GetResourcesByName(name string) ([]*fileResource, error) {
	return i.getResources(name, "")
}

func (i *version) getResources(name, resource_type string) ([]*fileResource, error) {
	q := wyc.QueryDesc{Parent: i.data.Id, Name: name, ResourceType: resource_type}
	query := []wyc.QueryDesc{q}

	res := []wyc.FileResource{}
	err := i.conn.requestData(wyc.MSG_FIND_FILERESOURCE, &query, &res)
	if err != nil {
		return nil, err
	}

	items := []*fileResource{}
	for _, data := range res {
		items = append(items, &fileResource{
			conn: i.conn,
			data: data,
		})
	}
	return items, nil
}

func (i *version) getLinkedVersions(name string) ([]*version, map[string]string, error) {
	query := []wyc.QueryDesc{}
	for _, link_id := range i.data.Links {
		q := wyc.QueryDesc{LinkSrc: i.data.Id, Id: link_id, Name: name}
		query = append(query, q)
	}

	lnks := []wyc.Link{}
	err := i.conn.requestData(wyc.MSG_FIND_LINK, &query, &lnks)
	if err != nil {
		return nil, nil, err
	}

	lnk_name_map := map[string]string{}
	iquery := []wyc.QueryDesc{}
	for _, lnk := range lnks {
		lnk_name_map[lnk.Dst] = lnk.Name
		iquery = append(iquery, wyc.QueryDesc{Id: lnk.Dst})
	}

	data := []wyc.Version{}
	err = i.conn.requestData(wyc.MSG_FIND_VERSION, &iquery, &data)
	if err != nil {
		return nil, nil, err
	}

	items := []*version{}
	for _, item := range data {
		items = append(items, &version{
			conn: i.conn,
			data: item,
		})
	}
	return items, lnk_name_map, nil
}

func (i *version) GetLinkedVersionsByName(name string) ([]*version, error) {
	versions, _, err := i.getLinkedVersions(name)
	return versions, err
}

func (i *version) GetLinkedVersions() (map[string][]*version, error) {
	versions, name_map, err := i.getLinkedVersions("")
	if err != nil {
		return nil, err
	}

	results := map[string][]*version{}
	for _, v := range versions {
		name, ok := name_map[v.Id()]
		if !ok {
			continue
		}

		ls, ok := results[name]
		if !ok {
			ls = []*version{v}
		} else {
			ls = append(ls, v)
		}
		results[name] = ls
	}
	return results, nil
}

func (i *version) LinkTo(name string, other *version) error {
	if i.Id() == other.Id() { // Prevent linking to self
		return nil
	}
	for _, lid := range i.data.Links { // Prevent duplicate links
		if lid == other.Id() {
			return nil
		}
	}

	lnk := wyc.Link{
		Name: name,
		Src: i.data.Id,
		Dst: other.data.Id,
	}
	err := i.conn.requestData(wyc.MSG_CREATE_LINK, lnk, &lnk)
	if err != nil {
		return err
	}
	i.data.Links = append(i.data.Links, lnk.Id)
	return i.update()
}

func (i *version) Parent() string {
	return i.data.Parent
}

func (i *version) GetParent() (*item, error) {
	qry := []*wyc.QueryDesc{
		{Id: i.data.Parent},
	}

	results := []wyc.Item{}
	err := i.conn.requestData(wyc.MSG_FIND_ITEM, &qry, &results)
	if err != nil {
		return nil, err
	}

	if len(results) < 1 {
		return nil, errors.New(fmt.Sprintf("Item with Id %s not found", i.data.Parent))
	}
	return &item{conn: i.conn, data: results[0]}, nil
}
