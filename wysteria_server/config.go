package main

import (
	gcfg "gopkg.in/gcfg.v1"
	"os"
	"errors"
	wdb "wysteria/wysteria_server/database"
	wsb "wysteria/wysteria_server/searchbase"
	wcm "wysteria/wysteria_common/middleware"
)

const (
	default_config  = "wysteria-server.ini"
	default_envvar  = "WYSTERIA_SERVER_INI"
)

var Config configuration

type configuration struct { // forms a universal config
	MiddlewareSettings wcm.MiddlewareSettings
	DatabaseSettings wdb.DatabaseSettings
	SearchbaseSettings wsb.SearchbaseSettings
}

func init () {
	err := readConfig()
	if err != nil {
		setDefaults()
	}
}

func setDefaults() {
	Config.MiddlewareSettings.Driver = "nats"
	Config.MiddlewareSettings.Host = "127.0.0.1"
	Config.MiddlewareSettings.EncryptionKey = ""
	Config.MiddlewareSettings.User = ""
	Config.MiddlewareSettings.Pass = ""
	Config.MiddlewareSettings.Port = 4222
	Config.MiddlewareSettings.RoutePublic = "WYSTERIA.PUBLIC."
	Config.MiddlewareSettings.RouteServer = "WYSTERIA.SERVER."
	Config.MiddlewareSettings.RouteClient = "WYSTERIA.CLIENT."
	Config.MiddlewareSettings.RouteInternalServer = "WYSTERIA.INTERNAL."
	Config.MiddlewareSettings.PemFile = "/path/to/some/nats.pem"

	Config.DatabaseSettings.Driver = "mongo"
	Config.DatabaseSettings.Host = "127.0.0.1"
	Config.DatabaseSettings.Port = 27017
	Config.DatabaseSettings.User = ""
	Config.DatabaseSettings.Pass = ""
	Config.DatabaseSettings.Database = "wysteria_db"
	Config.DatabaseSettings.PemFile = "/path/to/some/mongodb.pem"

	Config.SearchbaseSettings.Driver = "elastic"
	Config.SearchbaseSettings.Host = "127.0.0.1"
	Config.SearchbaseSettings.Port = 9200
	Config.SearchbaseSettings.User = ""
	Config.SearchbaseSettings.Pass = ""
	Config.SearchbaseSettings.Database = "wysteria_idx"
	Config.SearchbaseSettings.PemFile = "/path/to/some/mongodb.pem"
}

func readConfig() error {
	paths := []string {
		default_config,
		os.Getenv(default_envvar),
	}

	for _, path := range paths {
		if path == "" {
			continue
		}

		err := gcfg.ReadFileInto(&Config, path)
		if err == nil {
			return nil
		}
	}
	return errors.New("No config file found to read.")
}
