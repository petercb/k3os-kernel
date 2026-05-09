# Tasks: K3OS Kernel Modernization

## Current Progress
- **Status**: Onboarding / Initializing Linux 7.0 Update.

---

## [TASK-001] Fix failing boot validation tests for Linux 7.0
- **Status**: Done
- **Priority**: High
- **PRD Reference**: Section 5 (Boot Validation Suite)
- **Implementation Checklist**:
    - [x] Create unit tests for `u-root-init` with mocked filesystem/kallsyms.
    - [x] Refactor `u-root-init` for testability (extracted symbol and testing logic).
    - [x] Identify and fix broken feature tests for Linux 7.0 using unit tests.
    - [x] Run `CONFIGMODE=update ./build.sh` in the container to apply annotation changes.
    - [x] Run `./build.sh` to rebuild the kernel with new configs.
    - [x] Run `./test-kernel.sh` inside the container to verify fixes.
    - [x] Analyze `qemu.log` for symbol or driver mismatches in Linux 7.0.
    - [x] Fix symbol names in `u-root-init/main.go` and `main_arm64.go` if they changed in 7.0.
    - [x] Update `overlay/annotations` if specific modules failed to load.
    - [x] Verify fix with `./test-kernel.sh` for both `amd64` and `arm64`.
- **Acceptance Criteria**: All tests in `u-root-init` report `[PASS]`.
- **Complexity**: Medium

## [TASK-002] Complete transition from Bash to Taskfile
- **Priority**: Medium
- **Status**: Cancelled
- **PRD Reference**: Automation - Fully migrate to Taskfile.yml.
- **Implementation Checklist**:
    - [ ] Audit remaining bash scripts (`build.sh`, `test-kernel.sh`).
    - [ ] Port `build.sh` logic to `Taskfile.yml`.
    - [ ] Port `test-kernel.sh` logic to `Taskfile.yml`.
    - [ ] Remove legacy scripts after verification.
- **Acceptance Criteria**: All build and test operations can be performed via `task`.
- **Complexity**: Medium

## [TASK-003] Add support for RPi 5
- **Priority**: Medium
- **Status**: Planned
- **Dependencies**: [TASK-001]
- **Implementation Checklist**:
    - [ ] Research required RPi 5 kernel configs for Linux 7.0.
    - [ ] Add RPi 5 specific tests to `u-root-init`.
    - [ ] Update `overlay/annotations` for RPi 5 hardware support.
- **Acceptance Criteria**: Kernel boots and validates on RPi 5.
- **Complexity**: High

## [TASK-004] Add support for Rock 4/5
- **Priority**: Low
- **Status**: Planned
- **Dependencies**: [TASK-001]
- **Implementation Checklist**:
    - [ ] Research Rockchip kernel requirements.
    - [ ] Add Rockchip specific tests to `u-root-init`.
    - [ ] Update `overlay/annotations` for Rock 4/5.
- **Acceptance Criteria**: Kernel boots and validates on Rockchip hardware.
- **Complexity**: High

## [TASK-005] Investigate `pault.ag/go/modprobe` integration
- **Priority**: Low
- **Status**: Done
- **PRD Reference**: Section 5.3 (Module Loading Strategy)
- **Dependencies**: [TASK-001]
- **Implementation Checklist**:
    - [x] Evaluate `pault.ag/go/modprobe` for loading modules in `u-root-init`.
    - [x] Determine if `initrd` needs to include `.ko` files for validation.
    - [x] Test loading a driver (e.g., `v3d`) as a module and validating via symbol/sysfs.
- **Acceptance Criteria**: Proof of concept showing module loading and validation.
- **Complexity**: Medium

## [TASK-006] Implement basic kernel stress testing in u-root-init
- **Priority**: Medium
- **Status**: Done
- **Dependencies**:
- **Implementation Checklist**:
    - [x] Write unit tests for stress testing logic in `u-root-init/stress_test.go` (TDD).
    - [x] Implement CPU and Memory stress functions in `u-root-init/stress.go`.
    - [x] Integrate the stress test into the test runner in `u-root-init/main_linux.go`.
    - [x] Verify the stress tests run successfully in QEMU via `./test-kernel.sh`.
- **Acceptance Criteria**: Kernel gracefully handles a short burst of CPU and memory stress during boot validation.
- **Complexity**: Low

