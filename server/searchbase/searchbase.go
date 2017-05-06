/*
Searchbase module

Provides interface and util functions for a 'searchbase' - a database whose main concern is the searching of given
data for quick retrieval of matching IDs. Since we need to support the updating of searchable facets we also
include 'update' functions for Items & Versions.

Note that we never return fields from here *except* for Id(s), so internally implementations are free to store the given
documents in whatever way makes them most efficient for searching.

Due to the nature of us having disjointed search & data stores the data held in a searchbase should not be considered
completely in-sync with the database (canonical) version. But this replication lag between when something is created
in the database and is searchable via the searchbase should be minimal when everything is running sweetly.
*/

package searchends

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

const (
	DriverElastic = "elastic"
	DriverBleve   = "bleve"
)

var (
	connectors = map[string]func(*Settings) (Searchbase, error){
		DriverElastic: elasticConnect,
		DriverBleve:   bleveConnect,
	}
)

// Return a connect function for the given settings, or err if it can't be found
func Connect(settings *Settings) (Searchbase, error) {
	connector, ok := connectors[settings.Driver]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Connector not found for %s", settings.Driver))
	}
	return connector(settings)
}

// Interface to a data store whose primary goal is running search queries rather than storage.
type Searchbase interface {
	Close() error

	InsertCollection(string, *wyc.Collection) error
	InsertItem(string, *wyc.Item) error
	InsertVersion(string, *wyc.Version) error
	InsertResource(string, *wyc.Resource) error
	InsertLink(string, *wyc.Link) error

	UpdateItem(string, *wyc.Item) error
	UpdateVersion(string, *wyc.Version) error


	// Delete collection search data by Id(s)
	//  Note that deleting something and making it unavailable to search for effectively means
	//  that wysteria will never return the data for it, even if it still existed in the database.
	DeleteCollection(...string) error

	// Delete item search data by Id(s)
	DeleteItem(...string) error

	// Delete version search data by Id(s)
	DeleteVersion(...string) error

	// Delete resource search data by Id(s)
	DeleteResource(...string) error

	// Delete link search data by Id(s)
	DeleteLink(...string) error


	// Query functions
	// we only ever get IDs back from our search - the db holds the canonical data
	// Where each takes:
	//  int - limit results to at most int (where 0 indicates there is no limit)
	//  int - from (results page)
	//  ...QueryDesc - description(s) of thing(s) to search for (required)
	//     Ids will be returned for any doc matching all of the fields of any of the given QueryDesc
	QueryCollection(int, int, ...*wyc.QueryDesc) ([]string, error)
	QueryItem(int, int, ...*wyc.QueryDesc) ([]string, error)
	QueryVersion(int, int, ...*wyc.QueryDesc) ([]string, error)
	QueryResource(int, int, ...*wyc.QueryDesc) ([]string, error)
	QueryLink(int, int, ...*wyc.QueryDesc) ([]string, error)
}
