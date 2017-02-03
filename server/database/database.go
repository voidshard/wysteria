package database

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

var (
	connectors = map[string]func(*DatabaseSettings) (Database, error){
		"mongo": mongo_connect,
	}
)

func Connect(settings *DatabaseSettings) (Database, error) {
	connector, ok := connectors[settings.Driver]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Connector not found for %s", settings.Driver))
	}
	return connector(settings)
}

// datastore whose primary goal is long term storage, not searching said data
type Database interface {
	InsertCollection(string, wyc.Collection) error      // Ensure only one collection with given Name
	InsertItem(string, wyc.Item) error                  // Ensure only one item with the same Collection, Type and Variant
	InsertNextVersion(string, wyc.Version) (int, error) // Ensure only one version of an Item with a given Number
	InsertResource(string, wyc.Resource) error
	InsertLink(string, wyc.Link) error

	RetrieveCollection(...string) ([]wyc.Collection, error)
	RetrieveCollectionByName(...string) ([]wyc.Collection, error)
	RetrieveItem(...string) ([]wyc.Item, error)
	RetrieveVersion(...string) ([]wyc.Version, error)
	RetrieveResource(...string) ([]wyc.Resource, error)
	RetrieveLink(...string) ([]wyc.Link, error)

	// We only update the facets on Items and Versions
	UpdateItem(string, wyc.Item) error
	UpdateVersion(string, wyc.Version) error

	// delete data in table/collection with ids
	DeleteCollection(...string) error
	DeleteItem(...string) error
	DeleteVersion(...string) error
	DeleteResource(...string) error
	DeleteLink(...string) error

	Close() error // kill connection to db
}
