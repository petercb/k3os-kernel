package main

import (
	"bufio"
	"io"
	"strings"
)

var symbols map[string]bool

func init() {
	symbols = make(map[string]bool)
}

func loadSymbolsFromReader(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) >= 3 {
			symbols[parts[2]] = true
		}
	}
}

func hasFilesystem(r io.Reader, name string) bool {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), name) {
			return true
		}
	}
	return false
}

func hasSymbol(name string) bool {
	return symbols[name]
}
