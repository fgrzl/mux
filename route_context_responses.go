package mux

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

const ProblemTypeAboutBlank = "about:blank"

// ServerError writes a 500 Internal Server Error with problem details.
func (c *RouteContext) ServerError(title, detail string) {
	if title == "" {
		title = http.StatusText(http.StatusInternalServerError)
	}
	c.Problem(&ProblemDetails{
		Title:    title,
		Detail:   detail,
		Status:   http.StatusInternalServerError,
		Type:     ProblemTypeAboutBlank,
		Instance: getInstanceURI(c.Request),
	})
}

// BadRequest writes a 400 Bad Request with problem details.
func (c *RouteContext) BadRequest(title, detail string) {
	if title == "" {
		title = http.StatusText(http.StatusBadRequest)
	}
	c.Problem(&ProblemDetails{
		Title:    title,
		Detail:   detail,
		Status:   http.StatusBadRequest,
		Type:     ProblemTypeAboutBlank,
		Instance: getInstanceURI(c.Request),
	})
}

// Conflict writes a 409 Conflict with problem details.
func (c *RouteContext) Conflict(title, detail string) {
	c.Problem(&ProblemDetails{
		Title:    title,
		Detail:   detail,
		Status:   http.StatusConflict,
		Type:     ProblemTypeAboutBlank,
		Instance: getInstanceURI(c.Request),
	})
}

// Problem writes a problem+json response using RFC 7807.
func (c *RouteContext) Problem(problem *ProblemDetails) {
	if problem.Status == 0 {
		problem.Status = http.StatusInternalServerError
	}
	slog.Warn("problem response",
		slog.Int("status", problem.Status),
		slog.String("title", problem.Title),
		slog.String("detail", problem.Detail),
	)
	c.Response.Header().Set(HeaderContentType, MimeProblemJSON)
	c.Response.WriteHeader(problem.Status)
	json.NewEncoder(c.Response).Encode(problem)
}

// OK writes a 200 OK response with a JSON payload.
func (c *RouteContext) OK(model any) {
	c.JSON(http.StatusOK, model)
}

// JSON writes a JSON response with custom status code.
func (c *RouteContext) JSON(status int, model any) {
	c.Response.Header().Set(HeaderContentType, MimeJSON)
	c.Response.WriteHeader(status)
	json.NewEncoder(c.Response).Encode(model)
}

// Plain writes a plain text response.
func (c *RouteContext) Plain(status int, data []byte) {
	c.Response.Header().Set(HeaderContentType, MimeTextPlain)
	c.Response.WriteHeader(status)
	if len(data) > 0 {
		c.Response.Write(data)
	}
}

// HTML writes an HTML response.
func (c *RouteContext) HTML(status int, html string) {
	c.Response.Header().Set(HeaderContentType, MimeTextHTML)
	c.Response.WriteHeader(status)
	c.Response.Write([]byte(html))
}

// Created writes a 201 Created response with an optional JSON payload.
func (c *RouteContext) Created(model any) {
	c.JSON(http.StatusCreated, model)
}

// Accept writes a 202 Accepted response with an optional JSON payload.
func (c *RouteContext) Accept(model any) {
	c.JSON(http.StatusAccepted, model)
}

// NoContent writes a 204 No Content response.
func (c *RouteContext) NoContent() {
	c.Response.WriteHeader(http.StatusNoContent)
}

// NotFound writes a 404 Not Found response.
func (c *RouteContext) NotFound() {
	http.NotFound(c.Response, c.Request)
}

// Unauthorized writes a 401 Unauthorized response.
func (c *RouteContext) Unauthorized() {
	http.Error(c.Response, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

// Forbidden writes a 403 Forbidden response with a custom message.
func (c *RouteContext) Forbidden(message string) {
	http.Error(c.Response, message, http.StatusForbidden)
}

// File serves a static file.
func (c *RouteContext) File(filePath string) {
	http.ServeFile(c.Response, c.Request, filePath)
}

// Download serves a file with Content-Disposition attachment.
func (c *RouteContext) Download(filePath, filename string) {
	f, err := os.Open(filePath)
	if err != nil {
		c.ServerError("File not found", err.Error())
		return
	}
	defer f.Close()

	c.Response.Header().Set(HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response.Header().Set(HeaderContentType, MimeOctetStream)
	c.Response.WriteHeader(http.StatusOK)
	io.Copy(c.Response, f)
}

// Redirect issues a redirect with custom status code.
func (c *RouteContext) Redirect(status int, url string) {
	target := c.ensureAbsoluteURL(url)
	http.Redirect(c.Response, c.Request, target, status)
}

// TemporaryRedirect issues a 307 Temporary Redirect.
func (c *RouteContext) TemporaryRedirect(url string) {
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// PermanentRedirect issues a 308 Permanent Redirect.
func (c *RouteContext) PermanentRedirect(url string) {
	c.Redirect(http.StatusPermanentRedirect, url)
}

// GetRedirectScheme returns the scheme to use when constructing redirect URLs.
func (c *RouteContext) GetRedirectScheme() string {
	if scheme := c.Request.Header.Get(HeaderXForwardedProto); scheme != "" {
		return scheme
	}
	if c.Request.URL.Scheme != "" {
		return c.Request.URL.Scheme
	}
	return "http"
}

// ensureAbsoluteURL resolves a relative path to a full URL using request info.
func (c *RouteContext) ensureAbsoluteURL(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}

	// Ensure url starts with a slash and does not start with '//' or '/\'
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	if len(url) > 1 && (url[1] == '/' || url[1] == '\\') {
		url = "/"
	}

	if c.ClientURL != nil {
		// Ensure no double slash
		return strings.TrimRight(c.ClientURL.String(), "/") + url
	}
	scheme := c.GetRedirectScheme()
	return fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, url)
}
