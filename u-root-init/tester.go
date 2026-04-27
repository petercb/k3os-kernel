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
}

var (
	results  []TestResult
	Features = []FeatureTest{
		{"Veth Support", []string{"veth_setup"}, ""},
		{"Bridge Support", []string{"br_init"}, "/sys/module/bridge"},
		{"IP Advanced Router Support", []string{"fib_rules_register"}, ""},
		{"USB UAS Support", []string{"uas_driver"}, "/sys/bus/usb/drivers/uas"},
		{"VXLAN Support", []string{"vxlan_newlink", "vxlan_dev_setup", "vxlan_validate"}, ""},
		{"Netfilter Support", []string{"nf_register_net_hooks", "nf_register_net_hook", "nf_register_hook"}, ""},
		{"IPTables/NFTables Support", []string{"ipt_register_table", "ipt_do_table", "xt_register_table", "nft_register_table", "nft_do_chain"}, ""},
		{"Netfilter Masquerade Support", []string{"masquerade_tg_reg", "nf_nat_masquerade_ipv4", "nf_nat_masquerade"}, ""},
		{"XT Match Comment Support", []string{"comment_mt"}, ""},
		{"NVMe over TCP Support", []string{"nvme_tcp_init"}, "/sys/module/nvme_tcp"},
		{"VFIO PCI Support", []string{"vfio_pci_init"}, "/sys/module/vfio_pci"},
		{"UIO PCI Generic Support", []string{"uio_pci_generic_init"}, "/sys/module/uio_pci_generic"},
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
