package main

import (
	"fmt"
	common "github.com/voidshard/wysteria/common"
	wcm "github.com/voidshard/wysteria/common/middleware"
	wdb "github.com/voidshard/wysteria/server/database"
	wsi "github.com/voidshard/wysteria/server/instrumentation"
	wsb "github.com/voidshard/wysteria/server/searchbase"
	"log"
	"os"
	"path/filepath"
)

var Config *configuration

type configuration struct {
	Database        wdb.Settings
	Searchbase      wsb.Settings
	Middleware      wcm.Settings
	Health          wsi.WebserverConfig
	Instrumentation map[string]*wsi.Settings
}

// Load the server side configuration from somewhere.
//   Load order:
//    - local .ini file(s) if they are in the cwd, if found
//    - .ini filepath given by wysteria os.Env variable, if set
//    - default values
//
func init() {
	Config = makeDefaults()
	configSet := false
	configFilepath, err := common.ChooseServerConfig()
	if err == nil {
		cnf := &configuration{}
		err := common.ReadConfig(configFilepath, cnf)
		log.Println("Attempting to read", configFilepath, cnf, err)
		if err != nil {
			log.Println(fmt.Sprintf("Unable to read config '%s' %s", configFilepath, err))
		} else {
			configSet = true
			Config = cnf
		}
	}
	if !configSet {
		log.Println("WARNING: No config found, using OS temporary folders for storage.")
	}
}

// Get the default settings.
//
func makeDefaults() *configuration {
	return &configuration{
		wdb.Settings{
			Driver:   wdb.DriverBolt,
			Database: filepath.Join(os.TempDir(), "wysteria_db"),
		},

		wsb.Settings{
			Driver:   wsb.DriverBleve,
			Host:     "",
			Port:     0,
			User:     "",
			Pass:     "",
			Database: filepath.Join(os.TempDir(), "wysteria_sb"),
			PemFile:  "",
		},

		wcm.Settings{
			Driver:       wcm.DriverNats,
			Config:       "",
			SSLEnableTLS: false,
			SSLVerify:    false,
			SSLCert:      "",
			SSLKey:       "",
		},

		wsi.WebserverConfig{
			Port:           8150,
			EndpointHealth: "/health",
		},

		map[string]*wsi.Settings{
			"default": &wsi.Settings{
				Driver:   wsi.DriverLogfile,
				Location: filepath.Join(os.TempDir(), "wysteria_logs"),
				Target:   "out.log",
			},
		},
	}
}
