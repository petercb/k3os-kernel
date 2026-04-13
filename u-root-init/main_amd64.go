package main

func validateArchSpecific() {
	runTest("HFS+ Support", func() (bool, string) {
		if hasSymbol("hfsplus_fill_super") {
			return true, ""
		}
		return false, "HFS+ filesystem support not found in kallsyms"
	})
}
