//go:build linux

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
)

func main() {
	fmt.Println("--- Starting K3s-Ready Kernel Validation ---")

	// 1. Mount essential filesystems
	syscall.Mount("proc", "/proc", "proc", 0, "")
	syscall.Mount("sysfs", "/sys", "sysfs", 0, "")

	// Mount cgroup2
	os.MkdirAll("/sys/fs/cgroup", 0755)
	if err := syscall.Mount("none", "/sys/fs/cgroup", "cgroup2", 0, ""); err != nil {
		fmt.Println("[DEBUG] Failed to mount cgroup2:", err)
	}
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
		fmt.Println("[PASS] OverlayFS support detected in /proc/filesystems")
	} else {
		fmt.Println("[FAIL] OverlayFS support MISSING")
	}

	// 3. Check for Cgroup v2 support
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err == nil {
		fmt.Println("[PASS] Cgroup v2 support detected in /sys/fs/cgroup")
	} else {
		fmt.Println("[FAIL] Cgroup v2 support MISSING (or /sys/fs/cgroup not mounted)")
	}

	// 4. Check for Namespace support (try to unshare)
	// We try to unshare the UTS namespace as a simple test
	if err := syscall.Unshare(syscall.CLONE_NEWUTS); err == nil {
		fmt.Println("[PASS] Namespace isolation (UTS) successfully tested via unshare")
	} else {
		fmt.Println("[FAIL] Namespace isolation test FAILED:", err)
	}

	// 5. Check for USB Storage support
	if _, err := os.Stat("/sys/bus/usb/drivers/usb-storage"); err == nil {
		fmt.Println("[PASS] USB Storage support detected")
	} else {
		fmt.Println("[FAIL] USB Storage support MISSING")
	}

	// 6. Check for Veth support
	if checkSymbols(" veth_setup") {
		fmt.Println("[PASS] Veth support detected (via kallsyms)")
	} else {
		fmt.Println("[FAIL] Veth support MISSING")
	}

	// 7. Check for Bridge support
	if _, err := os.Stat("/sys/module/bridge"); err == nil || checkSymbols(" br_init") {
		fmt.Println("[PASS] Bridge support detected")
	} else {
		fmt.Println("[FAIL] Bridge support MISSING")
	}

	// 8. Check for Advanced Router support
	if checkSymbols(" fib_rules_register") {
		fmt.Println("[PASS] IP Advanced Router support detected")
	} else {
		fmt.Println("[FAIL] IP Advanced Router support MISSING")
	}

	// 9. Check for USB UAS support
	if _, err := os.Stat("/sys/bus/usb/drivers/uas"); err == nil || checkSymbols(" uas_driver") {
		fmt.Println("[PASS] USB UAS support detected")
	} else {
		fmt.Println("[FAIL] USB UAS support MISSING")
	}

	// 10. Check for VXLAN support
	if checkSymbols(" vxlan_newlink") {
		fmt.Println("[PASS] VXLAN support detected")
	} else {
		fmt.Println("[FAIL] VXLAN support MISSING")
	}

	// 11. Check for Netfilter core support
	if checkSymbols(" nf_register_net_hooks") {
		fmt.Println("[PASS] Netfilter support detected")
	} else {
		fmt.Println("[FAIL] Netfilter support MISSING")
	}

	// 12. Check for IPTables support
	if checkSymbols(" ipt_register_table") || checkSymbols(" ipt_do_table") {
		fmt.Println("[PASS] IPTables support detected")
	} else {
		fmt.Println("[FAIL] IPTables support MISSING")
	}

	// 13. Check for Masquerade support
	if checkSymbols(" masquerade_tg_reg") || checkSymbols(" nf_nat_masquerade_ipv4") {
		fmt.Println("[PASS] Netfilter Masquerade support detected")
	} else {
		fmt.Println("[FAIL] Netfilter Masquerade support MISSING")
	}

	// 14. Check for XT Match Comment support
	if checkSymbols(" comment_mt") {
		fmt.Println("[PASS] Netfilter XT Match Comment support detected")
	} else {
		fmt.Println("[FAIL] Netfilter XT Match Comment support MISSING")
	}

	fmt.Println("SUCCESS: Kernel booted and validation completed (u-root)")

	// Direct syscall to power off the machine.
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)

	// Safety hang
	for {
		time.Sleep(time.Hour)
	}
}

func checkSymbols(pattern string) bool {
	f, err := os.Open("/proc/kallsyms")
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), pattern) {
			return true
		}
	}
	return false
}
