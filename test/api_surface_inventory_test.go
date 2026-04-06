package test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocsAndExamplesShouldNotReferenceRemovedPublicSurface(t *testing.T) {
	t.Parallel()

	root := filepath.Clean("..")
	targets := []string{
		filepath.Join(root, "README.md"),
		filepath.Join(root, "docs"),
		filepath.Join(root, "examples"),
	}
	forbidden := []string{
		"github.com/fgrzl/mux/v2",
		"github.com/fgrzl/mux/pkg/",
		"github.com/fgrzl/mux/openapi",
		"github.com/fgrzl/mux/compression",
		"github.com/fgrzl/mux/cors",
		"github.com/fgrzl/mux/forwarded",
		"github.com/fgrzl/mux/logging",
		"pkg/",
		"NewRouteGroup",
		"WithOperationIDErr(",
		"WithJsonBodyErr(",
		"WithCreatedResponseErr(",
		"WithBadRequestResponse(",
		"WithUnauthorizedResponse(",
		"WithForbiddenResponse(",
		"WithNotFoundResponse(",
		"With301Response(",
		"With302Response(",
		"With303Response(",
		"With307Response(",
		"With308Response(",
		"WithPathParamErr(",
		"WithQueryParamErr(",
		"WithRequiredQueryParamErr(",
		"WithParam(",
		"WithParamErr(",
		"RouteParam",
		"WithPathParams(",
		"WithQueryParams(",
		"WithRequiredQueryParams(",
		"WithHeaderParams(",
		"WithRequiredHeaderParams(",
		"WithCookieParams(",
		"WithRequiredCookieParams(",
		"DetachedRoute(",
		"HandleRoute(",
		"NewSelectiveRateLimiter(",
		"WithCleanupInterval(",
		"WithValidator(",
		"WithTokenCreator(",
		"WithTokenTTL(",
		"WithCSRFProtection(",
		"WithRoles(",
		"WithPermissions(",
		"WithScopes(",
		"WithAllowedOrigins(",
		"WithAllowedMethods(",
		"WithAllowedHeaders(",
		"WithExposedHeaders(",
		"WithAllowCredentials(",
		"WithOriginWildcard(",
		"WithMaxAge(",
		"WithGeoIPDatabase(",
		"WithOperation(",
		"router.UseAuthentication(",
		"router.UseAuthorization(",
		"router.UseCompression()",
		"router.UseEnforceHTTPS()",
		"Safe()",
		".Routes(",
		".InfoObject(",
		"GenerateSpecFromRoutes(",
		"GenerateAndSave(",
		"GenerateSchemaForType(",
		"NewOpenAPISpec(",
		"RouteData",
		"c.Respond().",
		"routeCtx.Respond().",
		"Invoke(ctx RouteContext, next HandlerFunc)",
		"Invoke(c RouteContext, next HandlerFunc)",
		"Invoke(c routing.RouteContext, next HandlerFunc)",
		"Invoke(c mux.RouteContext, next mux.HandlerFunc)",
		"Params().",
		"RouteBuilderParams",
		"Params().String(",
		"Params().UUID(",
		"Params().Int(",
		"Params().Int16(",
		"Params().Int32(",
		"Params().Int64(",
		"FormValue(",
		"FormValues(",
		"FormUUID(",
		"FormUUIDs(",
		"FormInt(",
		"FormInts(",
		"FormInt16(",
		"FormInt16s(",
		"FormInt32(",
		"FormInt32s(",
		"FormInt64(",
		"FormInt64s(",
		"FormBool(",
		"FormBools(",
		"FormFloat32(",
		"FormFloat32s(",
		"FormFloat64(",
		"FormFloat64s(",
		"SetService(",
		"GetService(",
		"HTTPHandler(",
		"HTTPHandlerFunc(",
		"RegisterMany(",
		"RedirectURL(",
		"DefaultAppSessionCookieName",
		"DefaultTwoFactorCookieName",
		"DefaultIDPSessionCookieName",
		"GetAppSessionCookieName(",
		"GetTwoFactorCookieName(",
		"GetIDPSessionCookieName(",
		"SetAppSessionCookieName(",
		"SetTwoFactorCookieName(",
		"SetIDPSessionCookieName(",
		"GenerateCSRFToken(",
		"GenerateCSRFTokenErr(",
		"Ãƒ",
		"Ã¢",
		"Ã‚",
		"Ã¯Â¸",
		"ï¿½",
	}

	var failures []string
	for _, target := range targets {
		info, err := os.Stat(target)
		if err != nil {
			t.Fatalf("stat %s: %v", target, err)
		}

		if !info.IsDir() {
			failures = append(failures, scanSurfaceFile(t, target, forbidden)...)
			continue
		}

		err = filepath.WalkDir(target, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".go" && ext != ".md" {
				return nil
			}
			failures = append(failures, scanSurfaceFile(t, path, forbidden)...)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", target, err)
		}
	}

	if len(failures) > 0 {
		t.Fatalf("forbidden public-surface references found:\n%s", strings.Join(failures, "\n"))
	}
}

func scanSurfaceFile(t *testing.T, path string, forbidden []string) []string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	text := string(data)
	var hits []string
	for _, token := range forbidden {
		if strings.Contains(text, token) {
			hits = append(hits, path+": "+token)
		}
	}
	return hits
}
