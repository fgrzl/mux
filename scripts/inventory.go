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
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

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
	return os.WriteFile("inventory.md", []byte(sb.String()), 0644)
}

// readModules reads and parses go.mod
func readModules() ([]string, error) {
	file, err := os.Open("go.mod")
	if err != nil {
		return nil, err
	}
	defer file.Close()

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

// findTests finds all test functions
func findTests() ([]TestEntry, error) {
	var tests []TestEntry
	testRegex := regexp.MustCompile(`^func (Test[A-Za-z0-9_]*)\(`)

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip non-test files and vendor/build directories
		if !strings.HasSuffix(path, "_test.go") || strings.Contains(path, "vendor") {
			return nil
		}

		// Get package name from path
		pkg := filepath.Dir(path)

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if matches := testRegex.FindStringSubmatch(line); matches != nil {
				tests = append(tests, TestEntry{
					Package: pkg,
					Name:    matches[1],
					File:    filepath.Base(path),
				})
			}
		}

		return scanner.Err()
	})

	sort.Slice(tests, func(i, j int) bool {
		if tests[i].Package != tests[j].Package {
			return tests[i].Package < tests[j].Package
		}
		return tests[i].Name < tests[j].Name
	})

	return tests, err
}

// BenchEntry represents a benchmark function
type BenchEntry struct {
	Package string
	Name    string
	File    string
}

// findBenchmarks finds all benchmark functions
func findBenchmarks() ([]BenchEntry, error) {
	var benches []BenchEntry
	benchRegex := regexp.MustCompile(`^func (Benchmark[A-Za-z0-9_]*)\(`)

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Only look at benchmark test files
		if !strings.HasSuffix(path, "_bench_test.go") || strings.Contains(path, "vendor") {
			return nil
		}

		// Get package name from path
		pkg := filepath.Dir(path)

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if matches := benchRegex.FindStringSubmatch(line); matches != nil {
				benches = append(benches, BenchEntry{
					Package: pkg,
					Name:    matches[1],
					File:    filepath.Base(path),
				})
			}
		}

		return scanner.Err()
	})

	sort.Slice(benches, func(i, j int) bool {
		if benches[i].Package != benches[j].Package {
			return benches[i].Package < benches[j].Package
		}
		return benches[i].Name < benches[j].Name
	})

	return benches, err
}

// formatModules formats modules section
func formatModules(modules []string) string {
	var sb strings.Builder
	sb.WriteString("## Dependencies\n\n")
	sb.WriteString(fmt.Sprintf("Total: %d\n\n", len(modules)))
	sb.WriteString("```\n")
	for _, m := range modules {
		sb.WriteString(m + "\n")
	}
	sb.WriteString("```\n\n")
	return sb.String()
}

// formatTests formats tests section
func formatTests(tests []TestEntry) string {
	var sb strings.Builder
	sb.WriteString("## Tests\n\n")
	sb.WriteString(fmt.Sprintf("Total: %d\n\n", len(tests)))

	// Group by package
	byPackage := make(map[string][]TestEntry)
	for _, t := range tests {
		byPackage[t.Package] = append(byPackage[t.Package], t)
	}

	// Sort packages
	var packages []string
	for p := range byPackage {
		packages = append(packages, p)
	}
	sort.Strings(packages)

	// Output by package
	for _, pkg := range packages {
		sb.WriteString(fmt.Sprintf("### %s (%d tests)\n\n", pkg, len(byPackage[pkg])))

		// Group by file within package
		byFile := make(map[string][]TestEntry)
		for _, t := range byPackage[pkg] {
			byFile[t.File] = append(byFile[t.File], t)
		}

		// Sort files
		var files []string
		for f := range byFile {
			files = append(files, f)
		}
		sort.Strings(files)

		// Output by file
		for _, file := range files {
			sb.WriteString(fmt.Sprintf("**%s**\n\n", file))
			for _, t := range byFile[file] {
				sb.WriteString(fmt.Sprintf("- `%s`\n", t.Name))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// formatBenchmarks formats benchmarks section
func formatBenchmarks(benches []BenchEntry) string {
	var sb strings.Builder
	sb.WriteString("## Benchmarks\n\n")
	sb.WriteString(fmt.Sprintf("Total: %d\n\n", len(benches)))

	// Group by package
	byPackage := make(map[string][]BenchEntry)
	for _, b := range benches {
		byPackage[b.Package] = append(byPackage[b.Package], b)
	}

	// Sort packages
	var packages []string
	for p := range byPackage {
		packages = append(packages, p)
	}
	sort.Strings(packages)

	// Output by package
	for _, pkg := range packages {
		sb.WriteString(fmt.Sprintf("### %s (%d benchmarks)\n\n", pkg, len(byPackage[pkg])))

		// Group by file within package
		byFile := make(map[string][]BenchEntry)
		for _, b := range byPackage[pkg] {
			byFile[b.File] = append(byFile[b.File], b)
		}

		// Sort files
		var files []string
		for f := range byFile {
			files = append(files, f)
		}
		sort.Strings(files)

		// Output by file
		for _, file := range files {
			sb.WriteString(fmt.Sprintf("**%s**\n\n", file))
			for _, b := range byFile[file] {
				sb.WriteString(fmt.Sprintf("- `%s`\n", b.Name))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
