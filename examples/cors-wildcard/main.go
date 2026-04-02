package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fgrzl/mux"
)

func main() {
	router := mux.NewRouter()

	mux.UseCORS(router,
		mux.WithCORSAllowedOrigins("https://example.com"),
		mux.WithCORSOriginWildcard("*.example.com"),
		mux.WithCORSAllowedMethods("GET", "POST", "PUT", "DELETE"),
		mux.WithCORSAllowedHeaders("Authorization", "Content-Type"),
		mux.WithCORSCredentials(true),
		mux.WithCORSMaxAge(3600), // Cache preflight for 1 hour
	)

	if err := router.Configure(func(router *mux.Router) {
		router.GET("/api/users", func(c mux.RouteContext) {
			c.OK(map[string]any{
				"users": []string{"alice", "bob", "charlie"},
			})
		})

		router.POST("/api/users", func(c mux.RouteContext) {
			c.Created(map[string]any{
				"id":      123,
				"message": "User created",
			})
		})
	}); err != nil {
		panic(err)
	}

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("CORS enabled for:")
	fmt.Println("   - https://example.com")
	fmt.Println("   - https://*.example.com (any subdomain)")
	fmt.Println("")
	fmt.Println("These origins will work:")
	fmt.Println("   - https://api.example.com")
	fmt.Println("   - https://www.example.com")
	fmt.Println("   - https://staging.api.example.com")
	fmt.Println("")
	fmt.Println("These origins will be rejected:")
	fmt.Println("   - https://evil.com")
	fmt.Println("   - https://notexample.com")
	fmt.Println("")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	server := mux.NewServer(":8080", router)
	if err := server.Listen(ctx); err != nil {
		panic(err)
	}
}
