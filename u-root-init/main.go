//go:build linux

package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
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

var (
	symbols map[string]bool
	results []TestResult
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

func main() {
	fmt.Println("--- Starting K3s-Ready Kernel Validation ---")

	// 1. Mount essential filesystems
	syscall.Mount("proc", "/proc", "proc", 0, "")
	syscall.Mount("sysfs", "/sys", "sysfs", 0, "")

	// Load symbols into memory once to speed up checks
	loadSymbols()

	// 2. Mount cgroup2
	os.MkdirAll("/sys/fs/cgroup", 0755)
	if err := syscall.Mount("none", "/sys/fs/cgroup", "cgroup2", 0, ""); err != nil {
		fmt.Println("[DEBUG] Failed to mount cgroup2:", err)
	}

	// 3. Run individual tests
	runTest("OverlayFS Support", func() (bool, string) {
		overlayFound := false
		if f, err := os.Open("/proc/filesystems"); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				if strings.Contains(scanner.Text(), "overlay") {
					overlayFound = true
					break
				}
			}
			f.Close()
		}
		if overlayFound {
			return true, ""
		}
		return false, "OverlayFS not found in /proc/filesystems"
	})

	runTest("Cgroup v2 Support", func() (bool, string) {
		if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err == nil {
			return true, ""
		}
		return false, "Cgroup v2 controllers not found or /sys/fs/cgroup not mounted"
	})

	runTest("Namespace Support", func() (bool, string) {
		if err := syscall.Unshare(syscall.CLONE_NEWUTS); err == nil {
			return true, ""
		} else {
			return false, fmt.Sprintf("Namespace unshare test failed: %v", err)
		}
	})

	runTest("USB Storage Support", func() (bool, string) {
		if _, err := os.Stat("/sys/bus/usb/drivers/usb-storage"); err == nil {
			return true, ""
		}
		return false, "USB storage driver not found in /sys/bus/usb/drivers"
	})

	// 4. Feature tests (via symbols or sysfs paths)
	features := []struct {
		name    string
		symbols []string
		path    string
	}{
		{"Veth Support", []string{"veth_setup"}, ""},
		{"Bridge Support", []string{"br_init"}, "/sys/module/bridge"},
		{"IP Advanced Router Support", []string{"fib_rules_register"}, ""},
		{"USB UAS Support", []string{"uas_driver"}, "/sys/bus/usb/drivers/uas"},
		{"VXLAN Support", []string{"vxlan_newlink"}, ""},
		{"Netfilter Support", []string{"nf_register_net_hooks"}, ""},
		{"IPTables Support", []string{"ipt_register_table", "ipt_do_table"}, ""},
		{"Netfilter Masquerade Support", []string{"masquerade_tg_reg", "nf_nat_masquerade_ipv4"}, ""},
		{"XT Match Comment Support", []string{"comment_mt"}, ""},
		{"NVMe over TCP Support", []string{"nvme_tcp_init"}, "/sys/module/nvme_tcp"},
		{"VFIO PCI Support", []string{"vfio_pci_init"}, "/sys/module/vfio_pci"},
		{"UIO PCI Generic Support", []string{"uio_pci_generic_init"}, "/sys/module/uio_pci_generic"},
	}

	for _, f := range features {
		f := f // capture range variable
		runTest(f.name, func() (bool, string) {
			if f.path != "" {
				if _, err := os.Stat(f.path); err == nil {
					return true, ""
				}
			}
			for _, sym := range f.symbols {
				if hasSymbol(sym) {
					return true, ""
				}
			}
			return false, "Required symbols or sysfs paths not found"
		})
	}

	validateArchSpecific()

	// Generate JUnit XML
	generateJUnit()

	// Check if all tests passed
	allPassed := true
	for _, res := range results {
		if !res.Passed {
			allPassed = false
			break
		}
	}

	if allPassed {
		fmt.Println("SUCCESS: Kernel booted and validation completed (u-root)")
	} else {
		fmt.Println("FAILURE: Some kernel validation tests FAILED")
	}

	// Direct syscall to power off the machine.
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)

	// Safety hang
	for {
		time.Sleep(time.Hour)
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

func loadSymbols() {
	symbols = make(map[string]bool)
	f, err := os.Open("/proc/kallsyms")
	if err != nil {
		fmt.Println("[DEBUG] Failed to open /proc/kallsyms:", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) >= 3 {
			symbols[parts[2]] = true
		}
	}
}

func hasSymbol(name string) bool {
	return symbols[name]
}
