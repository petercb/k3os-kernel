package main

import (
	"fmt"
	"os"
)

func validateArchSpecific() {
	runTest("DRM V3D Support", func() (bool, string) {
		check := func() bool {
			if hasSymbol("v3d_v71_ops") || hasSymbol("v3d_driver") {
				return true
			}
			if _, err := os.Stat("/sys/module/v3d"); err == nil {
				return true
			}
			return false
		}

		if check() {
			return true, ""
		}

		if err := tryLoadModule("v3d"); err != nil {
			fmt.Printf("[DEBUG] v3d load failed: %v\n", err)
		} else if check() {
			return true, ""
		}

		return false, "DRM V3D support not found in kallsyms or /sys/module (modprobe v3d failed)"
	})

	runTest("MMC SDHCI IPROC Support", func() (bool, string) {
		if hasSymbol("sdhci_iproc_probe") || hasSymbol("sdhci_iproc_driver") {
			return true, ""
		}
		if _, err := os.Stat("/sys/module/sdhci_iproc"); err == nil {
			return true, ""
		}
		return false, "MMC SDHCI IPROC support not found in kallsyms or /sys/module"
	})

	runTest("MMC BCM2835 Support", func() (bool, string) {
		check := func() bool {
			if hasSymbol("bcm2835_probe") || hasSymbol("bcm2835_driver_init") || hasSymbol("bcm2835_mmc_probe") || hasSymbol("bcm2835_mmc_driver") || hasSymbol("mci_bcm2835_driver") || hasSymbol("bcm2835_mmc_irq") || hasSymbol("sdhci_bcm2835_probe") || hasSymbol("sdhci_bcm2835_ops") || hasSymbol("bcm2835_mmc_ops") || hasSymbol("bcm2835_mmc_pdata") || hasSymbol("bcm2835_sdhost_probe") || hasSymbol("bcm2835_sdhost_driver") {
				return true
			}
			if _, err := os.Stat("/sys/module/bcm2835_mmc"); err == nil {
				return true
			}
			if _, err := os.Stat("/sys/module/bcm2835_sdhost"); err == nil {
				return true
			}
			return false
		}

		if check() {
			return true, ""
		}

		// Try both possible module names
		_ = tryLoadModule("bcm2835_mmc")
		_ = tryLoadModule("bcm2835_sdhost")

		if check() {
			return true, ""
		}

		return false, "MMC BCM2835 support not found in kallsyms or /sys/module"
	})

	runTest("PINCTRL BCM2835 Support", func() (bool, string) {
		if hasSymbol("bcm2835_pinctrl_probe") {
			return true, ""
		}
		return false, "PINCTRL BCM2835 support not found in kallsyms"
	})

	runTest("PINCTRL ROCKCHIP Support", func() (bool, string) {
		if hasSymbol("rockchip_pinctrl_probe") {
			return true, ""
		}
		return false, "PINCTRL ROCKCHIP support not found in kallsyms"
	})

	runTest("BCM2835_WDT Support", func() (bool, string) {
		if hasSymbol("bcm2835_wdt_probe") || hasSymbol("bcm2835_wdt_init") {
			return true, ""
		}
		_ = tryLoadModule("bcm2835_wdt")
		if hasSymbol("bcm2835_wdt_probe") || hasSymbol("bcm2835_wdt_init") {
			return true, ""
		}
		return false, "BCM2835_WDT support not found"
	})

	runTest("DW_WATCHDOG Support", func() (bool, string) {
		if hasSymbol("dw_wdt_probe") || hasSymbol("dw_wdt_init") || hasSymbol("dw_wdt_drv_probe") {
			return true, ""
		}
		_ = tryLoadModule("dw_wdt")
		if hasSymbol("dw_wdt_probe") || hasSymbol("dw_wdt_init") || hasSymbol("dw_wdt_drv_probe") {
			return true, ""
		}
		return false, "DW_WATCHDOG support not found"
	})

	runTest("HW_RANDOM_BCM2835 Support", func() (bool, string) {
		if hasSymbol("bcm2835_rng_probe") || hasSymbol("bcm2835_rng_init") {
			return true, ""
		}
		_ = tryLoadModule("bcm2835-rng")
		if hasSymbol("bcm2835_rng_probe") || hasSymbol("bcm2835_rng_init") {
			return true, ""
		}
		return false, "HW_RANDOM_BCM2835 support not found"
	})

	runTest("HW_RANDOM_ROCKCHIP Support", func() (bool, string) {
		if hasSymbol("rockchip_rng_probe") || hasSymbol("rockchip_rng_init") {
			return true, ""
		}
		_ = tryLoadModule("rockchip-rng")
		if hasSymbol("rockchip_rng_probe") || hasSymbol("rockchip_rng_init") {
			return true, ""
		}
		return false, "HW_RANDOM_ROCKCHIP support not found"
	})

	runTest("HW_RANDOM_ARM_SMCCC_TRNG Support", func() (bool, string) {
		if hasSymbol("smccc_trng_probe") || hasSymbol("smccc_trng_init") {
			return true, ""
		}
		_ = tryLoadModule("smccc_trng")
		if hasSymbol("smccc_trng_probe") || hasSymbol("smccc_trng_init") {
			return true, ""
		}
		return false, "HW_RANDOM_ARM_SMCCC_TRNG support not found"
	})

	runTest("HW_RANDOM_IPROC_RNG200 Support", func() (bool, string) {
		if hasSymbol("iproc_rng200_probe") || hasSymbol("iproc_rng200_init") {
			return true, ""
		}
		_ = tryLoadModule("iproc-rng200")
		if hasSymbol("iproc_rng200_probe") || hasSymbol("iproc_rng200_init") {
			return true, ""
		}
		return false, "HW_RANDOM_IPROC_RNG200 support not found"
	})
}
