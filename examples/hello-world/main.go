package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/fgrzl/mux"
)

func main() {
	// Create a new router
	router := mux.NewRouter()

	// Add a simple hello endpoint
	router.GET("/", func(c mux.RouteContext) {
		c.OK("Hello, World!")
	}).WithOperationID("helloRoot")

	// Add a greeting endpoint with path parameter
	router.GET("/hello/{name}", func(c mux.RouteContext) {
		name, ok := c.Param("name")
		if !ok {
			c.BadRequest("Missing name", "name parameter is required")
			return
		}

		c.OK(map[string]string{
			"message": "Hello, " + name + "!",
			"status":  "success",
		})
	}).WithOperationID("helloName")

	// Start the server with graceful shutdown
	server := mux.NewServer(":8080", router)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := server.Listen(ctx); err != nil {
		panic(err)
	}
}
