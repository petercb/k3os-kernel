package main

import (
	"strings"
	"testing"
)

func TestParseMakefile(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string][]string
		wantLen int
	}{
		{
			name:  "single obj-$(CONFIG_FOO)",
			input: "obj-$(CONFIG_FOO) += foo.o\n",
			want: map[string][]string{
				"FOO": {"foo.o"},
			},
			wantLen: 1,
		},
		{
			name:  "multiple objects on one line",
			input: "obj-$(CONFIG_NET) += net_core.o net_util.o\n",
			want: map[string][]string{
				"NET": {"net_core.o", "net_util.o"},
			},
			wantLen: 1,
		},
		{
			name: "multiple configs",
			input: `obj-$(CONFIG_FOO) += foo.o
obj-$(CONFIG_BAR) += bar.o baz.o`,
			want: map[string][]string{
				"FOO": {"foo.o"},
				"BAR": {"bar.o", "baz.o"},
			},
			wantLen: 2,
		},
		{
			name:  "tabs instead of spaces",
			input: "obj-$(CONFIG_FOO)\t+=\tfoo.o\n",
			want: map[string][]string{
				"FOO": {"foo.o"},
			},
			wantLen: 1,
		},
		{
			name: "obj-y unconditionally compiled",
			input: `obj-y += always.o
obj-$(CONFIG_FOO) += foo.o`,
			want: map[string][]string{
				"__BUILTIN__": {"always.o"},
				"FOO":         {"foo.o"},
			},
			wantLen: 2,
		},
		{
			name:  "obj-m unconditionally module",
			input: "obj-m += modonly.o\n",
			want: map[string][]string{
				"__MODULE__": {"modonly.o"},
			},
			wantLen: 1,
		},
		{
			name: "no CONFIG_ entries",
			input: `# SPDX-License-Identifier: GPL-2.0
# Makefile for something
ccflags-y += -DFOO`,
			want:    map[string][]string{},
			wantLen: 0,
		},
		{
			name:    "empty makefile",
			input:   "",
			want:    map[string][]string{},
			wantLen: 0,
		},
		{
			name: "continuation lines with backslash",
			input: `obj-$(CONFIG_DRM) += drm_core.o \
	drm_ioctl.o \
	drm_mm.o`,
			want: map[string][]string{
				"DRM": {"drm_core.o", "drm_ioctl.o", "drm_mm.o"},
			},
			wantLen: 1,
		},
		{
			name: "mixed continuation and non-continuation",
			input: `obj-$(CONFIG_A) += a1.o \
	a2.o
obj-$(CONFIG_B) += b1.o`,
			want: map[string][]string{
				"A": {"a1.o", "a2.o"},
				"B": {"b1.o"},
			},
			wantLen: 2,
		},
		{
			name:  "colon-equals syntax",
			input: "obj-$(CONFIG_FOO) := foo.o\n",
			want: map[string][]string{
				"FOO": {"foo.o"},
			},
			wantLen: 1,
		},
		{
			name: "subdir entries with trailing slash",
			input: `obj-$(CONFIG_FOO) += foo/
obj-$(CONFIG_BAR) += bar.o`,
			want: map[string][]string{
				"BAR": {"bar.o"},
			},
			wantLen: 1,
		},
		{
			name: "multi-part objects",
			input: `obj-$(CONFIG_AMDGPU) += amdgpu.o
amdgpu-y := amdgpu_drv.o amdgpu_device.o
obj-$(CONFIG_IWLWIFI) += iwlwifi.o
iwlwifi-objs += iwl-drv.o iwl-debug.o`,
			want: map[string][]string{
				"AMDGPU":  {"amdgpu_drv.o", "amdgpu_device.o"},
				"IWLWIFI": {"iwl-drv.o", "iwl-debug.o"},
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := ParseMakefile(r)
			if err != nil {
				t.Fatalf("ParseMakefile() error = %v", err)
			}

			if len(got) != tt.wantLen {
				t.Errorf("ParseMakefile() returned %d entries, want %d; got: %v", len(got), tt.wantLen, got)
			}

			assertMakefileResult(t, got, tt.want)
		})
	}
}

// assertMakefileResult compares the parsed Makefile output against expected values.
func assertMakefileResult(t *testing.T, got, want map[string][]string) {
	t.Helper()

	for key, wantObjs := range want {
		gotObjs, ok := got[key]
		if !ok {
			t.Errorf("ParseMakefile() missing key %q", key)
			continue
		}

		if len(gotObjs) != len(wantObjs) {
			t.Errorf("ParseMakefile()[%q] = %v, want %v", key, gotObjs, wantObjs)
			continue
		}

		for i, w := range wantObjs {
			if gotObjs[i] != w {
				t.Errorf("ParseMakefile()[%q][%d] = %q, want %q", key, i, gotObjs[i], w)
			}
		}
	}
}

func TestParseMakefileDuplicateConfig(t *testing.T) {
	input := `obj-$(CONFIG_FOO) += foo.o
obj-$(CONFIG_FOO) += foo_extra.o`

	r := strings.NewReader(input)
	got, err := ParseMakefile(r)
	if err != nil {
		t.Fatalf("ParseMakefile() error = %v", err)
	}

	objs := got["FOO"]
	if len(objs) != 2 {
		t.Fatalf("ParseMakefile() FOO has %d objects, want 2; got: %v", len(objs), objs)
	}

	if objs[0] != "foo.o" || objs[1] != "foo_extra.o" {
		t.Errorf("ParseMakefile() FOO = %v, want [foo.o foo_extra.o]", objs)
	}
}
