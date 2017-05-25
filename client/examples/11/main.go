package main

import (
	wysteria "github.com/voidshard/wysteria/client"
	"github.com/voidshard/wysteria/common/middleware"
	"log"
)

func main() {
	// Example 11: Overriding config variables

	_, err := wysteria.New(
		wysteria.Driver(middleware.DriverNats),
		wysteria.Host("nats://localhost:4222"),
	)
	if err != nil {
		panic(err)
	}

	log.Println("Connected!")
}
