package wysteria_client

import (
	common "github.com/voidshard/wysteria/common"
	wcm "github.com/voidshard/wysteria/common/middleware"
	"log"
)

var Config *configuration

type configuration struct { // forms a universal config
	Middleware wcm.MiddlewareSettings
}

// Takes care of loading the client config.
//  (1) Load some form of config
//   Load order:
//    - local .ini file(s) if they are in the cwd
//    - .ini filepath given by wysteria os.Env variable
//    - default values
//
func init() {
	Config = getDefaults()

	config_filepath, err := common.ChooseClientConfig()
	if err == nil {
		cnf := &configuration{}
		err := common.ReadConfig(config_filepath, cnf)
		if err != nil {
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
			Driver: wcm.DriverNats,
			Config: "",
		},
	}
}
