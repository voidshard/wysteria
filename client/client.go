/*
Client package implements the Go wysteria client over some middleware.

Essentially it provides a nicer user interface over the raw middleware that wysteria actually uses
to communicate.
*/

package wysteria_client

import (
	wyc "github.com/voidshard/wysteria/common"
	wcm "github.com/voidshard/wysteria/common/middleware"
	"fmt"
	"errors"
)

const (
	defaultSearchLimit = 500

	// These are copied over from github.com/voidshard/wysteria/common for convenience
	//
	FacetRootCollection = "/"
	FacetCollection     = "collection"
	FacetItemType       = "itemtype"
	FacetItemVariant    = "variant"
	FacetLinkType       = "linktype"
	FacetVersionLink    = "version"
	FacetItemLink       = "item"
)

// Client wraps the desired middleware and supplies a more user friendly interface to users
type Client struct {
	settings   *configuration
	middleware wcm.EndpointClient
}

// Start a new search request.
//  This builds up a query (or set of queries) clientside to send to the server.
//  'Find<ObjectType>' functions send this query(ies) and request matching objects
//  of the given <ObjectType>.
//  Since network round trip time is invariably expensive, it's recommended to make
//  few specific queries than many non specific.
func (w *Client) Search(opts ...SearchParam) *Search {
	s := &Search{
		limit:     defaultSearchLimit,
		conn:      w,
		query:     []*wyc.QueryDesc{},
		nextQuery: &wyc.QueryDesc{},
	}
	s.applyOptions(opts...)
	return s
}

type ClientOption func(*Client)

// Set a particular host for the client to connect to (overrides config var).
// The exact format of this depends on the middleware driver being used.
// (See example config for details & examples).
func Host(url string) ClientOption {
	return func(i *Client) {
		i.settings.Middleware.Config = url
	}
}

// Set a particular driver for the client to connect with (overrides config var).
// (See example config for details & examples).
func Driver(name string) ClientOption {
	return func(i *Client) {
		i.settings.Middleware.Driver = name
	}
}

// Set SSL certificate to use
//
func SSLCert(in string) ClientOption {
	return func(i *Client) {
		i.settings.Middleware.SSLCert = in
	}
}

// Set SSL key to use
//
func SSLKey(in string) ClientOption {
	return func(i *Client) {
		i.settings.Middleware.SSLKey = in
	}
}

// Verify SSL certificates on connecting
//  That is, if this is 'false' we'll accept insecure certificates (like self signed certs for example)
//
func SSLVerify(in bool) ClientOption {
	return func(i *Client) {
		i.settings.Middleware.SSLVerify = in
	}
}

// Enable SSL
//
func SSLEnable(in bool) ClientOption {
	return func(i *Client) {
		i.settings.Middleware.SSLEnableTLS = in
	}
}

// Create a new client and connect to the server
func New(opts ...ClientOption) (*Client, error) {
	client := &Client{
		settings: config,
	}

	for _, opt := range opts {
		opt(client)
	}

	middleware, err := wcm.NewClient(client.settings.Middleware.Driver)
	if err != nil {
		return nil, err
	}
	client.middleware = middleware
	return client, middleware.Connect(&client.settings.Middleware)
}

// Collection is a helpful wrapper that looks for a single collection
// with either the name or Id of the given 'identifier' and returns it if found
func (w *Client) Collection(identifier string) (*Collection, error) {
	results, err := w.Search(Id(identifier)).Or(Name(identifier)).FindCollections(Limit(1))
	if err != nil {
		return nil, err
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return nil, errors.New(fmt.Sprintf("Expected 1 result, got %d", len(results)))
}

// Item is a helpful wrapper that looks up an Item by ID and returns it (if found).
func (w *Client) Item(in string) (*Item, error) {
	results, err := w.Search(Id(in)).FindItems(Limit(1))
	if err != nil {
		return nil, err
	}
	if len(results) == 1 {
		return results[0], nil
	}
	return nil, errors.New(fmt.Sprintf("Expected 1 result, got %d", len(results)))
}

// Close any open server connection(s)
func (w *Client) Close() {
	w.middleware.Close()
}

// Definition of an option that can be passed in a Create function call.
type CreateOption func(clientStruct, clientStruct) // func(parent, child)

// Required internal functions to apply our CreateOption(s) below
type clientStruct interface {
	initUserFacets(map[string]string)
}

// Set the initial Facets on the soon to be created child
//  Note that you still cannot overwrite client reserved facets this way
func Facets(in map[string]string) CreateOption {
	return func(parent, child clientStruct) {
		if in != nil {
			child.initUserFacets(in)
		}
	}
}