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

	fmt.Println("SUCCESS: Kernel booted and validation completed (u-root)")

	// Direct syscall to power off the machine.
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)

	// Safety hang
	for {
		time.Sleep(time.Hour)
	}
}
