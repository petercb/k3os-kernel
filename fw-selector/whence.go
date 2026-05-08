package main

import (
	"bufio"
	"io"
	"regexp"
)

// whenceFileRe matches lines like: File: foo.bin or File: "foo bar.bin"
var whenceFileRe = regexp.MustCompile(`^File:\s*(?:"([^"]+)"|(\S+))`)

// whenceLinkRe matches lines like: Link: link.bin -> target.bin or Link: "link with spaces.bin" -> "target.bin"
// We only extract the link name, as that's the firmware path that would be requested.
var whenceLinkRe = regexp.MustCompile(`^Link:\s*(?:"([^"]+)"|([^\s:]+))\s*->`)

// ParseWhence reads a linux-firmware WHENCE file and returns a set of known
// firmware paths (both regular files and symlinks). It handles quotes for
// paths with spaces.
func ParseWhence(r io.Reader) (map[string]bool, error) {
	known := make(map[string]bool)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		if m := whenceFileRe.FindStringSubmatch(line); m != nil {
			if m[1] != "" {
				known[m[1]] = true
			} else {
				known[m[2]] = true
			}
			continue
		}

		if m := whenceLinkRe.FindStringSubmatch(line); m != nil {
			if m[1] != "" {
				known[m[1]] = true
			} else {
				known[m[2]] = true
			}
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return known, nil
}
