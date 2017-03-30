package searchends

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

const (
	DRIVER_ELASTIC = "elastic"
	DRIVER_BLEVE   = "bleve"
)

var (
	connectors = map[string]func(*SearchbaseSettings) (Searchbase, error){
		DRIVER_ELASTIC: elastic_connect,
		DRIVER_BLEVE:   bleve_connect,
	}
)

// Return a connect function for the given settings, or err if it can't be found
//
func Connect(settings *SearchbaseSettings) (Searchbase, error) {
	connector, ok := connectors[settings.Driver]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Connector not found for %s", settings.Driver))
	}
	return connector(settings)
}

// a datastore whose primary goal is running search queries rather than long term storage
type Searchbase interface {
	Close() error

	InsertCollection(string, *wyc.Collection) error
	InsertItem(string, *wyc.Item) error
	InsertVersion(string, *wyc.Version) error
	InsertResource(string, *wyc.Resource) error
	InsertLink(string, *wyc.Link) error

	UpdateItem(string, *wyc.Item) error
	UpdateVersion(string, *wyc.Version) error

	// delete data in table/collection with ids
	DeleteCollection(...string) error
	DeleteItem(...string) error
	DeleteVersion(...string) error
	DeleteResource(...string) error
	DeleteLink(...string) error

	// we only ever get IDs back from our search - the DB holds the canonical data
	// Where each takes:
	//  (optional) int - limit results to at most int
	//  (optional) int - from (results page)
	//  ...QueryDesc - description(s) of thing(s) to search for.
	//     results will be returned for any doc matching any QueryDesc
	QueryCollection(int, int, ...*wyc.QueryDesc) ([]string, error)
	QueryItem(int, int, ...*wyc.QueryDesc) ([]string, error)
	QueryVersion(int, int, ...*wyc.QueryDesc) ([]string, error)
	QueryResource(int, int, ...*wyc.QueryDesc) ([]string, error)
	QueryLink(int, int, ...*wyc.QueryDesc) ([]string, error)
}
