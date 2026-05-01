package main

func validateArchSpecific() {
	runTest("HFS+ Support", func() (bool, string) {
		if hasSymbol("hfsplus_fill_super") {
			return true, ""
		}
		return false, "HFS+ filesystem support not found in kallsyms"
	})

	runTest("NVRAM Support", func() (bool, string) {
		if hasSymbol("nvram_init") || hasSymbol("nvram_read_byte") {
			return true, ""
		}
		return false, "NVRAM support not found"
	})

	runTest("ITCO_WDT Support", func() (bool, string) {
		if hasSymbol("iTCO_wdt_init") || hasSymbol("iTCO_wdt_probe") {
			return true, ""
		}
		if err := tryLoadModule("iTCO_wdt"); err == nil {
			if hasSymbol("iTCO_wdt_init") || hasSymbol("iTCO_wdt_probe") {
				return true, ""
			}
		}
		return false, "ITCO_WDT support not found"
	})

	runTest("IT87_WDT Support", func() (bool, string) {
		if hasSymbol("it87_wdt_init") || hasSymbol("it87_wdt_probe") {
			return true, ""
		}
		if err := tryLoadModule("it87_wdt"); err == nil {
			if hasSymbol("it87_wdt_init") || hasSymbol("it87_wdt_probe") {
				return true, ""
			}
		}
		return false, "IT87_WDT support not found"
	})

	runTest("HW_RANDOM_AMD Support", func() (bool, string) {
		if hasSymbol("amd_rng_init") || hasSymbol("amd_rng_probe") {
			return true, ""
		}
		if err := tryLoadModule("amd-rng"); err == nil {
			if hasSymbol("amd_rng_init") || hasSymbol("amd_rng_probe") {
				return true, ""
			}
		}
		return false, "HW_RANDOM_AMD support not found"
	})

	runTest("HW_RANDOM_INTEL Support", func() (bool, string) {
		if hasSymbol("intel_rng_init") || hasSymbol("intel_rng_probe") {
			return true, ""
		}
		if err := tryLoadModule("intel-rng"); err == nil {
			if hasSymbol("intel_rng_init") || hasSymbol("intel_rng_probe") {
				return true, ""
			}
		}
		return false, "HW_RANDOM_INTEL support not found"
	})
}
