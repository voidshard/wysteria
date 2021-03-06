/*
The Database module provides an interface for the saving, retrieving, deleting & updating of wysteria
structs by their Id(s). It doesn't supply any 'find' or 'search' functions - they belong to the Searchbase module.

The aim here is for a database implementation to supply permanent storage & replication facilities without
needing to worry about search, for which a more efficient system may exist.

Technically, there is nothing preventing something from implementing both the search & database interfaces and
performing both, but this divide provides us the flexibility to use different systems for what they're good at.
*/

package database

import (
	"errors"
	"fmt"
	"time"
	wyc "github.com/voidshard/wysteria/common"
)

const (
	// The database driver names that we accept
	DriverMongo = "mongo"
	DriverBolt  = "bolt"
	maxAttempts = 3
)

var (
	// List of known drivers & the functions to start up a connection
	connectors = map[string]func(*Settings) (Database, error){
		DriverMongo: mongoConnect,
		DriverBolt:  boltConnect,
	}
)

// Return a connect function for the given settings, or err if it can't be found
//
func Connect(settings *Settings) (Database, error) {
	connector, ok := connectors[settings.Driver]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Connector not found for %s", settings.Driver))
	}

	// Attempt to connect to the given db, if something goes wrong, we'll re-attempt a few times
	// before quitting. Mostly this helps in docker setups / test situations where a
	// container may not have spun up yet.
	attempts := 0
	for {
		db, err := connector(settings)
		if err != nil {
			attempts += 1
			if attempts >= maxAttempts {
				return db, err
			}
			time.Sleep(1 * time.Second)
			continue
		}
		return db, err
	}
}

// Datastore whose primary goal is long term storage, not searching said data
type Database interface {
	// Given a Version ID, set the Version as "published"
	// That is, of all versions belonging to the parent Item, the version with
	// this ID is marked as the published "official" version
	//  - Must: Ensure at most one Version of a given Item is marked as published
	SetPublished(string) error

	// Given an Item ID, return the ID of the current PublishedVersion version (if any)
	Published(string) (*wyc.Version, error)

	// Insert a collection into the db, return created Id
	//  - Must: Ensure only one collection with given Name
	InsertCollection(*wyc.Collection) (string, error)

	// Insert an item into the db, return created Id
	//  - Must: Ensure only one item with the same Collection (parent Id), Type and Variant
	InsertItem(*wyc.Item) (string, error)

	// Insert a Version into the db, return created Id
	//  - Must: Ensure version number is both set on the obj & returned.
	//  - Must: Number version numbers starting with 1
	//  - Must: Ensure there is at most one version of a given number
	InsertNextVersion(*wyc.Version) (string, int32, error) // Ensure only one version of an Item with a given Number

	// Insert resource into the db, return created Id
	InsertResource(*wyc.Resource) (string, error)

	// Insert link into the db, return created Id
	InsertLink(*wyc.Link) (string, error)

	// Retrieve collections indicated by the given Id(s) from the db
	RetrieveCollection(...string) ([]*wyc.Collection, error)

	// Retrieve items indicated by the given Id(s) from the db
	RetrieveItem(...string) ([]*wyc.Item, error)

	// Retrieve versions indicated by the given Id(s) from the db
	RetrieveVersion(...string) ([]*wyc.Version, error)

	// Retrieve resources indicated by the given Id(s) from the db
	RetrieveResource(...string) ([]*wyc.Resource, error)

	// Retrieve links indicated by the given Id(s) from the db
	RetrieveLink(...string) ([]*wyc.Link, error)

	// Save the updated facets on the given version
	UpdateItem(string, *wyc.Item) error

	// Save the updated facets on the given item
	UpdateVersion(string, *wyc.Version) error

	// Save the updated facets on the given collection
	UpdateCollection(string, *wyc.Collection) error

	// Save the updated facets on the given resource
	UpdateResource(string, *wyc.Resource) error

	// Save the updated facets on the given link
	UpdateLink(string, *wyc.Link) error

	// Delete collection(s) with the given Id(s)
	DeleteCollection(...string) error

	// Delete item(s) with the given Id(s)
	DeleteItem(...string) error

	// Delete version(s) with the given Id(s)
	DeleteVersion(...string) error

	// Delete resource(s) with the given Id(s)
	DeleteResource(...string) error

	// Delete link(s) with the given Id(s)
	DeleteLink(...string) error

	// kill connection to db
	Close() error
}
