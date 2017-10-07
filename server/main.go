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
