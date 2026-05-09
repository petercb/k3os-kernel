package main

// GetPlatformFirmware returns a list of statically required firmware paths
// or directories for a given architecture. These are typically needed for
// boot or basic functionality regardless of which drivers are built as modules.
func GetPlatformFirmware(arch string) []string {
	if arch == "arm64" {
		return []string{
			"brcm/brcmfmac43455-sdio.bin",
			"brcm/brcmfmac43455-sdio.txt",
			"brcm/brcmfmac43455-sdio.clm_blob",
			"brcm/brcmfmac43456-sdio.bin",
			"brcm/brcmfmac43456-sdio.txt",
			"brcm/brcmfmac43456-sdio.clm_blob",
			"cypress/cyfmac43455-sdio.bin",
			"cypress/cyfmac43455-sdio.clm_blob",
			"raspberrypi/bootloader-2711/latest/pieeprom-2711-latest.bin",
			"raspberrypi/bootloader-2712/latest/pieeprom-2712-latest.bin",
			"rockchip/dptx.bin",
		}
	}

	if arch == "amd64" {
		// Firmware for GPUs like i915 and amdgpu are now automatically
		// resolved by parsing multi-part Makefile objects.
		return []string{}
	}

	return nil
}
