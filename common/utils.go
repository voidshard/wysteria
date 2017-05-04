/*
Provides some basic functions find & load config files.
*/

package wysteria_common

import (
	"errors"
	"fmt"
	"gopkg.in/gcfg.v1"
	"os"
)

const (
	defaultServerConfig = "wysteria-server.ini"
	defaultClientConfig = "wysteria-client.ini"
	defaultServerEnvvar = "WYSTERIA_SERVER_INI"
	defaultClientEnvvar = "WYSTERIA_CLIENT_INI"
)

func ChooseServerConfig() (string, error) {
	return chooseConfig(
		defaultServerConfig,
		os.Getenv(defaultServerEnvvar),
	)
}

func ChooseClientConfig() (string, error) {
	return chooseConfig(
		defaultClientConfig,
		os.Getenv(defaultClientEnvvar),
	)
}

// Choose some config file to load
//  The first file we can os.Stat we accept as the (probably) intended config
//
func chooseConfig(paths ...string) (string, error) {
	for _, path := range paths {
		if path == "" {
			continue
		}

		_, err := os.Stat(path)
		if err == nil {
			return path, nil
		}
	}
	return "", errors.New(fmt.Sprintf("Unable to stat any of %s", paths))
}

// Read configuration information from a file
//
func ReadConfig(path string, conf interface{}) error {
	return gcfg.ReadFileInto(conf, path)
}
