package main

import (
	wcm "github.com/voidshard/wysteria/common/middleware"
	common "github.com/voidshard/wysteria/common"
	wdb "github.com/voidshard/wysteria/server/database"
	wsb "github.com/voidshard/wysteria/server/searchbase"
	"log"
)

var Config *configuration

type configuration struct { // forms a universal config
	MiddlewareSettings wcm.MiddlewareSettings
	DatabaseSettings   wdb.DatabaseSettings
	SearchbaseSettings wsb.SearchbaseSettings
}

// Key tasks of config init();
//  (1) Load some form of config
//   Load order:
//    - local .ini file(s) if they are in the cwd
//    - .ini filepath given by wysteria os.Env variable
//    - default values
//
func init() {
	Config = getDefaults()

	config_filepath, err := common.ChooseServerConfig()
	if err != nil {
		cnf := &configuration{}
		err := common.ReadConfig(config_filepath, cnf)
		if err == nil {
			log.Println("Unable to read config", config_filepath, err)
		} else {
			Config = cnf
		}
	}
}

// Get the default settings.
// This naively assumes that all our required services are running on the localhost.
//
func getDefaults() *configuration {
	return &configuration{
		wcm.MiddlewareSettings{
			Driver: "nats",
			Host: "127.0.0.1",
			EncryptionKey: "",
			User: "",
			Pass: "",
			Port: 4222,
			RoutePublic: "WYSTERIA.PUBLIC.",
			RouteServer: "WYSTERIA.SERVER.",
			RouteClient: "WYSTERIA.CLIENT.",
			RouteInternalServer: "WYSTERIA.INTERNAL.",
			PemFile: "",
		},

		wdb.DatabaseSettings{
			Driver: "mongo",
			Host: "127.0.0.1",
			Port: 27017,
			User: "",
			Pass: "",
			Database: "wysteria_db",
			PemFile: "",
		},

		wsb.SearchbaseSettings{
			Driver: "elastic",
			Host: "127.0.0.1",
			Port: 9200,
			User: "",
			Pass: "",
			Database: "wysteria_idx",
			PemFile: "",
		},
	}
}
