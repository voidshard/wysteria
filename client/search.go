package wysteria_client

import (
	wyc "github.com/voidshard/wysteria/common"
	"errors"
)

// Search obj represents a query or set of queries that are about to be sent
// to the server
type search struct {
	conn       *wysteriaClient
	query      []*wyc.QueryDesc
	nextQuery  *wyc.QueryDesc
	nextQValid bool
	limit      int32
	offset     int32
}

type SearchOptionFunc func(*search)

// Apply given options to the build search query
func (i *search) applyOptions(opts ...SearchOptionFunc) {
	for _, option := range opts {
		option(i)
	}
}

// Create a new query description.
// A query description is a collection of fields that are together understood as an "AND" operation
func newQuery() *wyc.QueryDesc {
	return &wyc.QueryDesc{
		Facets: map[string]string{},
	}
}

// All of the search params before this are considered "AND" (to the Search obj or the last "Or"),
// this adds a new sub query to find another object based on more params.
func (i *search) Or(opts ...SearchOptionFunc) *search {
	if i.nextQValid {
		i.nextQValid = false
		i.query = append(i.query, i.nextQuery)
		i.nextQuery = newQuery()
	}
	i.applyOptions(opts...)
	return i
}

// Set limit on number of results
func Limit(val int) SearchOptionFunc {
	return func(i *search) {
		if val < 1 {
			return
		}
		i.limit = int32(val)
	}
}

// Get results from the given offset.
func Offset(val int) SearchOptionFunc {
	return func(i *search) {
		if val > -1 {
			i.offset = int32(val)
		}
	}
}

// Search for something with the given Id
func Id(s string) SearchOptionFunc {
	return func(i *search) {
		i.nextQValid = true
		i.nextQuery.Id = s
	}
}

// Search for a resource with the given ResourceType
func ResourceType(s string) SearchOptionFunc {
	return func(i *search) {
		i.nextQValid = true
		i.nextQuery.ResourceType = s
	}
}

// Search for something whose parent has the given Id
func ChildOf(s string) SearchOptionFunc {
	return func(i *search) {
		i.nextQValid = true
		i.nextQuery.Parent = s
	}
}

// Search for a link whose source is the given Id
func LinkSource(s string) SearchOptionFunc {
	return func(i *search) {
		i.nextQValid = true
		i.nextQuery.LinkSrc = s
	}
}

// Search for a link whose destination is the given Id
func LinkDestination(s string) SearchOptionFunc {
	return func(i *search) {
		i.nextQValid = true
		i.nextQuery.LinkDst = s
	}
}

// Search for an item with the given type
func ItemType(s string) SearchOptionFunc {
	return func(i *search) {
		i.nextQValid = true
		i.nextQuery.ItemType = s
	}
}

// Search for an item with the given variant
func ItemVariant(s string) SearchOptionFunc {
	return func(i *search) {
		i.nextQValid = true
		i.nextQuery.Variant = s
	}
}

// Search for a version with the given version number
func VersionNumber(n int32) SearchOptionFunc {
	return func(i *search) {
		i.nextQValid = true
		i.nextQuery.VersionNumber = n
	}
}

// Search for something that has all of the given facets
func HasFacets(f map[string]string) SearchOptionFunc {
	return func(i *search) {
		i.nextQValid = true
		i.nextQuery.Facets = f
	}
}

// Search for something with a name matching the given name
func Name(s string) SearchOptionFunc {
	return func(i *search) {
		i.nextQValid = true
		i.nextQuery.Name = s
	}
}

// Search for a resource with the given location
func ResourceLocation(s string) SearchOptionFunc {
	return func(i *search) {
		i.nextQValid = true
		i.nextQuery.Location = s
	}
}

// Find all matching Collections given our search params
func (i *search) FindCollections() ([]*Collection, error) {
	err := i.ready()
	if err != nil {
		return nil, err
	}

	results, err := i.conn.middleware.FindCollections(i.limit, i.offset, i.query)
	if err != nil {
		return nil, err
	}

	ret := []*Collection{}
	for _, r := range results {
		ret = append(ret, &Collection{
			conn: i.conn,
			data: r,
		})
	}
	return ret, nil
}

// Find all matching Items given our search params
func (i *search) FindItems() ([]*Item, error) {
	err := i.ready()
	if err != nil {
		return nil, err
	}

	results, err := i.conn.middleware.FindItems(i.limit, i.offset, i.query)
	if err != nil {
		return nil, err
	}

	ret := []*Item{}
	for _, r := range results {
		ret = append(ret, &Item{
			conn: i.conn,
			data: r,
		})
	}
	return ret, nil
}

// Find all matching Versions given our search params
func (i *search) FindVersions() ([]*Version, error) {
	err := i.ready()
	if err != nil {
		return nil, err
	}

	results, err := i.conn.middleware.FindVersions(i.limit, i.offset, i.query)
	if err != nil {
		return nil, err
	}

	ret := []*Version{}
	for _, r := range results {
		ret = append(ret, &Version{
			conn: i.conn,
			data: r,
		})
	}
	return ret, nil
}

// Find all matching Resources given our search params
func (i *search) FindResources() ([]*Resource, error) {
	err := i.ready()
	if err != nil {
		return nil, err
	}

	results, err := i.conn.middleware.FindResources(i.limit, i.offset, i.query)
	if err != nil {
		return nil, err
	}

	ret := []*Resource{}
	for _, r := range results {
		ret = append(ret, &Resource{
			conn: i.conn,
			data: r,
		})
	}
	return ret, nil
}

// Ready the current query description object if it is valid.
// If we have been given nothing to search for, error
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

// Find all matching Links given our search params
func (i *search) FindLinks() ([]*Link, error) {
	err := i.ready()
	if err != nil {
		return nil, err
	}

	results, err := i.conn.middleware.FindLinks(i.limit, i.offset, i.query)
	if err != nil {
		return nil, err
	}

	ret := []*Link{}
	for _, r := range results {
		ret = append(ret, &Link{
			conn: i.conn,
			data: r,
		})
	}
	return ret, nil
}
