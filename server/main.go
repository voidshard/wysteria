/*
Main server entry point.

Data flow through interfaces during a client request

	[client code] wysteria/client
		|
		| * ---------------------------- enter middleware ---------------------------- *
		|
		V
 	[client transport code]  Interface: EndpointClient   wysteria/common/middleware.go
		|
       === \\ network //
		|
		V
	[server transport code]  Interface: ServerHandler    wysteria/common/middleware.go
		|
		| * ---------------------------- exit middleware ---------------------------- *
		|
		V
	[middleware shim layer]  Interface: EndpointServer  wysteria/server/instrumentation_middleware.go
		|
		|
		V
	[main server]  Interface: EndpointServer  wysteria/server/server.go
		|
		|
		|   [Nb. What order this goes to the searchbase / database (or even if it does) depends on the operation]
		+------/--------+----------------------------------------------------------------------------------+
						|                                                                                  |
						V                                                                                  |
			[database shim layer] Interface: Database wysteria/server/instrumentation_database.go          |
						|                                                                                  V
						|      [searchbase shim layer] Interface: Searchbase wysteria/server/instrumentation_searchbase.go
						V                                                                                  |
			[database layer] Interface: Database wysteria/server/database/database.go                      |
						|                                                                                  V
						|      [searchbase layer] Interface: Searchbase wysteria/server/searchbase/searchbase.go
						V                                                                                  |
			[database shim layer] Interface: Database wysteria/server/instrumentation_database.go 		   |
						|                                                                                  V
                        |      [searchbase shim layer] Interface: Searchbase wysteria/server/instrumentation_searchbase.go
						|                                                                                  |
						|                                                                                  |
		+---------------+----------------------------------------------------------------------------------+
		|
		V
    [middleware shim layer]  Interface: EndpointServer  wysteria/server/instrumentation_middleware.go
		|
		|
		| * ---------------------------- enter middleware ---------------------------- *
		|
		V
	[server transport code]  Interface: ServerHandler    wysteria/common/middleware.go
		|
       === \\ network //
		|
		V
    [client transport code]  Interface: EndpointClient   wysteria/common/middleware.go
		|
		| * ---------------------------- exit middleware ---------------------------- *
		|
		V
 	[client code] wysteria/client

*/
package main

import (
	"os"
	"time"
	kp "gopkg.in/alecthomas/kingpin.v2"
	wsi "github.com/voidshard/wysteria/server/instrumentation"
)

func main() {
	app := kp.New("Wysteria", "Asset management system")
	verbose := app.Flag("verbose", "Log server output to shell").Short('v').Bool()
	config := app.Flag("config", "Explicitly set a config file").Short('c').String()
	graceful := app.Flag("shutdown", "Time allotted for graceful shutdown").Default("3s").Duration()
	kp.MustParse(app.Parse(os.Args[1:]))

	cfg := loadConfig(*config)

	if *verbose {
		cfg.Instrumentation["wys_verbose"] = &wsi.Settings{Driver: wsi.DriverStdout}
	}

	server := WysteriaServer{
		// Allocate 3 seconds for the server connections to be severed when shutting down
		GracefulShutdownTime: *graceful * time.Second,

		// Pass in all config settings
		settings: cfg,
	}

	// Connect to all required endpoints, die in the event of any failures
	err := server.Run()
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()
}
