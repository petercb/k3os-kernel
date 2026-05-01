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

	// Load symbols into memory once to speed up checks
	loadSymbols()
	if !symbolsLoaded() {
		fmt.Println("[WARN] /proc/kallsyms is empty or could not be read.")
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

	// 4. Feature tests (via symbols or sysfs paths)
	for _, f := range Features {
		f := f // capture range variable

		// Skip tests not applicable to this architecture
		if f.Arch != ArchAll && f.Arch != currentArch {
			continue
		}

		runTest(f.Name, func() (bool, string) {
			check := func() bool {
				if f.Path != "" {
					if _, err := os.Stat(f.Path); err == nil {
						return true
					}
				}
				for _, sym := range f.Symbols {
					if hasSymbol(sym) {
						return true
					}
				}
				return false
			}

			if f.Disabled {
				// Assert the feature is NOT present
				if check() {
					return false, fmt.Sprintf("%s should be disabled but was found", f.Name)
				}
				return true, ""
			}

			// First check
			if check() {
				return true, ""
			}

			// If failed, try loading module and re-check
			if f.Module != "" {
				if err := tryLoadModule(f.Module); err == nil {
					if check() {
						return true, ""
					}
				} else {
					fmt.Printf("[DEBUG] modprobe %s failed: %v\n", f.Module, err)
				}
			}

			return false, "Required symbols or sysfs paths not found (even after modprobe if applicable)"
		})
	}

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

func loadSymbols() {
	f, err := os.Open("/proc/kallsyms")
	if err != nil {
		fmt.Println("[DEBUG] Failed to open /proc/kallsyms:", err)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("[DEBUG] Failed to close /proc/kallsyms: %v\n", err)
		}
	}()
	loadSymbolsFromReader(f)
}
