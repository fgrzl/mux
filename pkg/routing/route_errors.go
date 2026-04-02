package routing

import "errors"

// ErrMissingBody is returned when a request body is required by the route
// OpenAPI metadata but the incoming request has no body.
var ErrMissingBody = errors.New("request body required")

// IsMissingBodyError returns true when the given error (or any wrapped error)
// indicates that the request body was required but not present.
func IsMissingBodyError(err error) bool {
	return errors.Is(err, ErrMissingBody)
}
