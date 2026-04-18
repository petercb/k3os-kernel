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

	runTest("Veth Support", func() (bool, string) {
		if hasSymbol("veth_setup") {
			return true, ""
		}
		return false, "Veth driver not found in kallsyms"
	})

	runTest("Bridge Support", func() (bool, string) {
		if _, err := os.Stat("/sys/module/bridge"); err == nil || hasSymbol("br_init") {
			return true, ""
		}
		return false, "Bridge driver not found in /sys/module or kallsyms"
	})

	runTest("IP Advanced Router Support", func() (bool, string) {
		if hasSymbol("fib_rules_register") {
			return true, ""
		}
		return false, "IP Advanced Router support not found in kallsyms"
	})

	runTest("USB UAS Support", func() (bool, string) {
		if _, err := os.Stat("/sys/bus/usb/drivers/uas"); err == nil || hasSymbol("uas_driver") {
			return true, ""
		}
		return false, "USB UAS driver not found in /sys/bus/usb/drivers or kallsyms"
	})

	runTest("VXLAN Support", func() (bool, string) {
		if hasSymbol("vxlan_newlink") {
			return true, ""
		}
		return false, "VXLAN driver not found in kallsyms"
	})

	runTest("Netfilter Support", func() (bool, string) {
		if hasSymbol("nf_register_net_hooks") {
			return true, ""
		}
		return false, "Netfilter core support not found in kallsyms"
	})

	runTest("IPTables Support", func() (bool, string) {
		if hasSymbol("ipt_register_table") || hasSymbol("ipt_do_table") {
			return true, ""
		}
		return false, "IPTables support not found in kallsyms"
	})

	runTest("Netfilter Masquerade Support", func() (bool, string) {
		if hasSymbol("masquerade_tg_reg") || hasSymbol("nf_nat_masquerade_ipv4") {
			return true, ""
		}
		return false, "Netfilter Masquerade support not found in kallsyms"
	})

	runTest("Netfilter XT Match Comment Support", func() (bool, string) {
		if hasSymbol("comment_mt") {
			return true, ""
		}
		return false, "Netfilter XT Match Comment support not found in kallsyms"
	})

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
