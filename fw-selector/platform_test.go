package main

import (
	"testing"
)

func TestGetPlatformFirmware(t *testing.T) {
	tests := []struct {
		name string
		arch string
		want []string
	}{
		{
			name: "arm64 firmware",
			arch: "arm64",
			want: []string{
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
			},
		},
		{
			name: "amd64 firmware",
			arch: "amd64",
			want: []string{
				"i915/",
				"amdgpu/",
			},
		},
		{
			name: "unknown architecture",
			arch: "unknown",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPlatformFirmware(tt.arch)
			assertStringSlice(t, got, tt.want)
		})
	}
}
