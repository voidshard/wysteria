package wysteria_client

import (
	"errors"
	wyc "github.com/voidshard/wysteria/common"
)

type search struct {
	conn       *wysteriaClient
	query      []*wyc.QueryDesc
	nextQuery  *wyc.QueryDesc
	nextQValid bool
}

func newQuery() *wyc.QueryDesc {
	return &wyc.QueryDesc{
		Facets: map[string]string{},
	}
}

func (i *search) Clear() *search {
	i.query = []*wyc.QueryDesc{}
	i.nextQuery = newQuery()
	i.nextQValid = false
	return i
}

func (i *search) ready() error {
	if i.nextQValid {
		i.query = append(i.query, i.nextQuery)
		i.nextQValid = false
		i.nextQuery = newQuery()
	}

	if len(i.query) < 1 {
		return errors.New("You must specify at least one query term.")
	}
	return nil
}

func (i *search) Collections() ([]*collection, error) {
	err := i.ready()
	if err != nil {
		return nil, err
	}

	results, err := i.conn.middleware.FindCollections(i.query)
	if err != nil {
		return nil, err
	}

	ret := []*collection{}
	for _, r := range results {
		ret = append(ret, &collection{
			conn: i.conn,
			data: r,
		})
	}
	return ret, nil
}

func (i *search) Items() ([]*item, error) {
	err := i.ready()
	if err != nil {
		return nil, err
	}

	results, err := i.conn.middleware.FindItems(i.query)
	if err != nil {
		return nil, err
	}

	ret := []*item{}
	for _, r := range results {
		ret = append(ret, &item{
			conn: i.conn,
			data: r,
		})
	}
	return ret, nil
}

func (i *search) Versions() ([]*version, error) {
	err := i.ready()
	if err != nil {
		return nil, err
	}

	results, err := i.conn.middleware.FindVersions(i.query)
	if err != nil {
		return nil, err
	}

	ret := []*version{}
	for _, r := range results {
		ret = append(ret, &version{
			conn: i.conn,
			data: r,
		})
	}
	return ret, nil
}

func (i *search) Resources() ([]*resource, error) {
	err := i.ready()
	if err != nil {
		return nil, err
	}

	results, err := i.conn.middleware.FindResources(i.query)
	if err != nil {
		return nil, err
	}

	ret := []*resource{}
	for _, r := range results {
		ret = append(ret, &resource{
			conn: i.conn,
			data: r,
		})
	}
	return ret, nil
}

func (i *search) Links() ([]*link, error) {
	err := i.ready()
	if err != nil {
		return nil, err
	}

	results, err := i.conn.middleware.FindLinks(i.query)
	if err != nil {
		return nil, err
	}

	ret := []*link{}
	for _, r := range results {
		ret = append(ret, &link{
			conn: i.conn,
			data: r,
		})
	}
	return ret, nil
}

func (i *search) Id(s string) *search {
	i.nextQValid = true
	i.nextQuery.Id = s
	return i
}

func (i *search) ResourceType(s string) *search {
	i.nextQValid = true
	i.nextQuery.ResourceType = s
	return i
}

func (i *search) ChildOf(s string) *search {
	i.nextQValid = true
	i.nextQuery.Parent = s
	return i
}

func (i *search) Src(s string) *search {
	i.nextQValid = true
	i.nextQuery.LinkSrc = s
	return i
}

func (i *search) Dst(s string) *search {
	i.nextQValid = true
	i.nextQuery.LinkDst = s
	return i
}

func (i *search) ItemType(s string) *search {
	i.nextQValid = true
	i.nextQuery.ItemType = s
	return i
}

func (i *search) ItemVariant(s string) *search {
	i.nextQValid = true
	i.nextQuery.Variant = s
	return i
}

func (i *search) Version(n int32) *search {
	i.nextQValid = true
	i.nextQuery.VersionNumber = n
	return i
}

func (i *search) HasFacets(f map[string]string) *search {
	i.nextQValid = true
	i.nextQuery.Facets = f
	return i
}

func (i *search) Name(s string) *search {
	i.nextQValid = true
	i.nextQuery.Name = s
	return i
}

func (i *search) Location(s string) *search {
	i.nextQValid = true
	i.nextQuery.Location = s
	return i
}

func (i *search) Or() *search {
	if i.nextQValid {
		i.nextQValid = false
		i.query = append(i.query, i.nextQuery)
		i.nextQuery = newQuery()
	}
	return i
}
