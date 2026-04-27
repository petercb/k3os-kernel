package main

import "os"

func validateArchSpecific() {
	runTest("DRM V3D Support", func() (bool, string) {
		if hasSymbol("v3d_v71_ops") || hasSymbol("v3d_driver") {
			return true, ""
		}
		if _, err := os.Stat("/sys/module/v3d"); err == nil {
			return true, ""
		}
		return false, "DRM V3D support not found in kallsyms or /sys/module"
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
		if hasSymbol("bcm2835_probe") || hasSymbol("bcm2835_driver_init") || hasSymbol("bcm2835_mmc_probe") || hasSymbol("bcm2835_mmc_driver") || hasSymbol("mci_bcm2835_driver") || hasSymbol("bcm2835_mmc_irq") || hasSymbol("sdhci_bcm2835_probe") || hasSymbol("sdhci_bcm2835_ops") || hasSymbol("bcm2835_mmc_ops") || hasSymbol("bcm2835_mmc_pdata") || hasSymbol("bcm2835_sdhost_probe") || hasSymbol("bcm2835_sdhost_driver") {
			return true, ""
		}
		if _, err := os.Stat("/sys/module/bcm2835_mmc"); err == nil {
			return true, ""
		}
		if _, err := os.Stat("/sys/module/bcm2835_sdhost"); err == nil {
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
}
