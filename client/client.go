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
	// Facet names the server expects us to populate on item & version creation requests
	FacetCollection = "collection"
	FacetItemType = "itemtype"
	FacetItemVariant = "variant"
)

// Client wraps the desired middleware and supplies a more user friendly interface to users
type wysteriaClient struct {
	settings   *configuration
	middleware wcm.EndpointClient
}

// Start a new search request.
//  This builds up a query (or set of queries) clientside to send to the server.
//  'Find<ObjectType>' functions send this query(ies) and request matching objects
//  of the given <ObjectType>.
//  Since network round trip time is invariably expensive, it's recommended to make
//  few specific queries than many non specific.
//  ToDo: Implement Limit & Page settings
func (w *wysteriaClient) Search() *search {
	return &search{
		conn:      w,
		query:     []*wyc.QueryDesc{},
		nextQuery: &wyc.QueryDesc{},
	}
}

// Create a new client and connect to the server
func New() (*wysteriaClient, error) {
	client := &wysteriaClient{
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
func (w *wysteriaClient) Close() {
	w.middleware.Close()
}
