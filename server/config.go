package main

import (
	common "github.com/voidshard/wysteria/common"
	wcm "github.com/voidshard/wysteria/common/middleware"
	wdb "github.com/voidshard/wysteria/server/database"
	wsb "github.com/voidshard/wysteria/server/searchbase"
	"log"
	"fmt"
)

var Config *configuration



type configuration struct {
	Database   wdb.DatabaseSettings
	Searchbase wsb.SearchbaseSettings
	Middleware wcm.MiddlewareSettings
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
	if err == nil {
		cnf := &configuration{}
		err := common.ReadConfig(config_filepath, cnf)
		log.Println("Attempting to read", config_filepath, cnf, err)
		if err != nil {
			log.Println(fmt.Sprintf("Unable to read config '%s' %s", config_filepath, err))
		} else {
			Config = cnf
		}
	}
	log.Println("Config loaded", Config)
}

// Get the default settings.
//
func getDefaults() *configuration {
	return &configuration{
		wdb.DatabaseSettings {
			Driver: wdb.DRIVER_MONGO,
			Host: "127.0.0.1",
			Port: 27017,
			User: "",
			Pass: "",
			Database: "wysteria_db",
			PemFile: "",
		},

		wsb.SearchbaseSettings {
			Driver: wsb.DRIVER_ELASTIC,
			Host: "127.0.0.1",
			Port: 9200,
			User: "",
			Pass: "",
			Database: "wysteria_idx",
			PemFile: "",
		},
		wcm.MiddlewareSettings {
			Driver: wcm.DRIVER_NATS,
			Config: "",
		},
	}
}
