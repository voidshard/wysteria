package database

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

const (
	DRIVER_MONGO = "mongo"
)

var (
	connectors = map[string]func(*DatabaseSettings) (Database, error){
		DRIVER_MONGO: mongo_connect,
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
	// Given a Version ID, set the Version as "published"
	// That is, of all versions belonging to the parent Item, the version with
	// this ID is marked as the published "official" version
	SetPublished(string) error

	// Given an Item ID, return the ID of the current Published version (if any)
	GetPublished(string) (*wyc.Version, error)

	InsertCollection(string, *wyc.Collection) error      // Ensure only one collection with given Name
	InsertItem(string, *wyc.Item) error                  // Ensure only one item with the same Collection, Type and Variant
	InsertNextVersion(string, *wyc.Version) (int32, error) // Ensure only one version of an Item with a given Number
	InsertResource(string, *wyc.Resource) error
	InsertLink(string, *wyc.Link) error

	RetrieveCollection(...string) ([]*wyc.Collection, error)
	RetrieveItem(...string) ([]*wyc.Item, error)
	RetrieveVersion(...string) ([]*wyc.Version, error)
	RetrieveResource(...string) ([]*wyc.Resource, error)
	RetrieveLink(...string) ([]*wyc.Link, error)

	// We only update the facets on Items and Versions
	UpdateItem(string, *wyc.Item) error
	UpdateVersion(string, *wyc.Version) error

	// delete_by_id data in table/collection with ids
	DeleteCollection(...string) error
	DeleteItem(...string) error
	DeleteVersion(...string) error
	DeleteResource(...string) error
	DeleteLink(...string) error

	Close() error // kill connection to db
}
