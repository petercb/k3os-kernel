package main

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// configObjRe matches lines like: obj-$(CONFIG_FOO) += foo.o bar.o
// It also matches obj-y, obj-m, and := syntax.
var configObjRe = regexp.MustCompile(
	`^obj-\$\(CONFIG_([A-Z0-9_]+)\)\s*[:+]?=\s*(.*)`,
)

// builtinObjRe matches lines like: obj-y += foo.o
var builtinObjRe = regexp.MustCompile(
	`^obj-y\s*[:+]?=\s*(.*)`,
)

// moduleObjRe matches lines like: obj-m += foo.o
var moduleObjRe = regexp.MustCompile(
	`^obj-m\s*[:+]?=\s*(.*)`,
)

// multiPartRe matches lines like: foo-y := a.o b.o
// or foo-objs := a.o b.o
var multiPartRe = regexp.MustCompile(
	`^([a-zA-Z0-9_-]+)-(?:objs|y|m)\s*[:+]?=\s*(.*)`,
)

const (
	stateNone = iota
	stateConfig
	stateMultiPart
)

// appendObjects appends objects to the appropriate map based on state.
func appendObjects(state int, key string, objs []string, result, multiPart map[string][]string) {
	switch state {
	case stateConfig:
		result[key] = append(result[key], objs...)
	case stateMultiPart:
		multiPart[key] = append(multiPart[key], objs...)
	}
}

// parseLine matches a line against known patterns.
func parseLine(line string) (newState int, key string, raw string, matched bool) {
	if m := configObjRe.FindStringSubmatch(line); m != nil {
		return stateConfig, m[1], m[2], true
	}
	if m := builtinObjRe.FindStringSubmatch(line); m != nil {
		return stateConfig, "__BUILTIN__", m[1], true
	}
	if m := moduleObjRe.FindStringSubmatch(line); m != nil {
		return stateConfig, "__MODULE__", m[1], true
	}
	if m := multiPartRe.FindStringSubmatch(line); m != nil {
		return stateMultiPart, m[1], m[2], true
	}
	return stateNone, "", "", false
}

// ParseMakefile reads a kernel Makefile and returns a mapping from config key
// (without CONFIG_ prefix) to a list of .o object filenames. Lines with obj-y
// are mapped to the special key "__BUILTIN__" and obj-m to "__MODULE__".
// Subdirectory entries (ending with /) are excluded.
// Continuation lines (trailing \) are handled.
// It also resolves multi-part objects like foo-y := a.o b.o.
func ParseMakefile(r io.Reader) (map[string][]string, error) {
	result := make(map[string][]string)
	multiPart := make(map[string][]string)
	scanner := bufio.NewScanner(r)

	var currentKey string
	var state int
	var continuing bool

	for scanner.Scan() {
		line := scanner.Text()

		if continuing {
			objs := extractObjects(stripContinuation(line))
			appendObjects(state, currentKey, objs, result, multiPart)
			continuing = hasContinuation(line)
			continue
		}

		newState, key, raw, matched := parseLine(line)
		if matched {
			state = newState
			currentKey = key
			objs := extractObjects(stripContinuation(raw))
			appendObjects(state, currentKey, objs, result, multiPart)
			continuing = hasContinuation(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Resolve multi-part objects
	for key, objs := range result {
		var resolved []string
		for _, obj := range objs {
			baseName := strings.TrimSuffix(obj, ".o")
			if parts, ok := multiPart[baseName]; ok && len(parts) > 0 {
				resolved = append(resolved, parts...)
			} else {
				resolved = append(resolved, obj)
			}
		}
		result[key] = resolved
	}

	// Remove keys that ended up with no .o objects (e.g., subdir-only entries).
	for key, objs := range result {
		if len(objs) == 0 {
			delete(result, key)
		}
	}

	return result, nil
}

// extractObjects splits whitespace-separated tokens and returns only
// those ending in ".o" (filtering out subdirectory references like "foo/").
func extractObjects(s string) []string {
	var objs []string
	for _, token := range strings.Fields(s) {
		if strings.HasSuffix(token, ".o") {
			objs = append(objs, token)
		}
	}
	return objs
}

// hasContinuation returns true if the line ends with a backslash.
func hasContinuation(line string) bool {
	return strings.HasSuffix(strings.TrimRight(line, " \t"), "\\")
}

// stripContinuation removes trailing backslash and surrounding whitespace.
func stripContinuation(s string) string {
	s = strings.TrimRight(s, " \t")
	s = strings.TrimSuffix(s, "\\")
	return strings.TrimSpace(s)
}
