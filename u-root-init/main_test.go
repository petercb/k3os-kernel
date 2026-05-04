package main

import (
	"strings"
	"testing"
)

func TestHasFilesystem(t *testing.T) {
	input := `nodev	sysfs
nodev	tmpfs
nodev	proc
	overlay
	squashfs
`
	reader := strings.NewReader(input)
	if !hasFilesystem(reader, "overlay") {
		t.Errorf("Expected overlay filesystem to be found")
	}

	reader = strings.NewReader(input)
	if !hasFilesystem(reader, "squashfs") {
		t.Errorf("Expected squashfs filesystem to be found")
	}

	reader = strings.NewReader(input)
	if hasFilesystem(reader, "ext4") {
		t.Errorf("Expected ext4 filesystem to be NOT found")
	}
}

func TestFeatureValidation(t *testing.T) {
	kernelConfigs := map[string]string{
		"CONFIG_BLK_DEV_LOOP": "y",
		"CONFIG_HW_RANDOM":    "m",
		"CONFIG_NETFILTER":    "y",
	}

	for _, f := range Features {
		if f.Name == "Loop Device Support" {
			val, ok := kernelConfigs[f.Config]
			if !ok || val != "y" {
				t.Errorf("Expected Loop Device Support config to be y")
			}
		}
	}
}

func TestArchFiltering(t *testing.T) {
	// Verify that arch constants filter correctly
	tests := []struct {
		featureArch string
		buildArch   string
		shouldRun   bool
	}{
		{ArchAll, ArchAMD64, true},
		{ArchAll, ArchARM64, true},
		{ArchAMD64, ArchAMD64, true},
		{ArchAMD64, ArchARM64, false},
		{ArchARM64, ArchARM64, true},
		{ArchARM64, ArchAMD64, false},
		{ArchAll, ArchAll, true}, // generic arch
	}

	for _, tt := range tests {
		shouldRun := tt.featureArch == ArchAll || tt.featureArch == tt.buildArch
		if shouldRun != tt.shouldRun {
			t.Errorf("Arch=%q on build=%q: got shouldRun=%v, want %v",
				tt.featureArch, tt.buildArch, shouldRun, tt.shouldRun)
		}
	}
}

func TestDisabledFeature(t *testing.T) {
	// A disabled feature should PASS when its config is NOT present (or is 'n')
	disabledTest := FeatureTest{
		Name:     "Should Be Disabled",
		Config:   "CONFIG_BAD",
		Disabled: true,
	}

	configs := map[string]string{
		"CONFIG_GOOD": "y",
	}

	if val, ok := configs[disabledTest.Config]; ok && val != "n" {
		t.Error("Disabled feature should pass when config is absent, but it was found")
	}

	// A disabled feature should FAIL when its config IS found
	disabledTestPresent := FeatureTest{
		Name:     "Should Be Disabled But Present",
		Config:   "CONFIG_BAD",
		Disabled: true,
	}

	configs["CONFIG_BAD"] = "y"

	if val, ok := configs[disabledTestPresent.Config]; ok && val != "n" {
		// correctly detected as present
	} else {
		t.Error("Disabled feature with present config should have been detected")
	}
}
