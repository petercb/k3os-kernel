package main

import (
	"bufio"
	"io"
	"regexp"
)

// moduleFirmwareRe matches MODULE_FIRMWARE("literal/path") declarations.
// It only captures literal string arguments, not macro-expanded paths.
var moduleFirmwareRe = regexp.MustCompile(
	`MODULE_FIRMWARE\s*\(\s*"([^"]+)"\s*\)`,
)

// ExtractModuleFirmware scans a C source file for MODULE_FIRMWARE("...")
// declarations and returns the firmware paths. Macro-expanded paths
// (e.g., MODULE_FIRMWARE(MACRO_NAME)) are silently skipped.
func ExtractModuleFirmware(r io.Reader) ([]string, error) {
	var paths []string
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		matches := moduleFirmwareRe.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			paths = append(paths, m[1])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return paths, nil
}
