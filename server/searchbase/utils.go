package searchends

import (
	"encoding/base64"
	"strings"
	wyc "github.com/voidshard/wysteria/common"
)

// base64 encodes a string, minus the '=' padding chars at the end.
// Util provided for searchbase implementations that may wish to encode strings to avoid weird chars, spaces
// or other symbols.
func b64encode(path string) string {
	return strings.TrimRight(base64.StdEncoding.EncodeToString([]byte(path)), "=")
}

// Return copy of the given collection
// Util provided for searchbase implementations that may want to mutate values of given documents before storage to
// facilitate searching / matching against later. Avoids use of reflection.
func copyCollection(in *wyc.Collection) *wyc.Collection {
	return &wyc.Collection{
		Id: in.Id,
		Name: in.Name,
	}
}

// Return copy of the given item minus facets
// Util provided for searchbase implementations that may want to mutate values of given documents before storage to
// facilitate searching / matching against later. Avoids use of reflection.
func copyItem(in *wyc.Item) *wyc.Item {
	return &wyc.Item{
		Id: in.Id,
		Parent: in.Parent,
		ItemType: in.ItemType,
		Variant: in.Variant,
		Facets: map[string]string{},
	}
}

// Return copy of given version minus facets
// Util provided for searchbase implementations that may want to mutate values of given documents before storage to
// facilitate searching / matching against later. Avoids use of reflection.
func copyVersion(in *wyc.Version) *wyc.Version {
	return &wyc.Version{
		Id: in.Id,
		Parent: in.Parent,
		Number: in.Number,
		Facets: map[string]string{},
	}
}

// Return copy of given resource
// Util provided for searchbase implementations that may want to mutate values of given documents before storage to
// facilitate searching / matching against later. Avoids use of reflection.
func copyResource(in *wyc.Resource) *wyc.Resource {
	return &wyc.Resource{
		Id: in.Id,
		Parent: in.Parent,
		Name: in.Name,
		ResourceType: in.ResourceType,
		Location: in.Location,
	}
}

// Return copy of given link
// Util provided for searchbase implementations that may want to mutate values of given documents before storage to
// facilitate searching / matching against later. Avoids use of reflection.
func copyLink(in *wyc.Link) *wyc.Link {
	return &wyc.Link{
		Id: in.Id,
		Name: in.Name,
		Src: in.Src,
		Dst: in.Dst,
	}
}
