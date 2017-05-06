package main

import (
	"strconv"
	"testing"
)

func TestGetDefaults(t *testing.T) {
	// act
	settings := makeDefaults()
	cases := []struct {
		Key      string
		Required string
	}{
		// Required minimum value(s) in order to connect and operate
		{"Middleware Driver", settings.Middleware.Driver},
		{"Middleware Host", settings.Middleware.Config},

		{"Database Driver", settings.Database.Driver},
		{"Database Host", settings.Database.Host},
		{"Database Port", strconv.Itoa(settings.Database.Port)},
		{"Database Database", settings.Database.Database},

		{"Searchbase Driver", settings.Searchbase.Driver},
		{"Searchbase Host", settings.Searchbase.Host},
		{"Searchbase Port", strconv.Itoa(settings.Searchbase.Port)},
		{"Searchbase Database", settings.Searchbase.Database},
	}

	// assert
	for _, tst := range cases {
		if tst.Required == "" {
			t.Errorf("Require a value to be set for %s", tst.Key)
		}
	}
}
