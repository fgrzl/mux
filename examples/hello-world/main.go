package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/fgrzl/mux"
)

func main() {
	router := mux.NewRouter()

	if err := router.Configure(func(router *mux.Router) {
		router.GET("/", func(c mux.RouteContext) {
			c.OK("Hello, World!")
		}).OperationID("helloRoot")

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
		}).OperationID("helloName")
	}); err != nil {
		panic(err)
	}

	server := mux.NewServer(":8080", router)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := server.Listen(ctx); err != nil {
		panic(err)
	}
}
