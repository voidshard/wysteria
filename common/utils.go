package wysteria_common

import (
	"os"
	"fmt"
	"gopkg.in/gcfg.v1"
	"errors"
	"log"
)

const (
	default_server_config = "wysteria-server.ini"
	default_client_config = "wysteria-client.ini"
	default_server_envvar = "WYSTERIA_SERVER_INI"
	default_client_envvar = "WYSTERIA_CLIENT_INI"
)

func ChooseServerConfig() (string, error) {
	return chooseConfig(
		default_server_config,
		os.Getenv(default_server_envvar),
	)
}

func ChooseClientConfig() (string, error) {
	return chooseConfig(
		default_client_config,
		os.Getenv(default_client_envvar),
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
		log.Println(fmt.Sprintf("Searching for config: %s %v", path, err))
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
