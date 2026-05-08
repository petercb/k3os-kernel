package main

import (
	"strings"
	"testing"
)

func TestParseWhence(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]bool
		wantLen int
	}{
		{
			name: "basic file entries",
			input: `Driver: foo
File: foo1.bin
File: foo2.bin
Licence: GPL
`,
			want: map[string]bool{
				"foo1.bin": true,
				"foo2.bin": true,
			},
			wantLen: 2,
		},
		{
			name: "link entries",
			input: `File: real.bin
Link: link1.bin -> real.bin
Link: link2.bin -> ../real.bin`,
			want: map[string]bool{
				"real.bin":  true,
				"link1.bin": true,
				"link2.bin": true,
			},
			wantLen: 3,
		},
		{
			name: "quoted paths",
			input: `File: "brcm/brcmfmac43455-sdio.MINIX-NEO Z83-4.txt"
Link: "brcm/spaced link.bin" -> "brcm/brcmfmac43455-sdio.MINIX-NEO Z83-4.txt"`,
			want: map[string]bool{
				"brcm/brcmfmac43455-sdio.MINIX-NEO Z83-4.txt": true,
				"brcm/spaced link.bin":                        true,
			},
			wantLen: 2,
		},
		{
			name: "ignored lines",
			input: `**********
* WHENCE *
**********
This file attempts to document...
Driver: BCM-0bb4-0306 - Cypress Bluetooth
File: brcm/BCM-0bb4-0306.hcd
Licence: Redistributable
Version: 1.0`,
			want: map[string]bool{
				"brcm/BCM-0bb4-0306.hcd": true,
			},
			wantLen: 1,
		},
		{
			name:    "empty file",
			input:   "",
			want:    map[string]bool{},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := ParseWhence(r)
			if err != nil {
				t.Fatalf("ParseWhence() error = %v", err)
			}

			if len(got) != tt.wantLen {
				t.Errorf("ParseWhence() returned %d entries, want %d", len(got), tt.wantLen)
			}

			for key := range tt.want {
				if !got[key] {
					t.Errorf("ParseWhence() missing key %q", key)
				}
			}
		})
	}
}
