// Package main implements the fw-selector CLI tool.
// It replaces select_firmware.sh with a precise, config-aware firmware selector.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	configPath := flag.String("config", "/boot/config", "Path to kernel config file")
	sourceDir := flag.String("source-dir", "/usr/src/linux", "Path to kernel source tree")
	whencePath := flag.String("whence", "", "Path to WHENCE manifest from linux-firmware.git")
	arch := flag.String("arch", "", "Target architecture (amd64, arm64)")
	output := flag.String("output", "/boot/firmware-list.txt", "Output firmware list path")

	flag.Parse()

	if *arch == "" {
		fmt.Fprintln(os.Stderr, "error: --arch is required")
		os.Exit(1)
	}

	fmt.Printf("fw-selector: config=%s source=%s whence=%s arch=%s output=%s\n",
		*configPath, *sourceDir, *whencePath, *arch, *output)
}
