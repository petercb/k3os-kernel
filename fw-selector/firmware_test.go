package main

import (
	"strings"
	"testing"
)

func TestExtractModuleFirmware(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single MODULE_FIRMWARE",
			input: `MODULE_FIRMWARE("brcm/brcmfmac43455-sdio.bin");`,
			want:  []string{"brcm/brcmfmac43455-sdio.bin"},
		},
		{
			name: "multiple MODULE_FIRMWARE in one file",
			input: `MODULE_FIRMWARE("fw/driver_a.bin");
MODULE_FIRMWARE("fw/driver_a_data.bin");
MODULE_FIRMWARE("fw/driver_a.conf");`,
			want: []string{
				"fw/driver_a.bin",
				"fw/driver_a_data.bin",
				"fw/driver_a.conf",
			},
		},
		{
			name: "MODULE_FIRMWARE among other code",
			input: `#include <linux/module.h>
static int foo_init(void) { return 0; }
MODULE_FIRMWARE("foo/firmware.bin");
MODULE_LICENSE("GPL");
module_init(foo_init);`,
			want: []string{"foo/firmware.bin"},
		},
		{
			name: "no MODULE_FIRMWARE",
			input: `#include <linux/module.h>
static int bar_init(void) { return 0; }
MODULE_LICENSE("GPL");`,
			want: nil,
		},
		{
			name:  "empty file",
			input: "",
			want:  nil,
		},
		{
			name:  "macro-expanded path is skipped",
			input: `MODULE_FIRMWARE(FOO_FW_FILE ".bin");`,
			want:  nil,
		},
		{
			name: "mixed literal and macro paths",
			input: `MODULE_FIRMWARE("valid/path.bin");
MODULE_FIRMWARE(MACRO_PATH);
MODULE_FIRMWARE("another/valid.fw");`,
			want: []string{"valid/path.bin", "another/valid.fw"},
		},
		{
			name:  "firmware path with define concatenation",
			input: `MODULE_FIRMWARE(QAT_FW_DIR FW_DH895XCC);`,
			want:  nil,
		},
		{
			name:  "firmware with spaces before paren",
			input: `MODULE_FIRMWARE ("spaced/firmware.bin");`,
			want:  []string{"spaced/firmware.bin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := ExtractModuleFirmware(r)
			if err != nil {
				t.Fatalf("ExtractModuleFirmware() error = %v", err)
			}

			assertStringSlice(t, got, tt.want)
		})
	}
}

// assertStringSlice compares two string slices for equality.
func assertStringSlice(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("got %d items %v, want %d items %v", len(got), got, len(want), want)
		return
	}

	for i, w := range want {
		if got[i] != w {
			t.Errorf("item[%d] = %q, want %q", i, got[i], w)
		}
	}
}
