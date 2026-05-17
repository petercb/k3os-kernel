package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration(t *testing.T) {
	// 1. Build the binary
	cmd := exec.Command("go", "build", "-o", "fw-selector-bin", ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}
	defer func() { _ = os.Remove("fw-selector-bin") }()

	tmpDir, err := os.MkdirTemp("", "fw-selector-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// 2. Setup mock source tree
	setupMockSource(t, tmpDir)

	// 3. Create config file
	configFile := filepath.Join(tmpDir, "config")
	configContent := `
CONFIG_DRIVER_A=y
CONFIG_DRIVER_B=m
# CONFIG_DRIVER_C is not set
`
	err = os.WriteFile(configFile, []byte(configContent), 0o644)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// 4. Create WHENCE file
	whenceFile := filepath.Join(tmpDir, "WHENCE")
	whenceContent := `
File: a.bin
File: b.bin
File: c.bin
`
	err = os.WriteFile(whenceFile, []byte(whenceContent), 0o644)
	if err != nil {
		t.Fatalf("failed to write WHENCE: %v", err)
	}

	// 5. Run the binary
	outPath := filepath.Join(tmpDir, "output.txt")
	runCmd := exec.Command("./fw-selector-bin",
		"--config", configFile,
		"--source-dir", tmpDir,
		"--whence", whenceFile,
		"--arch", "arm64",
		"--output", outPath,
	)

	var stderr bytes.Buffer
	runCmd.Stderr = &stderr
	err = runCmd.Run()
	if err != nil {
		t.Fatalf("failed to run fw-selector: %v\nstderr: %s", err, stderr.String())
	}

	// 6. Verify output
	outData, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	// For arm64, we expect a.bin, b.bin, and the platform firmwares.
	// We do NOT expect c.bin because it is not enabled in the config.
	lines := strings.Split(strings.TrimSpace(string(outData)), "\n")

	expectedMap := map[string]bool{
		"a.bin":                             true,
		"b.bin":                             true,
		"brcm/brcmfmac43455-sdio.bin":       true,
		"brcm/brcmfmac43455-sdio.txt":       true,
		"brcm/brcmfmac43455-sdio.clm_blob":  true,
		"brcm/brcmfmac43456-sdio.bin":       true,
		"brcm/brcmfmac43456-sdio.txt":       true,
		"brcm/brcmfmac43456-sdio.clm_blob":  true,
		"cypress/cyfmac43455-sdio.bin":      true,
		"cypress/cyfmac43455-sdio.clm_blob": true,
		"raspberrypi/bootloader-2711/latest/pieeprom-2711-latest.bin": true,
		"raspberrypi/bootloader-2712/latest/pieeprom-2712-latest.bin": true,
		"rockchip/dptx.bin": true,
	}

	for _, line := range lines {
		if !expectedMap[line] {
			t.Errorf("unexpected firmware in output: %q", line)
		}
		delete(expectedMap, line)
	}

	for missing := range expectedMap {
		t.Errorf("expected firmware missing from output: %q", missing)
	}
}
