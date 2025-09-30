package common

// ServiceKey is the type used for service registry keys stored on RouteContext.
type ServiceKey string

// MIME types shared by internal packages.
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
	HeaderXForwardedHost     = "X-Forwarded-Host"
	HeaderXForwardedPort     = "X-Forwarded-Port"
	HeaderXRealIP            = "X-Real-IP"
	HeaderForwarded          = "Forwarded" // RFC 7239
	HeaderAcceptEncoding     = "Accept-Encoding"
	HeaderContentEncoding    = "Content-Encoding"
	HeaderVary               = "Vary"
	HeaderUserAgent          = "User-Agent"
	HeaderAuthorization      = "Authorization"
	HeaderLocation           = "Location"
	HeaderSetCookie          = "Set-Cookie"
	HeaderCookie             = "Cookie"
	HeaderAccept             = "Accept"
	HeaderRetryAfter         = "Retry-After"
	HeaderCacheControl       = "Cache-Control"
	HeaderETag               = "ETag"
	HeaderContentLength      = "Content-Length"
	HeaderTransferEncoding   = "Transfer-Encoding"
)