## [TASK-007] Optimize firmware selection tool
- **Status**: Done (2026-05-09)
- **Priority**: Medium
- **PRD Reference**: N/A (Internal optimization)
- **Dependencies**: [TASK-001]
- **Implementation Checklist**:
    - [x] **Step 1: Scaffold `fw-selector` module**
        - [x] Create `fw-selector/go.mod` with `module fw-selector` and `go 1.22.1`.
        - [x] Create `fw-selector/main.go` with CLI flags: `--config`, `--source-dir`, `--whence`, `--arch`, `--output`.
        - [x] Verify the module compiles with `go build ./fw-selector/...`.
    - [x] **Step 2: Implement kernel config parser (TDD)**
        - [x] Write `config_test.go` (y/m values, comments, blank lines, empty, string/numeric values).
        - [x] Run test → verify RED.
        - [x] Implement `config.go`: `ParseKernelConfig(reader io.Reader) (map[string]bool, error)`.
        - [x] Run test → verify GREEN.
    - [x] **Step 3: Implement Makefile parser (TDD)**
        - [x] Write `makefile_test.go` (single/multi obj, tabs, continuation lines, obj-y, obj-m, subdir filtering).
        - [x] Run test → verify RED.
        - [x] Implement `makefile.go`: `ParseMakefile(reader io.Reader) (map[string][]string, error)`.
        - [x] Run test → verify GREEN.
    - [x] **Step 4: Implement MODULE_FIRMWARE extractor (TDD)**
        - [x] Write `firmware_test.go` (single/multiple, mixed code, macro-expanded paths, spacing).
        - [x] Run test → verify RED.
        - [x] Implement `firmware.go`: `ExtractModuleFirmware(reader io.Reader) ([]string, error)`.
        - [x] Run test → verify GREEN.
    - [x] **Step 5: Implement WHENCE manifest parser (TDD)**
        - [x] Write `whence_test.go` (File: entries, Link: entries, quoted paths, ignored lines, empty file).
        - [x] Run test → verify RED.
        - [x] Implement `whence.go`: `ParseWhence(reader io.Reader) (map[string]bool, error)`.
        - [x] Run test → verify GREEN.
    - [x] **Step 6: Implement platform firmware lists (TDD)**
        - [x] Write `platform_test.go` (arm64 RPi/Rockchip, amd64 i915/amdgpu, unknown arch).
        - [x] Run test → verify RED.
        - [x] Implement `platform.go`: `GetPlatformFirmware(arch string) []string`.
        - [x] Run test → verify GREEN.
    - [x] **Step 7: Implement selector orchestrator (TDD)**
        - [x] Write `selector_test.go` (enabled/disabled drivers, false-positive scenario, WHENCE validation, missing source files).
        - [x] Run test → verify RED.
        - [x] Implement `selector.go`: `Selector.SelectFirmware() ([]string, error)`.
        - [x] Run test → verify GREEN.
    - [x] **Step 8: Implement CLI entrypoint**
        - [x] Wire up `main.go` (parse flags, load WHENCE, call Selector, merge platform, deduplicate, output).
        - [x] Verify `go build ./fw-selector` produces a working binary.
    - [x] **Step 9: Integration validation**
        - [x] Create golden-file test with synthetic kernel source layout + testdata/WHENCE.
        - [x] Run test → compare output to golden file.
        - [x] Compare output to current `select_firmware.sh` output within Docker.
    - [x] **Step 10: Update Dockerfile**
        - [x] Fetch WHENCE file during compile stage (`ADD` from linux-firmware.git).
        - [x] Add build stage to compile `fw-selector` from Go source.
        - [x] Replace `select_firmware.sh` invocation with Go binary + `--whence /tmp/WHENCE`.
    - [x] **Step 11: Final validation**
        - [x] Run full Docker build for amd64 (`./local-build.sh`).
        - [x] Run full Docker build for arm64 (`./local-build.sh` or CI).
        - [x] Compare `firmware-list.txt` — confirm fewer false-positive entries.
        - [x] Verify boot tests pass (QEMU).
        - [x] Run `task lint` and `task fmt`.
- **Acceptance Criteria**:
    - `firmware-list.txt` contains only firmware for enabled drivers, avoiding false positives from same-directory disabled drivers.
    - Extracted firmware paths are validated against the upstream WHENCE manifest.
    - Docker build succeeds for both amd64 and arm64; boot validation passes.
    - Build logs are clean — no spurious firmware warnings.
- **Complexity**: Medium
