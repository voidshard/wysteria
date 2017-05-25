/*
Client package implements the Go wysteria client over some middleware.

Essentially it provides a nicer user interface over the raw middleware that wysteria actually uses
to communicate.
*/

package wysteria_client

import (
	wyc "github.com/voidshard/wysteria/common"
	wcm "github.com/voidshard/wysteria/common/middleware"
)

const (
	defaultSearchLimit = 500
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

// Create a new client and connect to the server
func New() (*Client, error) {
	client := &Client{
		settings: Config,
	}

	middleware, err := wcm.NewClient(client.settings.Middleware.Driver)
	if err != nil {
		return nil, err
	}
	client.middleware = middleware
	return client, middleware.Connect(client.settings.Middleware.Config)
}

// Close any open server connection(s)
func (w *Client) Close() {
	w.middleware.Close()
}
