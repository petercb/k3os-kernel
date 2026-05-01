package main

import (
	"strings"
	"testing"
)

func TestLoadSymbolsFromReader(t *testing.T) {
	input := `ffffffff81001000 T _text
ffffffff81001000 T startup_64
ffffffff81001100 t some_local_func
`
	reader := strings.NewReader(input)
	loadSymbolsFromReader(reader)

	if !hasSymbol("_text") {
		t.Errorf("Expected symbol _text to be loaded")
	}
	if !hasSymbol("startup_64") {
		t.Errorf("Expected symbol startup_64 to be loaded")
	}
	if !hasSymbol("some_local_func") {
		t.Errorf("Expected symbol some_local_func to be loaded")
	}
}

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
	// Mock kallsyms with symbols for common features
	input := `ffffffff81001000 T veth_setup
ffffffff81001000 T br_init
ffffffff81001100 T fib_rules_register
ffffffff81001200 T uas_driver
ffffffff81001300 T vxlan_dev_setup
ffffffff81001400 T nf_register_net_hook
ffffffff81001500 T ipt_register_table
ffffffff81001600 T ipt_do_table
ffffffff81001700 T masquerade_tg_reg
ffffffff81001800 T nf_nat_masquerade_ipv4
ffffffff81001900 T comment_mt
ffffffff81001a00 T nvme_tcp_init
ffffffff81001b00 T vfio_pci_init
ffffffff81001c00 T uio_pci_generic_init
ffffffff81001d00 T loop_init
ffffffff81001e00 T hwrng_register
ffffffff81001f00 T tpm_chip_register
`
	// Reset symbols for this test
	symbols = make(map[string]bool)
	loadSymbolsFromReader(strings.NewReader(input))

	for _, f := range Features {
		// Only test ArchAll features (symbol-only, no sysfs path)
		if f.Arch != ArchAll || f.Path != "" || f.Disabled {
			continue
		}
		found := false
		for _, sym := range f.Symbols {
			if hasSymbol(sym) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Feature %s: Expected one of %v to be found", f.Name, f.Symbols)
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
	// Reset symbols
	symbols = make(map[string]bool)
	loadSymbolsFromReader(strings.NewReader("ffffffff81001000 T bad_symbol\n"))

	// A disabled feature should PASS when its symbols are NOT found
	disabledTest := FeatureTest{
		Name:     "Should Be Disabled",
		Symbols:  []string{"not_present_symbol"},
		Disabled: true,
	}
	found := false
	for _, sym := range disabledTest.Symbols {
		if hasSymbol(sym) {
			found = true
			break
		}
	}
	if found {
		t.Error("Disabled feature should pass when symbols are absent, but symbol was found")
	}

	// A disabled feature should FAIL when its symbols ARE found
	disabledTestPresent := FeatureTest{
		Name:     "Should Be Disabled But Present",
		Symbols:  []string{"bad_symbol"},
		Disabled: true,
	}
	found = false
	for _, sym := range disabledTestPresent.Symbols {
		if hasSymbol(sym) {
			found = true
			break
		}
	}
	if !found {
		t.Error("Disabled feature with present symbol should have been detected")
	}
}
