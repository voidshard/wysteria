package database

import (
	wyc "github.com/voidshard/wysteria/common"
)

const (
	divisor = ":"
)

// Create new collection ID using Parent & Name variables
func NewCollectionId(in *wyc.Collection) string {
	return wyc.NewId("collection", divisor, in.Parent, divisor, in.Name)
}

// Create new item ID using parent type & variant vars
func NewItemId(in *wyc.Item) string {
	return wyc.NewId("item", divisor, in.Parent, divisor, in.ItemType, divisor, in.Variant)
}

// Create new version ID using parent & number
func NewVersionId(in *wyc.Version) string {
	return wyc.NewId("version", divisor, in.Parent, divisor, in.Number)
}

// Create new resource ID using parent, name, type and location
func NewResourceId(in *wyc.Resource) string {
	return wyc.NewId("resource", divisor, in.Parent, divisor, in.Name, divisor, in.ResourceType, divisor, in.Location)
}

// Create new link ID using name, src and dst
func NewLinkId(in *wyc.Link) string {
	return wyc.NewId("link", divisor, in.Name, divisor, in.Src, divisor, in.Dst)
}
