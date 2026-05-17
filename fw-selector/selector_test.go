package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSelectorSelectFirmware(t *testing.T) {
	// Create a temporary directory to act as the kernel source tree
	tmpDir, err := os.MkdirTemp("", "fw-selector-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Setup mock source tree
	// driver_a is enabled and has firmware "a.bin"
	// driver_b is disabled but has firmware "b.bin" in the same directory
	// driver_c is enabled but its firmware "c.bin" is not in WHENCE
	setupMockSource(t, tmpDir)

	selector := &Selector{
		EnabledConfigs: map[string]bool{
			"DRIVER_A": true,
			"DRIVER_C": true,
		},
		SourceDir: tmpDir,
		KnownFirmware: map[string]bool{
			"a.bin": true,
			"b.bin": true,
		},
		Arch: "arm64", // doesn't matter for this core test as platform fw is appended separately in main
	}

	got, err := selector.SelectFirmware()
	if err != nil {
		t.Fatalf("SelectFirmware() error = %v", err)
	}

	// We expect exactly "a.bin"
	// "b.bin" should be excluded because DRIVER_B is not enabled
	// "c.bin" should be excluded because it's not in KnownFirmware (WHENCE)
	want := []string{"a.bin"}

	assertStringSlice(t, got, want)
}

func setupMockSource(t *testing.T, dir string) {
	t.Helper()

	// Create drivers/net
	netDir := filepath.Join(dir, "drivers", "net")
	if err := os.MkdirAll(netDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	// Create drivers/net/Makefile
	makefileContent := `
obj-$(CONFIG_DRIVER_A) += driver_a.o
obj-$(CONFIG_DRIVER_B) += driver_b.o
obj-$(CONFIG_DRIVER_C) += driver_c.o
`
	if err := os.WriteFile(filepath.Join(netDir, "Makefile"), []byte(makefileContent), 0o644); err != nil {
		t.Fatalf("write Makefile failed: %v", err)
	}

	// Create driver_a.c
	driverAContent := `
#include <linux/module.h>
MODULE_FIRMWARE("a.bin");
`
	if err := os.WriteFile(filepath.Join(netDir, "driver_a.c"), []byte(driverAContent), 0o644); err != nil {
		t.Fatalf("write driver_a.c failed: %v", err)
	}

	// Create driver_b.c
	driverBContent := `
#include <linux/module.h>
MODULE_FIRMWARE("b.bin");
`
	if err := os.WriteFile(filepath.Join(netDir, "driver_b.c"), []byte(driverBContent), 0o644); err != nil {
		t.Fatalf("write driver_b.c failed: %v", err)
	}

	// Create driver_c.c
	driverCContent := `
#include <linux/module.h>
MODULE_FIRMWARE("c.bin");
`
	if err := os.WriteFile(filepath.Join(netDir, "driver_c.c"), []byte(driverCContent), 0o644); err != nil {
		t.Fatalf("write driver_c.c failed: %v", err)
	}
}
