package middleware

import (
	"errors"
	"fmt"
	"time"
	wyc "github.com/voidshard/wysteria/common"
)

const (
	DRIVER_GRPC = "grpc"
)

var (
	Timeout    = time.Second * 30
	client_endpoints = map[string]func() EndpointClient{
		DRIVER_GRPC: newGrpcClient,
	}
	server_endpoints = map[string]func() EndpointServer {
		DRIVER_GRPC: newGrpcServer,
	}

)

func NewClient(driver string) (EndpointClient, error) {
	spawner, ok := client_endpoints[driver]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Connector not found for %s", driver))
	}
	return spawner(), nil
}

func NewServer(driver string) (EndpointServer, error) {
	spawner, ok := server_endpoints[driver]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Connector not found for %s", driver))
	}
	return spawner(), nil
}

// The client side needs to implement connecting to the server, and calling the
// appropriate middleware functions to send data across
type EndpointClient interface {
	Connect(string) error
	Close() error

	CreateCollection(string) (string, error)
	CreateItem(*wyc.Item) (string, error)
	CreateVersion(*wyc.Version) (string, error)
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

	GetPublishedVersion(string) (*wyc.Version, error)
	PublishVersion(string) error

	UpdateVersionFacets(string, map[string]string) error
	UpdateItemFacets(string, map[string]string) error
}

// The server side middleware needs to handle starting up, listening, shutting down
// and calling the appropriate handlers from the given server handler when listening
type EndpointServer interface {
	// Start up and serve client requests
	ListenAndServe(string, ServerHandler) error

	// We're going to shutdown so refuse any more client connections
	BeginShutdown()

	// You're time is up, kill everything
	Shutdown() error
}

// Implemented by the Wysteria Server
type ServerHandler interface {
	CreateCollection(string) (string, error)
	CreateItem(*wyc.Item) (string, error)
	CreateVersion(*wyc.Version) (string, error)
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

	GetPublishedVersion(string) (*wyc.Version, error)
	PublishVersion(string) error

	UpdateVersionFacets(string, map[string]string) error
	UpdateItemFacets(string, map[string]string) error
}
