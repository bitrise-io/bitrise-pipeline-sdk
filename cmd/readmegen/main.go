// readmegen rewrites the typed step builders tables in README.md.
//
// It scans the step directory for generated builder files (gen_*.go), extracts
// the step ID and constructor function name from each file, skips deprecated
// steps, sorts the results alphabetically by step ID, and replaces the content
// between two pairs of HTML comment markers in the target README:
//
//   - <!-- step-table-start / end -->  — the main unversioned builders table.
//   - <!-- step-versions-start / end --> — the versioned builders table for
//     multi-major steps (only present when the README contains those markers).
//
// Usage (standalone):
//
//	go run ./cmd/readmegen [--step-dir=step] [--readme=README.md]
//
// Usage (via go generate from the step/ directory):
//
//	go generate ./step/
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	markerStart    = "<!-- step-table-start -->"
	markerEnd      = "<!-- step-table-end -->"
	markerVerStart = "<!-- step-versions-start -->"
	markerVerEnd   = "<!-- step-versions-end -->"
)

// versionedFileRe matches generated per-major versioned builder files
// (e.g. gen_xcode_test_v6.go).  These files define TypeNameVNBuilder and
// should be excluded from the README table; the alias file provides the
// canonical unversioned entry for such steps.
var versionedFileRe = regexp.MustCompile(`_v\d+\.go$`)

func main() {
	stepDir := flag.String("step-dir", "step", "directory containing gen_*.go builder files")
	readme := flag.String("readme", "README.md", "path to the README file to update")
	flag.Parse()

	entries, err := collectBuilders(*stepDir)
	if err != nil {
		fatalf("scanning %s: %v", *stepDir, err)
	}

	vEntries, err := collectVersionedBuilders(*stepDir)
	if err != nil {
		fatalf("scanning %s for versioned builders: %v", *stepDir, err)
	}

	if err := updateReadme(*readme, entries, vEntries); err != nil {
		fatalf("updating %s: %v", *readme, err)
	}

	fmt.Printf("updated %s with %d step builders (%d versioned)\n", *readme, len(entries), len(vEntries))
}

// builderEntry holds the data needed for one row in the main builders table.
type builderEntry struct {
	stepID   string // e.g. "git-clone"
	funcName string // e.g. "GitClone"
}

// versionedBuilderEntry holds the data needed for one row in the versioned
// builders table (one row per major version of a multi-major step).
type versionedBuilderEntry struct {
	stepID   string // e.g. "git-clone"
	funcName string // e.g. "GitCloneV8"
	major    int    // e.g. 8
}

// collectBuilders scans dir for gen_*.go files, parses each one, and returns
// a sorted slice of non-deprecated builder entries.
func collectBuilders(dir string) ([]builderEntry, error) {
	pattern := filepath.Join(dir, "gen_*.go")
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no gen_*.go files found in %s", dir)
	}

	var entries []builderEntry
	for _, p := range paths {
		// Skip per-major versioned files (gen_*_v{N}.go).  The alias file
		// (gen_*_alias.go) provides the canonical table entry for multi-major
		// steps and is processed normally below.
		if versionedFileRe.MatchString(p) {
			continue
		}
		entry, deprecated, err := parseBuilderFile(p)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", p, err)
		}
		if deprecated {
			continue
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].stepID < entries[j].stepID
	})
	return entries, nil
}

// collectVersionedBuilders scans dir for gen_*_v{N}.go files, parses each
// one, and returns a sorted slice of non-deprecated versioned builder entries.
// Steps are sorted alphabetically; majors within a step are sorted ascending.
func collectVersionedBuilders(dir string) ([]versionedBuilderEntry, error) {
	pattern := filepath.Join(dir, "gen_*.go")
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var entries []versionedBuilderEntry
	for _, p := range paths {
		// Only process per-major versioned files (gen_*_v{N}.go).
		if !versionedFileRe.MatchString(p) {
			continue
		}
		entry, deprecated, err := parseVersionedBuilderFile(p)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", p, err)
		}
		if deprecated {
			continue
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].stepID != entries[j].stepID {
			return entries[i].stepID < entries[j].stepID
		}
		return entries[i].major < entries[j].major // ascending within a step
	})
	return entries, nil
}

