package wysteria_client

import (
	"encoding/json"
	"errors"
	"strings"
	wyc "wysteria/wysteria_common"
	wcm "wysteria/wysteria_common/middleware"
)

type wysteriaClient struct {
	SettingsMiddleware wcm.MiddlewareSettings
	middleware         wcm.WysteriaMiddleware
}

func (w *wysteriaClient) Search() *search {
	return &search{
		conn:      w,
		nextQuery: wyc.QueryDesc{},
	}
}

func (w *wysteriaClient) requestData(route string, send, recv interface{}) error {
	if recv == nil {
		recv = []byte{}
	}

	packed, err := json.Marshal(send)
	if err != nil {
		return err
	}

	reply, err := w.middleware.Request(w.SettingsMiddleware.RouteServer+route, packed)
	if err != nil {
		return err
	}

	err = json.Unmarshal(reply, recv)
	if err != nil {
		// Error parsing json, either:
		//  - transport has failed
		//  - the server replied with a string (which isn't valid json ..) but may be a valid message
		//  - something unexpected has happened (read: horribly wrong)
		err_string := string(reply)
		if strings.HasPrefix(err_string, wyc.WYSTERIA_SERVER_ERR) {
			// The server has replied with an error message
			return errors.New(err_string)
		} else if strings.HasPrefix(err_string, wyc.WYSTERIA_SERVER_ACK) {
			// The server has replied with our prearranged "acknowledged" message
			return nil
		}
		return err
	}

	return nil
}

// setup funcs
func New() *wysteriaClient {
	return &wysteriaClient{
		SettingsMiddleware: Config.MiddlewareSettings,
	}
}

func (w *wysteriaClient) Close() {
	w.middleware.Close()
}

func (w *wysteriaClient) Set(settings wcm.MiddlewareSettings) *wysteriaClient {
	w.SettingsMiddleware = settings
	return w
}

func (w *wysteriaClient) Connect() (*wysteriaClient, error) {
	middleware, err := wcm.Connect(w.SettingsMiddleware)
	if err != nil {
		return w, err
	}
	w.middleware = middleware
	return w, nil
}
