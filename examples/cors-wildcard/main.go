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
	// Create a new router
	router := mux.NewRouter()

	// Configure CORS with wildcard support
	// This allows requests from any subdomain of example.com
	mux.UseCORS(router,
		// Allow the main domain
		mux.WithAllowedOrigins("https://example.com"),

		// Allow all subdomains using wildcard pattern
		mux.WithOriginWildcard("*.example.com"),

		// Configure other CORS options
		mux.WithAllowedMethods("GET", "POST", "PUT", "DELETE"),
		mux.WithAllowedHeaders("Authorization", "Content-Type"),
		mux.WithCredentials(true),
		mux.WithMaxAge(3600), // Cache preflight for 1 hour
	)

	if err := router.Configure(func(router *mux.Router) {
		// Add a simple endpoint
		router.GET("/api/users", func(c mux.RouteContext) {
			c.OK(map[string]any{
				"users": []string{"alice", "bob", "charlie"},
			})
		})

		// Add another endpoint
		router.POST("/api/users", func(c mux.RouteContext) {
			c.Created(map[string]any{
				"id":      123,
				"message": "User created",
			})
		})
	}); err != nil {
		panic(err)
	}

	// Start the server
	fmt.Println("🚀 Server starting on http://localhost:8080")
	fmt.Println("📡 CORS enabled for:")
	fmt.Println("   - https://example.com")
	fmt.Println("   - https://*.example.com (any subdomain)")
	fmt.Println("")
	fmt.Println("✅ These origins will work:")
	fmt.Println("   - https://api.example.com")
	fmt.Println("   - https://www.example.com")
	fmt.Println("   - https://staging.api.example.com")
	fmt.Println("")
	fmt.Println("❌ These origins will be rejected:")
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
