package mux

import (
	"net/http"
	"strings"
	"time"
)

func (r *Router) UseAuthentication(options *AuthenticationOptions) {
	r.middleware = append(r.middleware, &authenticationMiddleware{options: options})
}

type AuthenticationOptions struct {
	Validate               func(token string) bool
	ValidateSession        func(sessionID string) (bool, error)            // Validates session from DB or cache
	GetSessionDetails      func(sessionID string) (*SessionDetails, error) // Fetches session info like IP and User-Agent
	GetSessionCreationTime func(sessionID string) time.Time                // Retrieves session creation time
}

type authenticationMiddleware struct {
	options *AuthenticationOptions
}

type SessionDetails struct {
	IP        string
	UserAgent string
	CreatedAt time.Time
}

func (m *authenticationMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	r := c.Request

	// Check if the request URL allows anonymous access (e.g., "/public")
	if c.Options.AllowAnonymous {
		next(c)
		return
	}

	// Try Bearer token authentication first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if m.isValidBearerToken(token) {
			next(c) // Token is valid, pass to the next handler
			return
		}
		c.Unauthorized()
		return
	}

	// Fall back to cookie-based authentication if no Bearer token is present

	cookie, err := r.Cookie(GetAppSessionCookieName())

	if err == nil && cookie != nil && m.isValidSessionCookie(cookie.Value, r, c) {
		// If the session cookie is valid, pass the request along to the next handler
		next(c)
		return
	}

	// If neither Bearer token nor valid session cookie is found, return Unauthorized
	c.Unauthorized()
}

// Validates the Bearer token.
func (m *authenticationMiddleware) isValidBearerToken(token string) bool {
	// Use the provided Validate function to validate the token
	return m.options.Validate(token)
}

// Validates the session cookie (Check integrity, expiration, and revocation).
func (m *authenticationMiddleware) isValidSessionCookie(sessionID string, r *http.Request, c *RouteContext) bool {
	// Perform the session validation checks as before
	isValid, err := m.options.ValidateSession(sessionID)
	if err != nil || !isValid {
		// Log error: session validation failure
		return false
	}

	// Fetch session details like IP and User-Agent for session binding
	sessionDetails, err := m.options.GetSessionDetails(sessionID)
	if err != nil || sessionDetails == nil {
		return false
	}

	// Check if IP and User-Agent from the session match with the incoming request
	if sessionDetails.IP != r.RemoteAddr || sessionDetails.UserAgent != r.UserAgent() {
		// Log potential session hijacking attempt
		return false
	}

	// Check for session expiration
	if m.hasSessionExpired(sessionDetails.CreatedAt) {
		return false
	}

	// Session is valid, so extend its expiration by updating the cookie
	m.extendSessionExpiration(c)

	return true
}

// Extend the session cookie expiration by updating the `Expires` and `Max-Age` values
func (m *authenticationMiddleware) extendSessionExpiration(c *RouteContext) {

	// Define the session duration (e.g., 1 hour)
	sessionDuration := 1 * time.Hour

	// Set the cookie with the new expiration
	cookie, err := c.Request.Cookie(GetAppSessionCookieName())
	if err != nil {
		// Handle the error (e.g., no cookie found, or some other error)
		// You may want to return or handle the failure
		return
	}

	// this pushes the cookie, but if the contents are a jwt with a exp. that would need to be handled
	cookie.Expires = time.Now().Add(sessionDuration)
	cookie.MaxAge = int(sessionDuration.Seconds())
	http.SetCookie(c.Response, cookie)

	// Optionally, log the session expiration extension for auditing purposes
	// log.Printf("Session cookie expiration extended for session: %s", sessionID)
}

// This function checks whether the session has expired.
func (m *authenticationMiddleware) hasSessionExpired(createdAt time.Time) bool {
	// Example: Sessions expire after 1 hour
	sessionTimeout := 1 * time.Hour
	if time.Since(createdAt) > sessionTimeout {
		return true
	}
	return false
}

// This function fetches session details such as IP and User-Agent from session store.
func (m *authenticationMiddleware) getSessionDetails(sessionID string) (*SessionDetails, error) {
	// Call the function set in AuthenticationOptions to retrieve session details
	return m.options.GetSessionDetails(sessionID)
}

// This function should retrieve the session creation time.
func (m *authenticationMiddleware) getSessionCreationTime(sessionID string) time.Time {
	// Call the function set in AuthenticationOptions to retrieve the session creation time
	return m.options.GetSessionCreationTime(sessionID)
}

// Example implementation for validating session from a store.
func (m *authenticationMiddleware) validateSession(sessionID string) (bool, error) {
	// Use the function provided in AuthenticationOptions to validate the session
	return m.options.ValidateSession(sessionID)
}
