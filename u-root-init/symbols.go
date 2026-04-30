package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"pault.ag/go/modprobe"
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

func symbolsLoaded() bool {
	return len(symbols) > 0
}

// tryLoadModule attempts to load a kernel module by name.
// It re-scans symbols if the load was successful.
func tryLoadModule(name string) error {
	if name == "" {
		return nil
	}

	fmt.Printf("[DEBUG] Attempting to modprobe: %s\n", name)
	err := modprobe.Load(name, "")
	if err == nil {
		fmt.Printf("[DEBUG] modprobe %s successful\n", name)
		return reloadSymbols()
	}

	fmt.Printf("[DEBUG] modprobe %s failed: %v\n", name, err)
	return err
}

func reloadSymbols() error {
	if f, err := os.Open("/proc/kallsyms"); err == nil {
		loadSymbolsFromReader(f)
		_ = f.Close()
	}
	return nil
}
