/*
Here we define the interfaces that are required of a client side & server side middleware.

Offers interfaces for EndpointClient, EndpointServer and util funcs for creating new middleware
endpoints.
*/

package middleware

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
)

const (
	// Drivers known to our middleware
	DriverGrpc = "grpc"
	DriverNats = "nats"
)

var (
	client_endpoints = map[string]func() EndpointClient{
		DriverGrpc: newGrpcClient,
		DriverNats: newNatsClient,
	}
	server_endpoints = map[string]func() EndpointServer{
		DriverGrpc: newGrpcServer,
		DriverNats: newNatsServer,
	}
)

// Create a new middleware client using the diver matching the given string
func NewClient(driver string) (EndpointClient, error) {
	spawner, ok := client_endpoints[driver]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Connector not found for %s", driver))
	}
	return spawner(), nil
}

// Create a new middleware server using the diver matching the given string
func NewServer(driver string) (EndpointServer, error) {
	spawner, ok := server_endpoints[driver]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Connector not found for %s", driver))
	}
	return spawner(), nil
}

// Middleware client interface
// The client side needs to implement connecting to the server, and calling the
// appropriate middleware functions to send data across
type EndpointClient interface {
	// Connect to the server, given some url / connection string
	Connect(string) error

	// Close connection(s) to the server
	Close() error


	// Send collection creation request, return new Id
	//   - Collection name is required to be unique among collections
	CreateCollection(string) (string, error)

	// Send item creation request, return new Id
	// Required to include non empty fields for
	//   - parent (collection Id)
	//   - item type
	//   - item variant
	// Item facets are required to include
	//   - grandparent collection name
	CreateItem(*wyc.Item) (string, error)

	// Send version creation request, return new Id & new version number
	// Required to include non empty fields for
	//   - parent (item Id)
	// Version facets are required to include
	//   - grandparent collection name
	//   - parent item type
	//   - parent item variant
	CreateVersion(*wyc.Version) (string, int32, error)

	// Send resource creation request, return new Id
	// Required to include non empty fields for
	//   - parent (version Id)
	CreateResource(*wyc.Resource) (string, error)

	// Send link creation request, return new Id
	// Required to include non empty fields for
	//   - source Id
	//   - destination Id
	CreateLink(*wyc.Link) (string, error)


	// Given the Id, delete some collection
	DeleteCollection(string) error

	// Given the Id, delete some item
	DeleteItem(string) error

	// Given the Id, delete some version
	DeleteVersion(string) error

	// Given the Id, delete some resource
	DeleteResource(string) error


	// Given some list of QueryDescriptions, return matching collections
	FindCollections([]*wyc.QueryDesc) ([]*wyc.Collection, error)

	// Given some list of QueryDescriptions, return matching items
	FindItems([]*wyc.QueryDesc) ([]*wyc.Item, error)

	// Given some list of QueryDescriptions, return matching versions
	FindVersions([]*wyc.QueryDesc) ([]*wyc.Version, error)

	// Given some list of QueryDescriptions, return matching resources
	FindResources([]*wyc.QueryDesc) ([]*wyc.Resource, error)

	// Given some list of QueryDescriptions, return matching links
	FindLinks([]*wyc.QueryDesc) ([]*wyc.Link, error)


	// Given Id of some Item, return version marked as publish
	PublishedVersion(string) (*wyc.Version, error)

	// Given Id of some Version, mark version as publish
	//  - Only one version of a given Item is considered publish at a time
	SetPublishedVersion(string) error


	// Given Version Id update version facets with given facets
	UpdateVersionFacets(string, map[string]string) error

	// Given Item Id update item facets with given facets
	UpdateItemFacets(string, map[string]string) error
}

// The server side middleware needs to handle starting up, listening, shutting down
// and calling the appropriate handlers from the given server handler when listening
type EndpointServer interface {
	// Start up and serve client requests given some config string and
	// a reference to the main wysteria server's available functions
	ListenAndServe(string, ServerHandler) error

	// Time is up, kill everything and shutdown the server, kill all connections
	Shutdown() error
}

// Implemented by the Wysteria Server
//  This is passed into the middleware server side endpoint as the 'ServerHandler'
type ServerHandler interface {
	// All funcs here as documented in EndpointClient. These are simply the entry points
	// for the middlware receiving the request(s) to call down into the main server
	// and business logic.
	CreateCollection(string) (string, error)
	CreateItem(*wyc.Item) (string, error)
	CreateVersion(*wyc.Version) (string, int32, error)
	CreateResource(*wyc.Resource) (string, error)
	CreateLink(*wyc.Link) (string, error)

	DeleteCollection(string) error
	DeleteItem(string) error
	DeleteVersion(string) error
	DeleteResource(string) error

	FindCollections([]*wyc.QueryDesc) ([]*wyc.Collection, error)
	FindItems([]*wyc.QueryDesc) ([]*wyc.Item, error)
	FindVersions([]*wyc.QueryDesc) ([]*wyc.Version, error)
	FindResources([]*wyc.QueryDesc) ([]*wyc.Resource, error)
	FindLinks([]*wyc.QueryDesc) ([]*wyc.Link, error)

	PublishedVersion(string) (*wyc.Version, error)
	SetPublishedVersion(string) error

	UpdateVersionFacets(string, map[string]string) error
	UpdateItemFacets(string, map[string]string) error
}

// Generic middleware settings struct, holding the diver name and the config string.
// Each driver will expect it's own kind of config string, depending on the driver.
type Settings struct {
	Driver string
	Config string
}
