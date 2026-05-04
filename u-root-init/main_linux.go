//go:build linux

package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
)

func main() {
	fmt.Println("--- Starting K3s-Ready Kernel Validation ---")

	// 1. Mount essential filesystems
	_ = os.MkdirAll("/proc", 0o755)
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		fmt.Printf("[DEBUG] Failed to mount /proc: %v\n", err)
	}

	_ = os.MkdirAll("/sys", 0o755)
	if err := syscall.Mount("sysfs", "/sys", "sysfs", 0, ""); err != nil {
		fmt.Printf("[DEBUG] Failed to mount /sys: %v\n", err)
	}

	// Load configs
	kernelConfigs, err := LoadKernelConfigs()
	if err != nil {
		fmt.Printf("[WARN] %v\n", err)
	}

	// Check for modules directory
	kernelVersion := "unknown"
	if utsname, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
		kernelVersion = strings.TrimSpace(string(utsname))
	}
	modulePath := fmt.Sprintf("/lib/modules/%s", kernelVersion)
	if _, err := os.Stat(modulePath); err != nil {
		fmt.Printf("[DEBUG] Module directory %s not found: %v\n", modulePath, err)
	} else {
		fmt.Printf("[DEBUG] Found module directory: %s\n", modulePath)
		if _, err := os.Stat(modulePath + "/modules.dep"); err != nil {
			fmt.Printf("[WARN] modules.dep not found in %s. modprobe will fail.\n", modulePath)
		}
	}

	// 2. Mount cgroup2
	if err := os.MkdirAll("/sys/fs/cgroup", 0o755); err != nil {
		fmt.Printf("[DEBUG] Failed to create /sys/fs/cgroup: %v\n", err)
	}
	if err := syscall.Mount("none", "/sys/fs/cgroup", "cgroup2", 0, ""); err != nil {
		fmt.Println("[DEBUG] Failed to mount cgroup2:", err)
	}

	// 3. Run individual tests
	runTest("OverlayFS Support", func() (bool, string) {
		if f, err := os.Open("/proc/filesystems"); err == nil {
			defer func() { _ = f.Close() }()
			if hasFilesystem(f, "overlay") {
				return true, ""
			}
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

	// 4. Feature tests (via CONFIG check)
	for _, f := range Features {
		f := f // capture range variable

		// Skip tests not applicable to this architecture
		if f.Arch != ArchAll && f.Arch != currentArch {
			continue
		}

		runTest(f.Name, func() (bool, string) {
			if kernelConfigs == nil {
				return false, "Kernel configs could not be loaded"
			}

			val, ok := kernelConfigs[f.Config]
			if f.Disabled {
				// Assert the feature is NOT present
				if ok && val != "n" {
					return false, fmt.Sprintf("%s should be disabled but was found (val=%s)", f.Config, val)
				}
				return true, ""
			}

			if !ok {
				return false, fmt.Sprintf("Missing config %s", f.Config)
			}
			for _, a := range f.Allowed {
				if val == a {
					return true, ""
				}
			}
			return false, fmt.Sprintf("Incorrect config %s (got %s, expected %v)", f.Config, val, f.Allowed)
		})
	}

	// 5. Stress Testing
	runTest("Kernel Stress Test", func() (bool, string) {
		return RunStressTests()
	})

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
	if err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF); err != nil {
		fmt.Printf("[DEBUG] Failed to power off: %v\n", err)
	}

	// Safety hang
	for {
		time.Sleep(time.Hour)
	}
}
