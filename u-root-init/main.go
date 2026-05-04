package main

import (
	"bufio"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
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

// FeatureTest describes a kernel feature to validate at boot via CONFIG.
type FeatureTest struct {
	Name     string
	Config   string
	Allowed  []string
	Arch     string
	Disabled bool
}

var (
	results  []TestResult
	Features = []FeatureTest{
		// === All architectures ===
		{Name: "Loop Device Support", Config: "CONFIG_BLK_DEV_LOOP", Allowed: []string{"y", "m"}},
		{Name: "HW Random Support", Config: "CONFIG_HW_RANDOM", Allowed: []string{"y", "m"}},
		{Name: "TPM Support", Config: "CONFIG_TCG_TPM", Allowed: []string{"y", "m"}},
		{Name: "Veth Support", Config: "CONFIG_VETH", Allowed: []string{"y", "m"}},
		{Name: "Bridge Support", Config: "CONFIG_BRIDGE", Allowed: []string{"y", "m"}},
		{Name: "IP Advanced Router Support", Config: "CONFIG_IP_ADVANCED_ROUTER", Allowed: []string{"y", "m"}},
		{Name: "USB UAS Support", Config: "CONFIG_USB_UAS", Allowed: []string{"y", "m"}},
		{Name: "VXLAN Support", Config: "CONFIG_VXLAN", Allowed: []string{"y", "m"}},
		{Name: "Netfilter Support", Config: "CONFIG_NETFILTER", Allowed: []string{"y", "m"}},
		{Name: "IPTables/NFTables Support", Config: "CONFIG_NF_TABLES", Allowed: []string{"y", "m"}},
		{Name: "Netfilter Masquerade Support", Config: "CONFIG_NETFILTER_XT_TARGET_MASQUERADE", Allowed: []string{"y", "m"}},
		{Name: "XT Match Comment Support", Config: "CONFIG_NETFILTER_XT_MATCH_COMMENT", Allowed: []string{"y", "m"}},
		{Name: "NVMe over TCP Support", Config: "CONFIG_NVME_TCP", Allowed: []string{"y", "m"}},
		{Name: "VFIO PCI Support", Config: "CONFIG_VFIO_PCI", Allowed: []string{"y", "m"}},
		{Name: "UIO PCI Generic Support", Config: "CONFIG_UIO_PCI_GENERIC", Allowed: []string{"y", "m"}},
		{Name: "VLAN 802.1Q Support", Config: "CONFIG_VLAN_8021Q", Allowed: []string{"y"}},
		{Name: "OverlayFS Support", Config: "CONFIG_OVERLAY_FS", Allowed: []string{"y"}},
		{Name: "In-Kernel Config Support", Config: "CONFIG_IKCONFIG", Allowed: []string{"y"}},
		{Name: "Config via /proc Support", Config: "CONFIG_IKCONFIG_PROC", Allowed: []string{"y"}},
		{Name: "Namespaces Support", Config: "CONFIG_NAMESPACES", Allowed: []string{"y"}},
		{Name: "Network Namespaces Support", Config: "CONFIG_NET_NS", Allowed: []string{"y"}},
		{Name: "PID Namespaces Support", Config: "CONFIG_PID_NS", Allowed: []string{"y"}},
		{Name: "IPC Namespaces Support", Config: "CONFIG_IPC_NS", Allowed: []string{"y"}},
		{Name: "UTS Namespaces Support", Config: "CONFIG_UTS_NS", Allowed: []string{"y"}},
		{Name: "Control Groups Support", Config: "CONFIG_CGROUPS", Allowed: []string{"y"}},
		{Name: "PIDs Cgroup Support", Config: "CONFIG_CGROUP_PIDS", Allowed: []string{"y"}},
		{Name: "CPU Accounting Cgroup Support", Config: "CONFIG_CGROUP_CPUACCT", Allowed: []string{"y"}},
		{Name: "Device Cgroup Support", Config: "CONFIG_CGROUP_DEVICE", Allowed: []string{"y"}},
		{Name: "Freezer Cgroup Support", Config: "CONFIG_CGROUP_FREEZER", Allowed: []string{"y"}},
		{Name: "CPU Scheduler Cgroup Support", Config: "CONFIG_CGROUP_SCHED", Allowed: []string{"y"}},
		{Name: "CPU Sets Support", Config: "CONFIG_CPUSETS", Allowed: []string{"y"}},
		{Name: "Memory Cgroup Support", Config: "CONFIG_MEMCG", Allowed: []string{"y"}},
		{Name: "Seccomp Support", Config: "CONFIG_SECCOMP", Allowed: []string{"y"}},
		{Name: "Key Management Support", Config: "CONFIG_KEYS", Allowed: []string{"y"}},
		{Name: "Bridge Netfilter Support", Config: "CONFIG_BRIDGE_NETFILTER", Allowed: []string{"y"}},
		{Name: "IP_NF Target REJECT Support", Config: "CONFIG_IP_NF_TARGET_REJECT", Allowed: []string{"y"}},
		{Name: "XT Match Addrtype Support", Config: "CONFIG_NETFILTER_XT_MATCH_ADDRTYPE", Allowed: []string{"y"}},
		{Name: "XT Match Conntrack Support", Config: "CONFIG_NETFILTER_XT_MATCH_CONNTRACK", Allowed: []string{"y"}},
		{Name: "XT Match IPVS Support", Config: "CONFIG_NETFILTER_XT_MATCH_IPVS", Allowed: []string{"y"}},
		{Name: "XT Match Comment Support (Strict)", Config: "CONFIG_NETFILTER_XT_MATCH_COMMENT", Allowed: []string{"y"}},
		{Name: "XT Match Multiport Support", Config: "CONFIG_NETFILTER_XT_MATCH_MULTIPORT", Allowed: []string{"y"}},
		{Name: "XT Match Statistic Support", Config: "CONFIG_NETFILTER_XT_MATCH_STATISTIC", Allowed: []string{"y"}},
		{Name: "IP_NF NAT Support", Config: "CONFIG_IP_NF_NAT", Allowed: []string{"y", "m"}},
		{Name: "NF NAT Support", Config: "CONFIG_NF_NAT", Allowed: []string{"y"}},
		{Name: "POSIX Mqueue Support", Config: "CONFIG_POSIX_MQUEUE", Allowed: []string{"y"}},
		{Name: "User Namespaces Support", Config: "CONFIG_USER_NS", Allowed: []string{"y"}},
		{Name: "Net Prio Cgroup Support", Config: "CONFIG_CGROUP_NET_PRIO", Allowed: []string{"y"}},
		{Name: "Block Cgroup Support", Config: "CONFIG_BLK_CGROUP", Allowed: []string{"y"}},
		{Name: "Block Device Throttling Support", Config: "CONFIG_BLK_DEV_THROTTLING", Allowed: []string{"y"}},
		{Name: "Perf Cgroup Support", Config: "CONFIG_CGROUP_PERF", Allowed: []string{"y"}},
		{Name: "HugeTLB Cgroup Support", Config: "CONFIG_CGROUP_HUGETLB", Allowed: []string{"y"}},
		{Name: "Net Cls Cgroup Support", Config: "CONFIG_NET_CLS_CGROUP", Allowed: []string{"y"}},
		{Name: "CFS Bandwidth Support", Config: "CONFIG_CFS_BANDWIDTH", Allowed: []string{"y"}},
		{Name: "Fair Group Scheduler Support", Config: "CONFIG_FAIR_GROUP_SCHED", Allowed: []string{"y"}},
		{Name: "RT Group Scheduler Support", Config: "CONFIG_RT_GROUP_SCHED", Allowed: []string{"y"}},
		{Name: "IP_NF Target REDIRECT Support", Config: "CONFIG_IP_NF_TARGET_REDIRECT", Allowed: []string{"y", "m"}},
		{Name: "IP Set Support", Config: "CONFIG_IP_SET", Allowed: []string{"y"}},
		{Name: "IP Virtual Server Support", Config: "CONFIG_IP_VS", Allowed: []string{"y"}},
		{Name: "IPVS NFCT Support", Config: "CONFIG_IP_VS_NFCT", Allowed: []string{"y"}},
		{Name: "IPVS TCP Protocol Support", Config: "CONFIG_IP_VS_PROTO_TCP", Allowed: []string{"y"}},
		{Name: "IPVS UDP Protocol Support", Config: "CONFIG_IP_VS_PROTO_UDP", Allowed: []string{"y"}},
		{Name: "IPVS Round-Robin Support", Config: "CONFIG_IP_VS_RR", Allowed: []string{"y"}},
		{Name: "EXT4 File System Support", Config: "CONFIG_EXT4_FS", Allowed: []string{"y"}},
		{Name: "EXT4 POSIX ACL Support", Config: "CONFIG_EXT4_FS_POSIX_ACL", Allowed: []string{"y"}},
		{Name: "EXT4 Security Labels Support", Config: "CONFIG_EXT4_FS_SECURITY", Allowed: []string{"y"}},
		{Name: "EXT4 for EXT2 Support", Config: "CONFIG_EXT4_USE_FOR_EXT2", Allowed: []string{"y"}},
		{Name: "IPVLAN Driver Support", Config: "CONFIG_IPVLAN", Allowed: []string{"y"}},
		{Name: "MACVLAN Driver Support", Config: "CONFIG_MACVLAN", Allowed: []string{"y"}},
		{Name: "Dummy Network Driver Support", Config: "CONFIG_DUMMY", Allowed: []string{"y", "m"}},
		{Name: "Netfilter NAT FTP Support", Config: "CONFIG_NF_NAT_FTP", Allowed: []string{"y"}},
		{Name: "Netfilter Conntrack FTP Support", Config: "CONFIG_NF_CONNTRACK_FTP", Allowed: []string{"y"}},
		{Name: "Netfilter NAT TFTP Support", Config: "CONFIG_NF_NAT_TFTP", Allowed: []string{"y"}},
		{Name: "Netfilter Conntrack TFTP Support", Config: "CONFIG_NF_CONNTRACK_TFTP", Allowed: []string{"y"}},
		{Name: "Cryptographic API Support", Config: "CONFIG_CRYPTO", Allowed: []string{"y"}},
		{Name: "Authenticated Encryption Support", Config: "CONFIG_CRYPTO_AEAD", Allowed: []string{"y"}},
		{Name: "GCM Support", Config: "CONFIG_CRYPTO_GCM", Allowed: []string{"y"}},
		{Name: "Sequence Number IV Support", Config: "CONFIG_CRYPTO_SEQIV", Allowed: []string{"y"}},
		{Name: "GHASH Support", Config: "CONFIG_CRYPTO_GHASH", Allowed: []string{"y"}},
		{Name: "Transformation (XFRM) Support", Config: "CONFIG_XFRM", Allowed: []string{"y"}},
		{Name: "XFRM User Interface Support", Config: "CONFIG_XFRM_USER", Allowed: []string{"y"}},
		{Name: "XFRM Algorithm Support", Config: "CONFIG_XFRM_ALGO", Allowed: []string{"y"}},
		{Name: "IPsec ESP Support", Config: "CONFIG_INET_ESP", Allowed: []string{"y"}},

		// === AMD64-only ===
		{Name: "HFS+ Support", Config: "CONFIG_HFSPLUS_FS", Allowed: []string{"y", "m"}, Arch: ArchAMD64},
		{Name: "NVRAM Support", Config: "CONFIG_NVRAM", Allowed: []string{"y", "m"}, Arch: ArchAMD64},
		{Name: "ITCO_WDT Support", Config: "CONFIG_ITCO_WDT", Allowed: []string{"y", "m"}, Arch: ArchAMD64},
		{Name: "IT87_WDT Support", Config: "CONFIG_IT87_WDT", Allowed: []string{"y", "m"}, Arch: ArchAMD64},

		// === ARM64-only ===
		{Name: "DRM V3D Support", Config: "CONFIG_DRM_V3D", Allowed: []string{"y", "m"}, Arch: ArchARM64},
		{Name: "MMC SDHCI IPROC Support", Config: "CONFIG_MMC_SDHCI_IPROC", Allowed: []string{"y", "m"}, Arch: ArchARM64},
		{Name: "MMC BCM2835 Support", Config: "CONFIG_MMC_BCM2835", Allowed: []string{"y", "m"}, Arch: ArchARM64},
		{Name: "PINCTRL BCM2835 Support", Config: "CONFIG_PINCTRL_BCM2835", Allowed: []string{"y", "m"}, Arch: ArchARM64},
		{Name: "PINCTRL ROCKCHIP Support", Config: "CONFIG_PINCTRL_ROCKCHIP", Allowed: []string{"y", "m"}, Arch: ArchARM64},
		{Name: "BCM2835_WDT Support", Config: "CONFIG_BCM2835_WDT", Allowed: []string{"y", "m"}, Arch: ArchARM64},
		{Name: "DW_WATCHDOG Support", Config: "CONFIG_DW_WATCHDOG", Allowed: []string{"y", "m"}, Arch: ArchARM64},
		{Name: "HW_RANDOM_BCM2835 Support", Config: "CONFIG_HW_RANDOM_BCM2835", Allowed: []string{"y", "m"}, Arch: ArchARM64},
		{Name: "HW_RANDOM_ROCKCHIP Support", Config: "CONFIG_HW_RANDOM_ROCKCHIP", Allowed: []string{"y", "m"}, Arch: ArchARM64},
		{Name: "HW_RANDOM_ARM_SMCCC_TRNG Support", Config: "CONFIG_HW_RANDOM_ARM_SMCCC_TRNG", Allowed: []string{"y", "m"}, Arch: ArchARM64},
		{Name: "HW_RANDOM_IPROC_RNG200 Support", Config: "CONFIG_HW_RANDOM_IPROC_RNG200", Allowed: []string{"y", "m"}, Arch: ArchARM64},
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

// LoadKernelConfigs reads /proc/config.gz and extracts the CONFIG_ map.
func LoadKernelConfigs() (map[string]string, error) {
	configs := make(map[string]string)
	f, err := os.Open("/proc/config.gz")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/config.gz: %w (is IKCONFIG_PROC enabled?)", err)
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()

	scanner := bufio.NewScanner(gz)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "CONFIG_") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				configs[parts[0]] = parts[1]
			}
		}
	}
	return configs, nil
}

func hasFilesystem(r io.Reader, name string) bool {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), name) {
			return true
		}
	}
	return false
}
