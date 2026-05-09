// Package main implements the fw-selector CLI tool.
// It replaces select_firmware.sh with a precise, config-aware firmware selector.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
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

	configFile, err := os.Open(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening config: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = configFile.Close() }()

	enabledConfigs, err := ParseKernelConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing config: %v\n", err)
		os.Exit(1)
	}

	knownFirmware := make(map[string]bool)
	if *whencePath != "" {
		var whenceFile *os.File
		whenceFile, err = os.Open(*whencePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fw-selector warning: could not open WHENCE file: %v\n", err)
		} else {
			defer func() { _ = whenceFile.Close() }()
			knownFirmware, err = ParseWhence(whenceFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "fw-selector warning: error parsing WHENCE file: %v\n", err)
			}
		}
	}

	selector := &Selector{
		EnabledConfigs: enabledConfigs,
		SourceDir:      *sourceDir,
		KnownFirmware:  knownFirmware,
		Arch:           *arch,
	}

	firmwareList, err := selector.SelectFirmware()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error selecting firmware: %v\n", err)
		os.Exit(1)
	}

	platformFw := GetPlatformFirmware(*arch)
	firmwareList = append(firmwareList, platformFw...)

	firmwareSet := make(map[string]bool)
	var uniqueFirmware []string
	for _, fw := range firmwareList {
		if !firmwareSet[fw] {
			firmwareSet[fw] = true
			uniqueFirmware = append(uniqueFirmware, fw)
		}
	}

	sort.Strings(uniqueFirmware)

	outFile, err := os.Create(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = outFile.Close() }()

	for _, fw := range uniqueFirmware {
		_, _ = fmt.Fprintln(outFile, fw)
	}
}
