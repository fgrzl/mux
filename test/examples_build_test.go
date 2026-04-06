package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExamplesShouldBuild(t *testing.T) {
	root := filepath.Clean("..")
	examplesRoot := filepath.Join(root, "examples")

	exampleDirs, err := discoverExampleMainDirs(examplesRoot)
	require.NoError(t, err)
	require.NotEmpty(t, exampleDirs)

	for _, dir := range exampleDirs {
		rel, err := filepath.Rel(root, dir)
		require.NoError(t, err)

		exampleName := filepath.ToSlash(rel)
		t.Run(exampleName, func(t *testing.T) {
			if fileExists(filepath.Join(dir, "go.mod")) {
				runGoCommand(t, dir, "test", "./...")
				return
			}

			runGoCommand(t, root, "test", "./"+filepath.ToSlash(rel))
		})
	}
}

func discoverExampleMainDirs(examplesRoot string) ([]string, error) {
	seen := map[string]struct{}{}
	err := filepath.WalkDir(examplesRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "main.go" {
			return nil
		}
		seen[filepath.Dir(path)] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0, len(seen))
	for dir := range seen {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	return dirs, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func runGoCommand(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go %s failed in %s:\n%s", strings.Join(args, " "), dir, string(output))
	}
}
