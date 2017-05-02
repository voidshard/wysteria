package wysteria_client

import (
	wyc "github.com/voidshard/wysteria/common"
	wcm "github.com/voidshard/wysteria/common/middleware"
)

const (
	FACET_COLLECTION = "collection"
	FACET_ITEM_TYPE = "itemtype"
	FACET_ITEM_VARIANT = "variant"
)

type wysteriaClient struct {
	settings   *configuration
	middleware wcm.EndpointClient
}

func (w *wysteriaClient) Search() *search {
	return &search{
		conn:      w,
		query:     []*wyc.QueryDesc{},
		nextQuery: &wyc.QueryDesc{},
	}
}

// setup funcs
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

func (w *wysteriaClient) Close() {
	w.middleware.Close()
}
