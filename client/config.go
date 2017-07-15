package wysteria_client

import (
	common "github.com/voidshard/wysteria/common"
	wcm "github.com/voidshard/wysteria/common/middleware"
	"log"
)

var config *configuration

type configuration struct {
	Middleware wcm.Settings
}

// Takes care of loading the client config.
//   Load order:
//    - local .ini file(s) if they are in the cwd
//    - .ini filepath given by wysteria os.Env variable
//    - default values
//
func init() {
	config = makeDefaults()

	config_filepath, err := common.ChooseClientConfig()
	if err == nil {
		cnf := &configuration{}
		err := common.ReadConfig(config_filepath, cnf)
		if err != nil {
			log.Println("Unable to read config", config_filepath, err)
		} else {
			config = cnf
		}
	}
}

// Get the default settings.
// This naively assumes that all our required services are running on the localhost.
//
func makeDefaults() *configuration {
	return &configuration{
		wcm.Settings{
			Driver: wcm.DriverNats,
			Config: "",
			SSLEnableTLS: false,
			SSLVerify: false,
			SSLCert: "",
			SSLKey: "",
		},
	}
}
