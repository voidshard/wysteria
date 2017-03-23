package wysteria_client

import (
	wyc "github.com/voidshard/wysteria/common"
)

type Link struct {
	conn *wysteriaClient
	data *wyc.Link
}

func (i *Link) Name() string {
	return i.data.Name
}

func (i *Link) Id() string {
	return i.data.Id
}

func (i *Link) SourceId() string {
	return i.data.Src
}

func (i *Link) DestinationId() string {
	return i.data.Dst
}
