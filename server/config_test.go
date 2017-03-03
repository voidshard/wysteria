package main

import (
	"testing"
	"strconv"
)

func TestGetDefaults(t *testing.T) {
	// act
	settings := getDefaults()
	cases := []struct {
		Key string
		Required string
	} {
		// Required minimum value(s) in order to connect and operate
		{"Middleware Driver", settings.MiddlewareSettings.Driver, },
		{"Middleware Host", settings.MiddlewareSettings.Host, },
		{"Middleware Port", strconv.Itoa(settings.MiddlewareSettings.Port), },

		{"Middleware RoutePublic", settings.MiddlewareSettings.RoutePublic, },
		{"Middleware RouteServer", settings.MiddlewareSettings.RouteServer, },
		{"Middleware RouteClient", settings.MiddlewareSettings.RouteClient, },
		{"Middleware RouteInternalServer", settings.MiddlewareSettings.RouteInternalServer, },

		{"Database Driver", settings.DatabaseSettings.Driver, },
		{"Database Host", settings.DatabaseSettings.Host, },
		{"Database Port", strconv.Itoa(settings.DatabaseSettings.Port), },
		{"Database Database", settings.DatabaseSettings.Database, },

		{"Searchbase Driver", settings.SearchbaseSettings.Driver, },
		{"Searchbase Host", settings.SearchbaseSettings.Host, },
		{"Searchbase Port", strconv.Itoa(settings.SearchbaseSettings.Port), },
		{"Searchbase Database", settings.SearchbaseSettings.Database, },
	}

	// assert
	for _, tst := range cases {
		if tst.Required == "" {
			t.Errorf("Require a value to be set for %s", tst.Key)
		}
	}
}