// parseVersionedBuilderFile extracts the step ID, constructor name, and major
// version number from a gen_*_v{N}.go file.
func parseVersionedBuilderFile(path string) (entry versionedBuilderEntry, deprecated bool, err error) {
	f, err := os.Open(path)
	if err != nil {
		return versionedBuilderEntry{}, false, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)

	// Line 1: // Code generated by cmd/stepgen. DO NOT EDIT.
	if !sc.Scan() {
		return versionedBuilderEntry{}, false, fmt.Errorf("file is empty")
	}

	// Line 2: // Step: git-clone (8.5.0)
	if !sc.Scan() {
		return versionedBuilderEntry{}, false, fmt.Errorf("missing step header on line 2")
	}
	headerLine := sc.Text()
	stepID, err := parseStepID(headerLine)
	if err != nil {
		return versionedBuilderEntry{}, false, err
	}
	version, err := parseVersionFromHeader(headerLine)
	if err != nil {
		return versionedBuilderEntry{}, false, err
	}
	major, _ := strconv.Atoi(majorFromVersion(version))

	// Scan the rest of the file for the constructor and any deprecation notice.
	var funcName string
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "// Deprecated:") {
			deprecated = true
		}
		if strings.HasPrefix(line, "func ") && funcName == "" {
			if name, ok := parseConstructorName(line); ok {
				funcName = name
			}
		}
	}
	if err := sc.Err(); err != nil {
		return versionedBuilderEntry{}, false, err
	}
	if funcName == "" {
		return versionedBuilderEntry{}, false, fmt.Errorf("could not find constructor function")
	}

	return versionedBuilderEntry{stepID: stepID, funcName: funcName, major: major}, deprecated, nil
}

// parseVersionFromHeader extracts the version string from a step header line:
//
//	// Step: git-clone (8.5.0)  →  "8.5.0"
func parseVersionFromHeader(line string) (string, error) {
	i := strings.LastIndex(line, "(")
	j := strings.LastIndex(line, ")")
	if i < 0 || j <= i {
		return "", fmt.Errorf("no version in header %q", line)
	}
	return line[i+1 : j], nil
}

// majorFromVersion returns the major component of a semver string.
//
//	"8.5.0" → "8"
func majorFromVersion(v string) string {
	if idx := strings.IndexByte(v, '.'); idx >= 0 {
		return v[:idx]
	}
	return v
}

// parseBuilderFile extracts the step ID and constructor function name from a
// generated builder file, and reports whether the step is deprecated.
//
// It relies on two invariants guaranteed by cmd/stepgen:
//   - Line 2 is always: // Step: <step-id> (<version>)
//   - The constructor is always: func TypeName() *TypeNameBuilder {
//   - Deprecated steps have a line starting with: // Deprecated:
func parseBuilderFile(path string) (entry builderEntry, deprecated bool, err error) {
	f, err := os.Open(path)
	if err != nil {
		return builderEntry{}, false, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)

	// Line 1: // Code generated by cmd/stepgen. DO NOT EDIT.
	if !sc.Scan() {
		return builderEntry{}, false, fmt.Errorf("file is empty")
	}

	// Line 2: // Step: <step-id> (<version>)
	if !sc.Scan() {
		return builderEntry{}, false, fmt.Errorf("missing step header on line 2")
	}
	stepID, err := parseStepID(sc.Text())
	if err != nil {
		return builderEntry{}, false, err
	}

	// Scan the rest of the file for a // Deprecated: line and the func constructor.
	var funcName string
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "// Deprecated:") {
			deprecated = true
		}
		// Match both legacy zero-arg constructors and the current variadic form:
		//   func GitClone() *GitCloneBuilder {
		//   func GitClone(version ...string) *GitCloneBuilder {
		//   func GitClone(version ...string) *GitCloneV8Builder {  (alias file)
		if strings.HasPrefix(line, "func ") && funcName == "" {
			if name, ok := parseConstructorName(line); ok {
				funcName = name
			}
		}
	}
	if err := sc.Err(); err != nil {
		return builderEntry{}, false, err
	}
	if funcName == "" {
		return builderEntry{}, false, fmt.Errorf("could not find constructor function")
	}

	return builderEntry{stepID: stepID, funcName: funcName}, deprecated, nil
}

// parseStepID extracts the step ID from a header line of the form:
//
//	// Step: git-clone (8.2.1)
func parseStepID(line string) (string, error) {
	// Strip the "// Step: " prefix.
	line = strings.TrimPrefix(line, "// Step: ")
	// Everything before the space before "(" is the step ID.
	i := strings.LastIndex(line, " (")
	if i < 0 {
		return "", fmt.Errorf("unexpected step header format: %q", line)
	}
	return strings.TrimSpace(line[:i]), nil
}

