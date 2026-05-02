package main

import (
	"encoding/xml"
	"fmt"
	"os"
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

// Arch constants used to filter which tests run on which architecture.
const (
	ArchAll   = ""      // runs on all architectures
	ArchAMD64 = "amd64" // runs only on amd64
	ArchARM64 = "arm64" // runs only on arm64
)

// FeatureTest describes a kernel feature to validate at boot.
//
// Fields:
//   - Name:     Human-readable test name.
//   - Symbols:  Kernel symbols to search for in /proc/kallsyms (any match = pass).
//   - Path:     Optional sysfs/procfs path; if it exists, the test passes.
//   - Module:   Optional kernel module to try loading if the initial check fails.
//   - Arch:     Architecture filter ("amd64", "arm64", or "" for all).
//   - Disabled: If true, assert the feature is ABSENT (none of the symbols/paths should exist).
type FeatureTest struct {
	Name     string
	Symbols  []string
	Path     string
	Module   string
	Arch     string
	Disabled bool
}

var (
	results  []TestResult
	Features = []FeatureTest{
		// === All architectures ===
		{Name: "Loop Device Support", Symbols: []string{"loop_init", "loop_configure"}},
		{Name: "HW Random Support", Symbols: []string{"hwrng_register"}},
		{Name: "TPM Support", Symbols: []string{"tpm_chip_register", "tpm_init"}},
		{Name: "Veth Support", Symbols: []string{"veth_setup"}},
		{Name: "Bridge Support", Symbols: []string{"br_init"}, Path: "/sys/module/bridge", Module: "bridge"},
		{Name: "IP Advanced Router Support", Symbols: []string{"fib_rules_register"}},
		{Name: "USB UAS Support", Symbols: []string{"uas_driver"}, Path: "/sys/bus/usb/drivers/uas", Module: "uas"},
		{Name: "VXLAN Support", Symbols: []string{"vxlan_dev_setup", "vxlan_newlink"}, Module: "vxlan"},
		{Name: "Netfilter Support", Symbols: []string{"nf_register_net_hook", "nf_register_net_hooks"}},
		{Name: "IPTables/NFTables Support", Symbols: []string{"ipt_register_table", "ipt_do_table", "nft_do_chain"}},
		{Name: "Netfilter Masquerade Support", Symbols: []string{"masquerade_tg_reg", "nf_nat_masquerade_ipv4"}},
		{Name: "XT Match Comment Support", Symbols: []string{"comment_mt"}},
		{Name: "NVMe over TCP Support", Symbols: []string{"nvme_tcp_init"}, Path: "/sys/module/nvme_tcp", Module: "nvme_tcp"},
		{Name: "VFIO PCI Support", Symbols: []string{"vfio_pci_init"}, Path: "/sys/module/vfio_pci", Module: "vfio_pci"},
		{Name: "UIO PCI Generic Support", Symbols: []string{"uio_pci_generic_init"}, Path: "/sys/module/uio_pci_generic", Module: "uio_pci_generic"},

		// === AMD64-only ===
		{Name: "HFS+ Support", Symbols: []string{"hfsplus_fill_super"}, Arch: ArchAMD64},
		{Name: "NVRAM Support", Symbols: []string{"nvram_init", "nvram_read_byte"}, Arch: ArchAMD64},
		{Name: "ITCO_WDT Support", Symbols: []string{"iTCO_wdt_init", "iTCO_wdt_probe"}, Module: "iTCO_wdt", Arch: ArchAMD64},
		{Name: "IT87_WDT Support", Symbols: []string{"it87_wdt_init", "it87_wdt_probe"}, Module: "it87_wdt", Arch: ArchAMD64},
		{Name: "HW_RANDOM_AMD Support", Symbols: []string{"amd_rng_mod_init", "amd_rng_init", "amd_rng_read"}, Module: "amd-rng", Arch: ArchAMD64},
		{Name: "HW_RANDOM_INTEL Support", Symbols: []string{"intel_rng_mod_init", "intel_rng_init", "intel_rng_hw_init"}, Module: "intel-rng", Arch: ArchAMD64},

		// === ARM64-only ===
		{Name: "DRM V3D Support", Symbols: []string{"v3d_v71_ops", "v3d_driver"}, Path: "/sys/module/v3d", Module: "v3d", Arch: ArchARM64},
		{Name: "MMC SDHCI IPROC Support", Symbols: []string{"sdhci_iproc_probe", "sdhci_iproc_driver"}, Path: "/sys/module/sdhci_iproc", Arch: ArchARM64},
		{Name: "MMC BCM2835 Support", Symbols: []string{
			"bcm2835_probe", "bcm2835_driver_init", "bcm2835_mmc_probe",
			"bcm2835_mmc_driver", "mci_bcm2835_driver", "bcm2835_mmc_irq",
			"sdhci_bcm2835_probe", "sdhci_bcm2835_ops", "bcm2835_mmc_ops",
			"bcm2835_mmc_pdata", "bcm2835_sdhost_probe", "bcm2835_sdhost_driver",
		}, Path: "/sys/module/bcm2835_mmc", Module: "bcm2835_mmc", Arch: ArchARM64},
		{Name: "PINCTRL BCM2835 Support", Symbols: []string{"bcm2835_pinctrl_probe"}, Arch: ArchARM64},
		{Name: "PINCTRL ROCKCHIP Support", Symbols: []string{"rockchip_pinctrl_probe"}, Arch: ArchARM64},
		{Name: "BCM2835_WDT Support", Symbols: []string{"bcm2835_wdt_probe", "bcm2835_wdt_init"}, Module: "bcm2835_wdt", Arch: ArchARM64},
		{Name: "DW_WATCHDOG Support", Symbols: []string{"dw_wdt_probe", "dw_wdt_init", "dw_wdt_drv_probe"}, Module: "dw_wdt", Arch: ArchARM64},
		{Name: "HW_RANDOM_BCM2835 Support", Symbols: []string{"bcm2835_rng_probe", "bcm2835_rng_init"}, Module: "bcm2835-rng", Arch: ArchARM64},
		{Name: "HW_RANDOM_ROCKCHIP Support", Symbols: []string{"rk_rng_probe", "rk_rng_driver", "rk3568_rng_read", "rk3568_rng_init", "rk3576_rng_init"}, Module: "rockchip-rng", Arch: ArchARM64},
		{Name: "HW_RANDOM_ARM_SMCCC_TRNG Support", Symbols: []string{"smccc_trng_probe", "smccc_trng_init", "smccc_trng_driver"}, Module: "smccc_trng", Arch: ArchARM64},
		{Name: "HW_RANDOM_IPROC_RNG200 Support", Symbols: []string{"iproc_rng200_probe", "iproc_rng200_init", "iproc_rng200_driver"}, Module: "iproc-rng200", Arch: ArchARM64},
	}
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
