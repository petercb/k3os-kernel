package main

import (
	"bufio"
	"io"
	"strings"
)

// ParseKernelConfig reads a kernel config file and returns a set of enabled
// config keys (stripped of the CONFIG_ prefix). Only configs set to "y" or "m"
// are included. String values, numeric values, and disabled configs are excluded.
func ParseKernelConfig(r io.Reader) (map[string]bool, error) {
	enabled := make(map[string]bool)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Expect lines like: CONFIG_FOO=y or CONFIG_BAR=m
		if !strings.HasPrefix(line, "CONFIG_") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		value := parts[1]
		if value != "y" && value != "m" {
			continue
		}

		key := strings.TrimPrefix(parts[0], "CONFIG_")
		enabled[key] = true
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return enabled, nil
}
