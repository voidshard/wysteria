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
	FacetRootCollection = "/"
	FacetCollection     = "collection"
	FacetItemType       = "itemtype"
	FacetItemVariant    = "variant"
	FacetLinkType       = "linktype"

	ValueLinkTypeItem = "item"
	ValueLinkTypeVersion = "version"

	defaultServerConfig = "wysteria-server.ini"
	defaultClientConfig = "wysteria-client.ini"
	defaultServerEnvvar = "WYSTERIA_SERVER_INI"
	defaultClientEnvvar = "WYSTERIA_CLIENT_INI"

	ErrorInvalid = "invalid-input"  // the input was not valid
	ErrorAlreadyExists = "already-exists"  // the input object can't be created - it exists already
	ErrorIllegal = "illegal-operation"  // the operation is not permitted
	ErrorNotFound = "not-found"  // an explicitly given Id was not found
	ErrorNotServing = "operation-rejected" // the server is not serving requests (ie. shutting down / maintenance)
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
