package main

import (
	"encoding/xml"
	"fmt"
	"os"
)

type TestResult struct {
	Name    string
	Passed  bool
	Message string
}

type JUnitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	TestCases []JUnitTestCase `xml:"testcase"`
}

type JUnitTestCase struct {
	Name    string        `xml:"name,attr"`
	Failure *JUnitFailure `xml:"failure,omitempty"`
}

type JUnitFailure struct {
	Message string `xml:"message,attr"`
}

type FeatureTest struct {
	Name    string
	Symbols []string
	Path    string
	Module  string // New: kernel module name to try loading
}

var (
	results  []TestResult
	Features = []FeatureTest{
		{Name: "Veth Support", Symbols: []string{"veth_setup"}, Path: "", Module: ""},
		{Name: "Bridge Support", Symbols: []string{"br_init"}, Path: "/sys/module/bridge", Module: "bridge"},
		{Name: "IP Advanced Router Support", Symbols: []string{"fib_rules_register"}, Path: "", Module: ""},
		{Name: "USB UAS Support", Symbols: []string{"uas_driver"}, Path: "/sys/bus/usb/drivers/uas", Module: "uas"},
		{Name: "VXLAN Support", Symbols: []string{"vxlan_dev_setup", "vxlan_newlink"}, Path: "", Module: "vxlan"},
		{Name: "Netfilter Support", Symbols: []string{"nf_register_net_hook", "nf_register_net_hooks"}, Path: "", Module: ""},
		{Name: "IPTables/NFTables Support", Symbols: []string{"ipt_register_table", "ipt_do_table", "nft_do_chain"}, Path: "", Module: ""},
		{Name: "Netfilter Masquerade Support", Symbols: []string{"masquerade_tg_reg", "nf_nat_masquerade_ipv4"}, Path: "", Module: ""},
		{Name: "XT Match Comment Support", Symbols: []string{"comment_mt"}, Path: "", Module: ""},
		{Name: "NVMe over TCP Support", Symbols: []string{"nvme_tcp_init"}, Path: "/sys/module/nvme_tcp", Module: "nvme_tcp"},
		{Name: "VFIO PCI Support", Symbols: []string{"vfio_pci_init"}, Path: "/sys/module/vfio_pci", Module: "vfio_pci"},
		{Name: "UIO PCI Generic Support", Symbols: []string{"uio_pci_generic_init"}, Path: "/sys/module/uio_pci_generic", Module: "uio_pci_generic"},
	}
)

func runTest(name string, check func() (bool, string)) {
	passed, msg := check()
	results = append(results, TestResult{Name: name, Passed: passed, Message: msg})
	if passed {
		fmt.Printf("[PASS] %s\n", name)
	} else {
		fmt.Printf("[FAIL] %s: %s\n", name, msg)
	}
}

func generateJUnit() {
	suite := JUnitTestSuite{
		Name:  "kernel-boot",
		Tests: len(results),
	}

	for _, res := range results {
		tc := JUnitTestCase{Name: res.Name}
		if !res.Passed {
			tc.Failure = &JUnitFailure{Message: res.Message}
		}
		suite.TestCases = append(suite.TestCases, tc)
	}

	fmt.Println("--- JUNIT START ---")
	fmt.Print(xml.Header)
	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("  ", "  ")
	if err := enc.Encode(suite); err != nil {
		fmt.Printf("DEBUG: Failed to encode JUnit XML: %v\n", err)
	}
	fmt.Println("\n--- JUNIT END ---")
}
