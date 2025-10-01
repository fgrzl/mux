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
	HeaderAccept                        = "Accept"
	HeaderAcceptEncoding                = "Accept-Encoding"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAuthorization                 = "Authorization"
	HeaderCacheControl                  = "Cache-Control"
	HeaderContentDisposition            = "Content-Disposition"
	HeaderContentEncoding               = "Content-Encoding"
	HeaderContentLength                 = "Content-Length"
	HeaderContentType                   = "Content-Type"
	HeaderCookie                        = "Cookie"
	HeaderETag                          = "ETag"
	HeaderForwarded                     = "Forwarded" // RFC 7239
	HeaderLocation                      = "Location"
	HeaderOrigin                        = "Origin"
	HeaderRetryAfter                    = "Retry-After"
	HeaderSetCookie                     = "Set-Cookie"
	HeaderTransferEncoding              = "Transfer-Encoding"
	HeaderUserAgent                     = "User-Agent"
	HeaderVary                          = "Vary"
	HeaderXForwardedFor                 = "X-Forwarded-For"
	HeaderXForwardedHost                = "X-Forwarded-Host"
	HeaderXForwardedPort                = "X-Forwarded-Port"
	HeaderXForwardedProto               = "X-Forwarded-Proto"
	HeaderXRealIP                       = "X-Real-IP"
)
