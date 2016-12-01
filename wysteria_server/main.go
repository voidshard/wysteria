package main

import (
	"time"
)

func main() {
	server := WysteriaServer{
		// Allocate 3 seconds for the server connections to be severed when shutting down
		GracefulShutdownTime: time.Second * 3,

		// Pass in all config settings
		SettingsMiddleware: Config.MiddlewareSettings,
		SettingsBackend:    Config.DatabaseSettings,
	}

	// Connect to all required endpoints, die in the event of any failures
	err := server.Connect()
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	// Start processing requests
	server.Run()
}
