package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Selector orchestrates the process of finding firmware for enabled kernel modules.
type Selector struct {
	EnabledConfigs map[string]bool
	SourceDir      string
	KnownFirmware  map[string]bool // parsed from WHENCE
	Arch           string
}

// SelectFirmware walks the kernel source tree, finds Makefiles, maps enabled
// configs to source files, extracts MODULE_FIRMWARE declarations, and validates
// them against the WHENCE manifest.
func (s *Selector) SelectFirmware() ([]string, error) {
	var firmwareList []string

	err := filepath.WalkDir(s.SourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || d.Name() != "Makefile" {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		makefileMap, err := ParseMakefile(f)
		_ = f.Close()
		if err != nil {
			return err
		}

		fwPaths, err := s.processMakefile(filepath.Dir(path), makefileMap)
		if err != nil {
			return err
		}
		firmwareList = append(firmwareList, fwPaths...)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return firmwareList, nil
}

func (s *Selector) processMakefile(dir string, makefileMap map[string][]string) ([]string, error) {
	var fwList []string

	for config, objs := range makefileMap {
		if config != "__BUILTIN__" && config != "__MODULE__" && !s.EnabledConfigs[config] {
			continue
		}

		for _, obj := range objs {
			fwPaths, err := s.processObject(dir, obj)
			if err != nil {
				return nil, err
			}
			fwList = append(fwList, fwPaths...)
		}
	}

	return fwList, nil
}

func (s *Selector) processObject(dir, obj string) ([]string, error) {
	srcName := strings.TrimSuffix(obj, ".o") + ".c"
	srcPath := filepath.Join(dir, srcName)

	srcFile, err := os.Open(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	fwPaths, err := ExtractModuleFirmware(srcFile)
	_ = srcFile.Close()
	if err != nil {
		return nil, err
	}

	var validPaths []string
	for _, fw := range fwPaths {
		if len(s.KnownFirmware) == 0 || s.KnownFirmware[fw] {
			validPaths = append(validPaths, fw)
		} else {
			fmt.Fprintf(os.Stderr, "fw-selector warning: firmware %q in %s not found in WHENCE\n", fw, srcPath)
		}
	}

	return validPaths, nil
}
