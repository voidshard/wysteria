package ffjson_structs

import (
	wyc "github.com/voidshard/wysteria/common"
)

//
// These structs are passed to ffjson to create the encode & decode funcs
//   wysteria/common/middleware/ffjson_structs/structs_ffjson.go
// Using
//   ffjson wysteria/common/middleware/ffjson_structs/structs.go
//
// Docs: https://github.com/pquerna/ffjson
//

type CreateReqCollection struct {
	Name string
}

type CreateReqItem struct {
	Item wyc.Item
}

type CreateReqVersion struct {
	Version wyc.Version
}

type CreateReqResource struct {
	Resource wyc.Resource
}

type CreateReqLink struct {
	Link wyc.Link
}

type CreateReply struct {
	Error string
	Id string
}

type CreateReplyVersion struct {
	Error string
	Version int32
	Id string
}

type DeleteReq struct {
	Id string
}

type DeleteReply struct {
	Error string
}

type FindReq struct {
	Query []wyc.QueryDesc
}

type FindReplyCollection struct {
	All []wyc.Collection
	Error string
}

type FindReplyItem struct {
	All []wyc.Item
	Error string
}

type FindReplyVersion struct {
	All []wyc.Version
	Error string
}

type FindReplyResource struct {
	All []wyc.Resource
	Error string
}

type FindReplyLink struct {
	All []wyc.Link
	Error string
}

type PublishedReq struct {
	Id string
}

type GetPublishedReply struct {
	Error string
	Version wyc.Version
}

type SetPublishedReply struct {
	Error string
}

type UpdateFacetsReq struct {
	Id string
	Facets map[string]string
}

type UpdateFacetsReply struct {
	Id string
	Error string
	Facets map[string]string
}
