//go:build linux

package main

import (
	"fmt"
	"syscall"
	"time"
)

func main() {
	fmt.Println("SUCCESS: Kernel booted and executed init (u-root)")

	// Direct syscall to power off the machine.
	// This is the most reliable way for a PID 1 to signal shutdown.
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)

	// If the syscall returns, it means it failed or is delayed.
	// We MUST hang to avoid a kernel panic from PID 1 exiting.
	fmt.Println("Poweroff failed or delayed, sleeping to avoid panic...")
	for {
		time.Sleep(time.Hour)
	}
}
