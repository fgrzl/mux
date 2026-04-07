package test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocsShouldNotUseStalePublicAPINames(t *testing.T) {
	root := filepath.Clean("..")

	patterns := []struct {
		name    string
		pattern *regexp.Regexp
	}{
		{name: "Tags", pattern: regexp.MustCompile(`(^|[^A-Za-z])Tags\(`)},
		{name: "Summary", pattern: regexp.MustCompile(`(^|[^A-Za-z])Summary\(`)},
		{name: "Description", pattern: regexp.MustCompile(`(^|[^A-Za-z])Description\(`)},
		{name: "OperationID", pattern: regexp.MustCompile(`(^|[^A-Za-z])OperationID\(`)},
		{name: "AcceptJSON", pattern: regexp.MustCompile(`(^|[^A-Za-z])AcceptJSON\(`)},
		{name: "Responds", pattern: regexp.MustCompile(`(^|[^A-Za-z])Responds\(`)},
		{name: "RateLimit", pattern: regexp.MustCompile(`(^|[^A-Za-z])RateLimit\(`)},
		{name: "NewRouteGroup", pattern: regexp.MustCompile(`(^|[^A-Za-z])NewRouteGroup\(`)},
		{name: "RequirePermissions", pattern: regexp.MustCompile(`(^|[^A-Za-z])RequirePermissions\(`)},
		{name: "WithValidator", pattern: regexp.MustCompile(`(^|[^A-Za-z])WithValidator\(`)},
		{name: "WithTokenCreator", pattern: regexp.MustCompile(`(^|[^A-Za-z])WithTokenCreator\(`)},
		{name: "WithTokenTTL", pattern: regexp.MustCompile(`(^|[^A-Za-z])WithTokenTTL\(`)},
		{name: "WithRoles", pattern: regexp.MustCompile(`(^|[^A-Za-z])WithRoles\(`)},
		{name: "WithPermissions", pattern: regexp.MustCompile(`(^|[^A-Za-z])WithPermissions\(`)},
		{name: "cookiekit.WithPath", pattern: regexp.MustCompile(`cookiekit\.WithPath\(`)},
		{name: "cookiekit.WithDomain", pattern: regexp.MustCompile(`cookiekit\.WithDomain\(`)},
	}

	var failures []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		lines := strings.Split(string(data), "\n")
		for lineIndex, line := range lines {
			for _, entry := range patterns {
				if entry.pattern.MatchString(line) {
					relPath, err := filepath.Rel(root, path)
					if err != nil {
						relPath = path
					}
					failures = append(failures, fmt.Sprintf("%s:%d uses stale API token %q", filepath.ToSlash(relPath), lineIndex+1, entry.name))
				}
			}
		}

		return nil
	})
	require.NoError(t, err)

	sort.Strings(failures)
	require.Empty(t, failures, strings.Join(failures, "\n"))
}
