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

		{"Database Driver", settings.Database.Driver},
		{"Database Port", strconv.Itoa(settings.Database.Port)},
		{"Database Database", settings.Database.Database},

		{"Searchbase Driver", settings.Searchbase.Driver},
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
