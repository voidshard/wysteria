package main

import (
	"fmt"
	common "github.com/voidshard/wysteria/common"
	wcm "github.com/voidshard/wysteria/common/middleware"
	wdb "github.com/voidshard/wysteria/server/database"
	wsb "github.com/voidshard/wysteria/server/searchbase"
	"log"
	"os"
	"path/filepath"
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
}

// Get the default settings.
//
func getDefaults() *configuration {
	return &configuration{
		wdb.DatabaseSettings{
			Driver:   wdb.DriverBolt,
			Database: filepath.Join(os.TempDir(), "wysteria_db"),
		},

		wsb.SearchbaseSettings{
			Driver:   wsb.DriverBleve,
			Host:     "",
			Port:     0,
			User:     "",
			Pass:     "",
			Database: filepath.Join(os.TempDir(), "wysteria_sb"),
			PemFile:  "",
		},
		wcm.MiddlewareSettings{
			Driver: wcm.DriverNats,
			Config: "",
		},
	}
}
