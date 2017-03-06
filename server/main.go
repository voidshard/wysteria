package main

import (
	"time"
)

func main() {
	server := WysteriaServer{
		// Allocate 3 seconds for the server connections to be severed when shutting down
		GracefulShutdownTime: time.Second * 3,

		// Pass in all config settings
		settings: Config,
	}

	// Connect to all required endpoints, die in the event of any failures
	err := server.Run()
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()
}
