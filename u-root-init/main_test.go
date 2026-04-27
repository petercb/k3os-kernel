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

func TestFeatureValidation_Linux7(t *testing.T) {
	// Mock kallsyms with symbols that might have changed in "Linux 7.0"
	// For example, nf_register_net_hooks (plural) -> nf_register_net_hook (singular)
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
`
	reader := strings.NewReader(input)
	// Reset symbols for this test
	symbols = make(map[string]bool)
	loadSymbolsFromReader(reader)

	for _, f := range Features {
		if f.Path != "" {
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
