package mux

// HTTP constants for MIME types and headers used in web APIs and HTTP communication.
// This file provides commonly used constants to avoid magic strings and improve maintainability.

// Common MIME types used in web APIs and HTTP
const (
	MimeJSON              = "application/json"
	MimeXML               = "application/xml"
	MimeFormURLEncoded    = "application/x-www-form-urlencoded"
	MimeMultipartFormData = "multipart/form-data"
	MimeTextPlain         = "text/plain"
	MimeTextHTML          = "text/html"
	MimeTextCSV           = "text/csv"
	MimeOctetStream       = "application/octet-stream"
	MimePDF               = "application/pdf"
	MimeZIP               = "application/zip"
	MimePNG               = "image/png"
	MimeJPEG              = "image/jpeg"
	MimeGIF               = "image/gif"
	MimeSVG               = "image/svg+xml"
	MimeWebP              = "image/webp"
	MimeMP4               = "video/mp4"
	MimeMP3               = "audio/mpeg"
	MimeWAV               = "audio/wav"
	MimeOGG               = "audio/ogg"
	MimeJSONAPI           = "application/vnd.api+json"
	MimeOpenAPI           = "application/vnd.oai.openapi"
	MimeYAML              = "application/x-yaml"
	MimeProblemJSON       = "application/problem+json"
)

// Common HTTP header names
const (
	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
	HeaderXForwardedFor      = "X-Forwarded-For"
	HeaderXForwardedProto    = "X-Forwarded-Proto"
)
