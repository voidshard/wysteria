package wysteria_client

import (
	wyc "wysteria/wysteria_common"
	"errors"
)

type search struct {
	conn *wysteriaClient
	query []wyc.QueryDesc
	nextQuery wyc.QueryDesc
	nextQValid bool
}

func (i *search) Clear() *search {
	i.query = []wyc.QueryDesc{}
	i.nextQuery = wyc.QueryDesc{}
	i.nextQValid = false
	return i
}

func (i *search) do(route string, results interface{}) error {
	if i.nextQValid {
		i.query = append(i.query, i.nextQuery)
		i.nextQValid = false
		i.nextQuery = wyc.QueryDesc{}
	}

	if len(i.query) < 1 {
		return errors.New("You must specify at least one query term.")
	}

	err := i.conn.requestData(route, i.query, &results)
	if err != nil {
		return err
	}
	return nil
}

func (i *search) Items() ([]*item, error) {
	results := []wyc.Item{}
	err := i.do(wyc.MSG_FIND_ITEM, &results)
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
	var results []wyc.Version
	err := i.do(wyc.MSG_FIND_VERSION, &results)
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

func (i *search) FileResources() ([]*fileResource, error) {
	var results []wyc.FileResource
	err := i.do(wyc.MSG_FIND_FILERESOURCE, &results)
	if err != nil {
		return nil, err
	}

	ret := []*fileResource{}
	for _, r := range results {
		ret = append(ret, &fileResource{
			conn: i.conn,
			data: r,
		})
	}
	return ret, nil
}

func (i *search) Links() ([]*link, error) {
	var results []wyc.Link
	err := i.do(wyc.MSG_FIND_LINK, &results)
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

func (i *search) Version(n int) *search {
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
		i.nextQuery = wyc.QueryDesc{}
	}
	return i
}
