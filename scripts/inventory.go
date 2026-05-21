package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

func main() {
	if err := generateInventory(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func generateInventory() error {
	var sb strings.Builder

	// Header
	sb.WriteString("# Project Inventory\n\n")
	fmt.Fprintf(&sb, "Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Modules
	modules, err := readModules()
	if err != nil {
		return err
	}
	sb.WriteString(formatModules(modules))

	// Tests
	tests, err := findTests()
	if err != nil {
		return err
	}
	sb.WriteString(formatTests(tests))

	// Benchmarks
	benches, err := findBenchmarks()
	if err != nil {
		return err
	}
	sb.WriteString(formatBenchmarks(benches))

	// Write to file
	return os.WriteFile("inventory.md", []byte(sb.String()), 0600)
}

// readModules reads and parses go.mod
func readModules() ([]string, error) {
	file, err := os.Open("go.mod")
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var modules []string
	scanner := bufio.NewScanner(file)
	inRequire := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "require (" {
			inRequire = true
			continue
		}

		if inRequire && line == ")" {
			inRequire = false
			continue
		}

		if inRequire && line != "" && !strings.HasPrefix(line, "//") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				modules = append(modules, parts[0])
			}
		} else if !inRequire && strings.HasPrefix(line, "require ") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				modules = append(modules, parts[1])
			}
		}
	}

	sort.Strings(modules)
	return modules, scanner.Err()
}

// TestEntry represents a test function
type TestEntry struct {
	Package string
	Name    string
	File    string
}

// BenchEntry represents a benchmark function
type BenchEntry struct {
	Package string
	Name    string
	File    string
}

// namedFileRow is a package-qualified name within a file (tests or benchmarks).
type namedFileRow struct {
	Package string
	Name    string
	File    string
}

func walkSourceFiles(keep func(path string) bool, onFile func(pkg, path string, scanner *bufio.Scanner) error) error {
	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !keep(path) {
			return nil
		}
		pkg := filepath.Dir(path)
		f, err := os.Open(path) //nolint:gosec // G304: inventory walks *_test.go paths under the repo
		if err != nil {
			return nil
		}
		defer func() { _ = f.Close() }()
		return onFile(pkg, path, bufio.NewScanner(f))
	})
}

func writeInventoryMarkdown(sb *strings.Builder, section string, rows []namedFileRow, unitPlural string) {
	fmt.Fprintf(sb, "## %s\n\n", section)
	fmt.Fprintf(sb, "Total: %d\n\n", len(rows))

	byPackage := make(map[string][]namedFileRow)
	for _, row := range rows {
		byPackage[row.Package] = append(byPackage[row.Package], row)
	}

	packages := make([]string, 0, len(byPackage))
	for p := range byPackage {
		packages = append(packages, p)
	}
	sort.Strings(packages)

	for _, pkg := range packages {
		fmt.Fprintf(sb, "### %s (%d %s)\n\n", pkg, len(byPackage[pkg]), unitPlural)

		byFile := make(map[string][]namedFileRow)
		for _, row := range byPackage[pkg] {
			byFile[row.File] = append(byFile[row.File], row)
		}

		files := make([]string, 0, len(byFile))
		for f := range byFile {
			files = append(files, f)
		}
		sort.Strings(files)

		for _, file := range files {
			fmt.Fprintf(sb, "**%s**\n\n", file)
			for _, row := range byFile[file] {
				fmt.Fprintf(sb, "- `%s`\n", row.Name)
			}
			sb.WriteString("\n")
		}
	}
}

func collectDecls[E any](pathSuffix string, line *regexp.Regexp, row func(pkg, path, name string) E) ([]E, error) {
	var out []E
	err := walkSourceFiles(
		func(path string) bool {
			return strings.HasSuffix(path, pathSuffix) && !strings.Contains(path, "vendor")
		},
		func(pkg, path string, scanner *bufio.Scanner) error {
			for scanner.Scan() {
				if m := line.FindStringSubmatch(scanner.Text()); m != nil {
					out = append(out, row(pkg, path, m[1]))
				}
			}
			return scanner.Err()
		},
	)
	return out, err
}

// findTests finds all test functions
func findTests() ([]TestEntry, error) {
	testRegex := regexp.MustCompile(`^func (Test[A-Za-z0-9_]*)\(`)
	tests, err := collectDecls("_test.go", testRegex, func(pkg, path, name string) TestEntry {
		return TestEntry{Package: pkg, Name: name, File: filepath.Base(path)}
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(tests, func(i, j int) bool {
		if tests[i].Package != tests[j].Package {
			return tests[i].Package < tests[j].Package
		}
		return tests[i].Name < tests[j].Name
	})
	return tests, nil
}

// findBenchmarks finds all benchmark functions
func findBenchmarks() ([]BenchEntry, error) {
	benchRegex := regexp.MustCompile(`^func (Benchmark[A-Za-z0-9_]*)\(`)
	benches, err := collectDecls("_bench_test.go", benchRegex, func(pkg, path, name string) BenchEntry {
		return BenchEntry{Package: pkg, Name: name, File: filepath.Base(path)}
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(benches, func(i, j int) bool {
		if benches[i].Package != benches[j].Package {
			return benches[i].Package < benches[j].Package
		}
		return benches[i].Name < benches[j].Name
	})
	return benches, nil
}

// formatModules formats modules section
func formatModules(modules []string) string {
	var sb strings.Builder
	sb.WriteString("## Dependencies\n\n")
	fmt.Fprintf(&sb, "Total: %d\n\n", len(modules))
	sb.WriteString("```\n")
	for _, m := range modules {
		sb.WriteString(m + "\n")
	}
	sb.WriteString("```\n\n")
	return sb.String()
}

// formatTests formats tests section
func formatTests(tests []TestEntry) string {
	rows := make([]namedFileRow, len(tests))
	for i := range tests {
		rows[i] = namedFileRow{Package: tests[i].Package, Name: tests[i].Name, File: tests[i].File}
	}
	var sb strings.Builder
	writeInventoryMarkdown(&sb, "Tests", rows, "tests")
	return sb.String()
}

// formatBenchmarks formats benchmarks section
func formatBenchmarks(benches []BenchEntry) string {
	rows := make([]namedFileRow, len(benches))
	for i := range benches {
		rows[i] = namedFileRow{Package: benches[i].Package, Name: benches[i].Name, File: benches[i].File}
	}
	var sb strings.Builder
	writeInventoryMarkdown(&sb, "Benchmarks", rows, "benchmarks")
	return sb.String()
}
