package main

// GetPlatformFirmware returns a list of statically required firmware paths
// or directories for a given architecture. These are typically needed for
// boot or basic functionality regardless of which drivers are built as modules.
func GetPlatformFirmware(arch string) []string {
	if arch == "arm64" {
		return []string{
			"rockchip/dptx.bin", // needed for early boot video
		}
	}

	if arch == "amd64" {
		return []string{}
	}

	return nil
}
