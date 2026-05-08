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

// ParseMakefile reads a kernel Makefile and returns a mapping from config key
// (without CONFIG_ prefix) to a list of .o object filenames. Lines with obj-y
// are mapped to the special key "__BUILTIN__" and obj-m to "__MODULE__".
// Subdirectory entries (ending with /) are excluded.
// Continuation lines (trailing \) are handled.
func ParseMakefile(r io.Reader) (map[string][]string, error) {
	result := make(map[string][]string)
	scanner := bufio.NewScanner(r)

	var currentKey string
	var continuing bool

	for scanner.Scan() {
		line := scanner.Text()

		if continuing {
			objs := extractObjects(stripContinuation(line))
			result[currentKey] = append(result[currentKey], objs...)
			continuing = hasContinuation(line)
			continue
		}

		if m := configObjRe.FindStringSubmatch(line); m != nil {
			currentKey = m[1]
			raw := stripContinuation(m[2])
			objs := extractObjects(raw)
			result[currentKey] = append(result[currentKey], objs...)
			continuing = hasContinuation(line)
			continue
		}

		if m := builtinObjRe.FindStringSubmatch(line); m != nil {
			currentKey = "__BUILTIN__"
			raw := stripContinuation(m[1])
			objs := extractObjects(raw)
			result[currentKey] = append(result[currentKey], objs...)
			continuing = hasContinuation(line)
			continue
		}

		if m := moduleObjRe.FindStringSubmatch(line); m != nil {
			currentKey = "__MODULE__"
			raw := stripContinuation(m[1])
			objs := extractObjects(raw)
			result[currentKey] = append(result[currentKey], objs...)
			continuing = hasContinuation(line)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
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
