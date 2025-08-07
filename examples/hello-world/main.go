package main

import (
	"net/http"

	"github.com/fgrzl/mux"
)

func main() {
	// Create a new router
	router := mux.NewRouter()

	// Add a simple hello endpoint
	router.GET("/", func(c *mux.RouteContext) {
		c.OK("Hello, World!")
	})

	// Add a greeting endpoint with path parameter
	router.GET("/hello/{name}", func(c *mux.RouteContext) {
		name, ok := c.Param("name")
		if !ok {
			c.BadRequest("Missing name", "name parameter is required")
			return
		}

		c.OK(map[string]string{
			"message": "Hello, " + name + "!",
			"status":  "success",
		})
	})

	// Start the server
	http.ListenAndServe(":8080", router)
}
