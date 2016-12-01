package searchends

import (
	"fmt"
	"errors"
	wyc "wysteria/wysteria_common"
)

var (
	connectors = map[string] func(SearchbaseSettings) (Searchbase, error) {
		"elastic": elastic_connect,
	}
)

func Connect(settings SearchbaseSettings) (Searchbase, error) {
	connector, ok := connectors[settings.Driver]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Connector not found for %s", settings.Driver))
	}
	return connector(settings)
}

// a datastore whose primary goal is running search queries rather than long term storage
type Searchbase interface {
	Close () error

	InsertItem (string, wyc.Item) error
	InsertVersion (string, wyc.Version) error
	InsertFileResource (string, wyc.FileResource) error
	InsertLink (string, wyc.Link) error

	UpdateItem (string, wyc.Item) error
	UpdateVersion (string, wyc.Version) error

	// delete data in table/collection with ids
	DeleteItem (...string) error
	DeleteVersion (...string) error
	DeleteFileResource (...string) error
	DeleteLink (...string) error

	// we only ever get IDs back from our search - the DB holds the canonical data
	// Where each takes:
	//  (optional) string - field to sort results by
	//  (optional) bool - sort results ascending order
	//  (optional) int - limit results to at most int
	//  ...QueryDesc - description(s) of thing(s) to search for.
	//     results will be returned for any doc matching any QueryDesc
	QueryItem (string, bool, int, ...wyc.QueryDesc) ([]string, error)
	QueryVersion (string, bool, int, ...wyc.QueryDesc) ([]string, error)
	QueryFileResource (string, bool, int, ...wyc.QueryDesc) ([]string, error)
	QueryLink (string, bool, int, ...wyc.QueryDesc) ([]string, error)
}