// parseConstructorName extracts the function name from a constructor line.
// It recognises both the legacy zero-arg form and the current variadic form:
//
//	func GitClone() *GitCloneBuilder {
//	func GitClone(version ...string) *GitCloneBuilder {
//	func GitClone(version ...string) *GitCloneV8Builder {
//
// It deliberately does NOT match receiver methods such as:
//
//	func (b *GitCloneBuilder) WithBranch(...) ...
func parseConstructorName(line string) (string, bool) {
	line = strings.TrimPrefix(line, "func ")
	// Variadic constructor (post-patchgen and alias files).
	if i := strings.Index(line, "(version ...string)"); i >= 0 {
		return line[:i], true
	}
	// Legacy zero-arg constructor.
	if i := strings.Index(line, "()"); i >= 0 {
		return line[:i], true
	}
	return "", false
}

// updateReadme replaces the content between the step-table markers (and, if
// present, the step-versions markers) in the README file at path.
func updateReadme(path string, entries []builderEntry, vEntries []versionedBuilderEntry) error {
	original, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Always update the main builders table.
	table := buildTable(entries)
	updated, err := replaceMarkers(original, markerStart, markerEnd, table)
	if err != nil {
		return err
	}

	// Update the versioned builders table only when the markers exist in the file.
	vTable := buildVersionsTable(vEntries)
	updated, err = replaceMarkersIfPresent(updated, markerVerStart, markerVerEnd, vTable)
	if err != nil {
		return err
	}

	return os.WriteFile(path, updated, 0644)
}

// buildTable renders the main builders table wrapped in a <details> block so
// it is collapsed by default in GitHub Markdown. The summary line includes the
// live entry count so it stays accurate on every regeneration.
func buildTable(entries []builderEntry) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "<details>\n<summary>Show all %d step builders</summary>\n\n", len(entries))
	fmt.Fprintln(&b, "| Function | Step |")
	fmt.Fprintln(&b, "|---|---|")
	for _, e := range entries {
		fmt.Fprintf(&b, "| `step.%s()` | `%s` |\n", e.funcName, e.stepID)
	}
	fmt.Fprint(&b, "\n</details>")
	return b.Bytes()
}

// buildVersionsTable renders the versioned builders table wrapped in a
// <details> block. Returns nil when there are no entries.
func buildVersionsTable(entries []versionedBuilderEntry) []byte {
	if len(entries) == 0 {
		return nil
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "<details>\n<summary>Show all %d versioned builders</summary>\n\n", len(entries))
	fmt.Fprintln(&b, "| Function | Step | Major |")
	fmt.Fprintln(&b, "|---|---|---|")
	for _, e := range entries {
		fmt.Fprintf(&b, "| `step.%s()` | `%s` | v%d.x |\n", e.funcName, e.stepID, e.major)
	}
	fmt.Fprint(&b, "\n</details>")
	return b.Bytes()
}

// replaceMarkersIfPresent is like replaceMarkers but silently returns src
// unchanged when the start marker is not found, instead of returning an error.
// This allows the versioned-builders section to be optional in the README.
func replaceMarkersIfPresent(src []byte, startMarker, endMarker string, replacement []byte) ([]byte, error) {
	if !bytes.Contains(src, []byte(startMarker)) {
		return src, nil
	}
	return replaceMarkers(src, startMarker, endMarker, replacement)
}

// replaceMarkers returns a copy of src where the lines between startMarker and
// endMarker (exclusive) are replaced by replacement. The markers themselves are
// preserved. Returns an error if either marker is not found.
func replaceMarkers(src []byte, startMarker, endMarker string, replacement []byte) ([]byte, error) {
	lines := strings.Split(string(src), "\n")

	startIdx, endIdx := -1, -1
	for i, line := range lines {
		if strings.TrimSpace(line) == startMarker {
			startIdx = i
		}
		if strings.TrimSpace(line) == endMarker {
			endIdx = i
		}
	}
	if startIdx < 0 {
		return nil, fmt.Errorf("marker %q not found", startMarker)
	}
	if endIdx < 0 {
		return nil, fmt.Errorf("marker %q not found", endMarker)
	}
	if endIdx <= startIdx {
		return nil, fmt.Errorf("end marker appears before start marker")
	}

	// Replacement lines (trim trailing newline from buildTable output, then split).
	replLines := strings.Split(strings.TrimRight(string(replacement), "\n"), "\n")

	var out []string
	out = append(out, lines[:startIdx+1]...)   // everything up to and including start marker
	out = append(out, replLines...)             // new table content
	out = append(out, lines[endIdx:]...)        // end marker and everything after
	return []byte(strings.Join(out, "\n")), nil
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "readmegen: "+format+"\n", args...)
	os.Exit(1)
}
