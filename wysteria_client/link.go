package wysteria_client

import (
	wyc "wysteria/wysteria_common"
)

type link struct {
	conn *wysteriaClient
	data wyc.Link
}

func (i *link) Name() string {
	return i.data.Name
}

func (i *link) Id() string {
	return i.data.Id
}

func (i *link) SourceId() string {
	return i.data.Src
}

func (i *link) DestinationId() string {
	return i.data.Dst
}
