package main

import (
	"fmt"
	"net/http"

	"github.com/fgrzl/mux"
)

func main() {
	router := mux.NewRouter()

	// 301 - Moved Permanently (old URL no longer valid)
	router.GET("/old-api", func(c mux.RouteContext) {
		c.MovedPermanently("/api/v2")
	})

	// 302 - Found (temporary redirect, most common)
	router.GET("/login", func(c mux.RouteContext) {
		c.Found("/auth/login")
	})

	// 303 - See Other (POST -> GET redirect pattern)
	router.POST("/submit", func(c mux.RouteContext) {
		// After processing POST, redirect to GET result page
		c.SeeOther("/result?id=123")
	})

	// 307 - Temporary Redirect (preserves request method)
	router.POST("/api/v1/users", func(c mux.RouteContext) {
		// Redirect POST request to new endpoint, keeping POST method
		c.TemporaryRedirect("/api/v2/users")
	})

	// 308 - Permanent Redirect (preserves request method)
	router.POST("/old-webhook", func(c mux.RouteContext) {
		c.PermanentRedirect("/webhooks/v2")
	})

	// Example result pages
	router.GET("/auth/login", func(c mux.RouteContext) {
		c.HTML(http.StatusOK, "<h1>Login Page</h1>")
	})

	router.GET("/result", func(c mux.RouteContext) {
		id, _ := c.Query().String("id")
		c.HTML(http.StatusOK, fmt.Sprintf("<h1>Result Page</h1><p>ID: %s</p>", id))
	})

	router.GET("/api/v2", func(c mux.RouteContext) {
		c.JSON(http.StatusOK, map[string]string{"message": "New API endpoint"})
	})

	fmt.Println("Redirect Examples Server")
	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("")
	fmt.Println("Try these endpoints:")
	fmt.Println("  GET  /old-api       -> 301 redirect")
	fmt.Println("  GET  /login         -> 302 redirect")
	fmt.Println("  POST /submit        -> 303 redirect (POST to GET)")
	fmt.Println("  POST /api/v1/users  -> 307 redirect (preserves POST)")
	fmt.Println("  POST /old-webhook   -> 308 redirect (preserves POST)")
	fmt.Println("")

	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}
