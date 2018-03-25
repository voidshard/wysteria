package integration

import (
	"os"
	"testing"
	"github.com/fgrid/uuid"
	wyc "github.com/voidshard/wysteria/client"
)

var (
	conf = os.Getenv("WYS_MIDDLEWARE_CLIENT_CONFIG")
	driver = os.Getenv("WYS_MIDDLEWARE_CLIENT_DRIVER")

	sslcert = os.Getenv("WYS_MIDDLEWARE_CLIENT_SSL_CERT")
	sslkey = os.Getenv("WYS_MIDDLEWARE_CLIENT_SSL_KEY")

	sslVerify = os.Getenv("WYS_MIDDLEWARE_SSL_VERIY") == "true"
	sslEnable = os.Getenv("WYS_MIDDLEWARE_SSL_ENABLE") == "true"
)


// Return random uuid4 (as a string)
//
func randomString() string {
	return uuid.NewV4().String()
}

// Create new wysteria client with config info from ENV vars
//
func newClient(t *testing.T) *wyc.Client {
	client, err := wyc.New(
		wyc.Host(conf), wyc.Driver(driver),
		wyc.SSLKey(sslkey), wyc.SSLCert(sslcert), wyc.SSLVerify(sslVerify), wyc.SSLEnable(sslEnable),
	)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	return client
}

// Test that the first map[string]string is a subset of the second map[string]string
//
func facetsContain(subset, superset map[string]string) bool {
	for k, v := range subset {
		if superset[k] != v {
			return false
		}
	}
	return true
}
