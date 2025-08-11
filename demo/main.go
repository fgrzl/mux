package main

import (
	"fmt"

	"github.com/fgrzl/mux"
)

func main() {
	// Create a handler that uses the interface
	handler := func(c mux.RouteContext) {
		c.SetService("database", "mock-db-connection")
		c.SetService("logger", "mock-logger")

		// Retrieve services
		if db, ok := c.GetService("database"); ok {
			fmt.Printf("Database service: %v\n", db)
		}
		if logger, ok := c.GetService("logger"); ok {
			fmt.Printf("Logger service: %v\n", logger)
		}

		// Test response methods
		c.JSON(200, map[string]string{"message": "Hello, World!"})
	}

	// Test the handler
	fmt.Println("Testing RouteContextInterface with Mock Implementation:")
	mux.TestHandlerWithMock(handler)

	// Test error handling
	errorHandler := func(c mux.RouteContext) {
		c.BadRequest("Validation Error", "The request data is invalid")
	}

	fmt.Println("\nTesting error response:")
	mux.TestHandlerWithMock(errorHandler)

	fmt.Println("\n✅ Success! RouteContext is now mockable for unit testing!")
	fmt.Println("Developers can now create mock implementations of RouteContextInterface")
	fmt.Println("to test their handlers without needing actual HTTP requests.")
}
