package wysteria_client

import (
	wyc "github.com/voidshard/wysteria/common"
)

// Wrapper around wysteria/common Link object
type Link struct {
	conn *Client
	data *wyc.Link
}

// Return the name of this link
func (i *Link) Name() string {
	return i.data.Name
}

// Return the Id of this link
func (i *Link) Id() string {
	return i.data.Id
}

// Get the facet value and a bool indicating if the value exists for the given key.
func (i *Link) Facet(key string) (string, bool) {
	val, ok := i.data.Facets[key]
	return val, ok
}

// Get all facets
func (i *Link) Facets() map[string]string {
	return i.data.Facets
}

// Set all the key:value pairs given on this Link's facets.
func (i *Link) SetFacets(in map[string]string) error {
	if in == nil {
		return nil
	}
	for k, v := range in {
		i.data.Facets[k] = v
	}
	return i.conn.middleware.UpdateLinkFacets(i.data.Id, in)
}

// Return the Id of the object (either a Version or Item) that is considered
// the "source" of this link.
func (i *Link) SourceId() string {
	return i.data.Src
}

// Return the Id of the object (either a Version or Item) that is considered
// the "destination" of this link.
func (i *Link) DestinationId() string {
	return i.data.Dst
}

// Set initial user defined facets
func (i *Link) initUserFacets(in map[string]string) {
	i.data.Facets = in
}
