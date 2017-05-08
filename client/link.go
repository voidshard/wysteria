package wysteria_client

import (
	wyc "github.com/voidshard/wysteria/common"
)

// Wrapper around wysteria/common Link object
type Link struct {
	conn *wysteriaClient
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
