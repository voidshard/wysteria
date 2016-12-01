package wysteria_client

import (
	gcfg "gopkg.in/gcfg.v1"
	"os"
	"errors"
	wcm "wysteria/wysteria_common/middleware"
)

const (
	default_config  = "wysteria-client.ini"
	default_envvar  = "WYSTERIA_CLIENT_INI"
)

var Config configuration

type configuration struct { // forms a universal config
	MiddlewareSettings wcm.MiddlewareSettings
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
	Config.MiddlewareSettings.RouteInternalServer = ""
	Config.MiddlewareSettings.PemFile = "/path/to/some/nats.pem"
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
