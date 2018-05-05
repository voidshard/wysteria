package wysteria_client

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

// Wrapper around wysteria/common Version object
type Version struct {
	conn *Client
	data *wyc.Version
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

// Get the URI for this version
func (i *Version) Uri() string {
	if i.data.Uri == "" {
		results, err := i.conn.Search(Id(i.Id())).FindVersions(Limit(1))
		if len(results) == 1 && err == nil {
			i.data.Uri = results[0].Uri()
		}
	}
	return i.data.Uri
}

// Set all the key:value pairs given on this Item's facets.
// Note that the server will ignore the setting of reserved facets.
func (i *Version) SetFacets(in map[string]string) error {
	if in == nil {
		return nil
	}
	for k, v := range in {
		i.data.Facets[k] = v
	}
	return i.conn.middleware.UpdateVersionFacets(i.data.Id, in)
}

// Find and return all linked Versions for which links exist that name this as the source.
// That is, this first finds all links for which the source Id is this Version's Id, then
// gets all matching Versions.
// Since this would cause us to lose the link 'name' we return a map of link name -> []*Version
func (i *Version) Linked(opts ...SearchParam) (map[string][]*Version, error) {
	opts = append(opts, LinkSource(i.Id()))
	links, err := i.conn.Search(opts...).FindLinks()
	if err != nil {
		return nil, err
	}

	versionIdToLinks := map[string]*Link{}
	ids := []*wyc.QueryDesc{}
	for _, link := range links {
		id := link.SourceId()

		if link.SourceId() == i.data.Id {
			id = link.DestinationId()
		}

		versionIdToLinks[id] = link
		ids = append(ids, &wyc.QueryDesc{Id: id})
	}

	items, err := i.conn.middleware.FindVersions(int32(len(ids)), 0, ids)
	if err != nil {
		return nil, err
	}

	result := map[string][]*Version{}
	for _, ver := range items {
		lnk, ok := versionIdToLinks[ver.Id]
		if !ok {
			continue
		}

		resultList, ok := result[lnk.Name()]
		if resultList == nil {
			resultList = []*Version{}
		}

		wrappedItem := &Version{
			conn: i.conn,
			data: ver,
		}

		resultList = append(resultList, wrappedItem)
		result[lnk.Name()] = resultList
	}
	return result, nil
}

// Link this Version with a link described by 'name' to some other Version.
func (i *Version) LinkTo(name string, other *Version, opts ...CreateOption) (*Link, error) {
	if i.Id() == other.Id() { // Prevent linking to oneself
		return nil, errors.New("link src and dst IDs cannot be equal")
	}

	lnk := &wyc.Link{
		Name:   name,
		Src:    i.Id(),
		Dst:    other.Id(),
		Facets: map[string]string{},
	}

	child := &Link{conn: i.conn, data: lnk}
	for _, opt := range opts {
		opt(i, child)
	}
	lnk.Facets[FacetLinkType] = FacetVersionLink

	id, err := i.conn.middleware.CreateLink(lnk)
	lnk.Id = id
	return child, err
}

// Mark this Version as the published version.
// An item can only have one 'published' version at a time.
func (i *Version) Publish() error {
	return i.conn.middleware.SetPublishedVersion(i.data.Id)
}

// Add a resource with the given name, type and location to this version.
func (i *Version) AddResource(name, rtype, location string, opts ...CreateOption) (*Resource, error) {
	res := &wyc.Resource{
		Parent:       i.data.Id,
		Name:         name,
		ResourceType: rtype,
		Location:     location,
		Facets:       map[string]string{},
	}

	child := &Resource{conn: i.conn, data: res}
	for _, opt := range opts {
		opt(i, child)
	}

	id, err := i.conn.middleware.CreateResource(res)
	if err != nil {
		return nil, err
	}
	res.Id = id
	return child, nil
}

// Retrieve all child resources of this Version with the given name & resource type
func (i *Version) Resources(opts ...SearchParam) ([]*Resource, error) {
	opts = append(opts, ChildOf(i.Id()))
	return i.conn.Search(opts...).FindResources()
}

// Return the Id of this Version's parent
func (i *Version) ParentId() string {
	return i.data.Parent
}

// Get all facets
func (i *Version) Facets() map[string]string {
	return i.data.Facets
}

// Get the parent Item of this Version
func (i *Version) Parent() (*Item, error) {
	items, err := i.conn.middleware.FindItems(
		1, 0, []*wyc.QueryDesc{{Id: i.data.Parent}},
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

// Set initial user defined facets
func (i *Version) initUserFacets(in map[string]string) {
	i.data.Facets = in
}
