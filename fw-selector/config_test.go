package main

import (
	"strings"
	"testing"
)

func TestParseKernelConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
		wantLen  int
	}{
		{
			name: "basic enabled configs",
			input: `CONFIG_FOO=y
CONFIG_BAR=m
# CONFIG_BAZ is not set
CONFIG_QUX=y`,
			wantKeys: []string{"FOO", "BAR", "QUX"},
			wantLen:  3,
		},
		{
			name:     "empty file",
			input:    "",
			wantKeys: nil,
			wantLen:  0,
		},
		{
			name: "only comments and blank lines",
			input: `#
# Automatically generated file; DO NOT EDIT.
# Linux/arm64 7.0.0 Kernel Configuration
#
`,
			wantKeys: nil,
			wantLen:  0,
		},
		{
			name: "mixed content",
			input: `#
# Compiler: gcc (Ubuntu 14.2.0-16ubuntu1) 14.2.0
#
CONFIG_CC_VERSION_TEXT="gcc (Ubuntu 14.2.0-16ubuntu1) 14.2.0"
CONFIG_MODULES=y
CONFIG_MODULE_FORCE_LOAD=m
# CONFIG_MODULE_SRCVERSION_ALL is not set`,
			wantKeys: []string{"MODULES", "MODULE_FORCE_LOAD"},
			wantLen:  2,
		},
		{
			name: "string values are not included",
			input: `CONFIG_DEFAULT_HOSTNAME="(none)"
CONFIG_FOO=y`,
			wantKeys: []string{"FOO"},
			wantLen:  1,
		},
		{
			name: "numeric values are not included",
			input: `CONFIG_LOG_BUF_SHIFT=17
CONFIG_BAR=m`,
			wantKeys: []string{"BAR"},
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := ParseKernelConfig(r)
			if err != nil {
				t.Fatalf("ParseKernelConfig() error = %v", err)
			}

			if len(got) != tt.wantLen {
				t.Errorf("ParseKernelConfig() returned %d configs, want %d", len(got), tt.wantLen)
			}

			for _, key := range tt.wantKeys {
				if !got[key] {
					t.Errorf("ParseKernelConfig() missing expected key %q", key)
				}
			}
		})
	}
}

func TestParseKernelConfigDisabledNotIncluded(t *testing.T) {
	input := `CONFIG_FOO=y
# CONFIG_BAZ is not set
CONFIG_BAR=m`

	r := strings.NewReader(input)
	got, err := ParseKernelConfig(r)
	if err != nil {
		t.Fatalf("ParseKernelConfig() error = %v", err)
	}

	if got["BAZ"] {
		t.Error("ParseKernelConfig() should not include disabled config BAZ")
	}
}
