/*
These structs are what is passed back and forth when encoding to and from json.
That is, they are used in the communication protocol(s) only to transport rpc data to and from clients.

This file is passed to ffjson to create the encode & decode funcs
  wysteria/common/middleware/ffjson_structs/structs_ffjson.go
Using
  ffjson wysteria/common/middleware/ffjson_structs/structs.go

Docs: https://github.com/pquerna/ffjson
*/

package ffjson_structs

import (
	wyc "github.com/voidshard/wysteria/common"
)

// Sent for a Collection creation request (we only need the name, since the
// id is created server side
type CreateReqCollection struct {
	Name string
}

// Sent for Item creation request
type CreateReqItem struct {
	Item wyc.Item
}

// Sent for Version creation request
type CreateReqVersion struct {
	Version wyc.Version
}

// Sent for Resource creation request
type CreateReqResource struct {
	Resource wyc.Resource
}

// Sent for Link creation request
type CreateReqLink struct {
	Link wyc.Link
}

// Every create request except Version has the same reply - an Id and an err string
type CreateReply struct {
	Error string
	Id    string
}

// Create Version includes the newly created Id, the err string and the Version
// number of the new version
type CreateReplyVersion struct {
	Error   string
	Version int32
	Id      string
}

// Sent for a deletion request
type DeleteReq struct {
	Id string
}

// Sent for a deletion request reply
type DeleteReply struct {
	Error string
}

// All find requests are the same, a list of QueryDescription objects
type FindReq struct {
	Query []wyc.QueryDesc
}

// Find Collection reply
type FindReplyCollection struct {
	All   []wyc.Collection
	Error string
}

// Find Item reply
type FindReplyItem struct {
	All   []wyc.Item
	Error string
}

// Find Item reply
type FindReplyVersion struct {
	All   []wyc.Version
	Error string
}

// Find Resource reply
type FindReplyResource struct {
	All   []wyc.Resource
	Error string
}

// Find Link reply
type FindReplyLink struct {
	All   []wyc.Link
	Error string
}

// Whether we are getting the published version or setting, we only send an Id.
// If we're publishing a version, we send the Id of the Version
// If we're getting the published version, we send the Id of the parent Item
type PublishedReq struct {
	Id string
}

// A request to get the published version returns the Version and and err string
type GetPublishedReply struct {
	Error   string
	Version wyc.Version
}

// A request to set a version as published returns an error string indicating if it worked or not
type SetPublishedReply struct {
	Error string
}

// Update facets request (for Items or Versions) both send the Id and the Facets
// to update.
type UpdateFacetsReq struct {
	Id     string
	Facets map[string]string
}

// Update facets only returns an error string indicating if it worked or not
type UpdateFacetsReply struct {
	Error  string
}
