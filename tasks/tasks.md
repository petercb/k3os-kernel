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
- **Priority**: Medium
- **Status**: Planned
- **PRD Reference**: N/A (Internal optimization)
- **Dependencies**: [TASK-001]
- **Implementation Checklist**:
    - [ ] **Step 1: Scaffold `fw-selector` module**
        - [ ] Create `fw-selector/go.mod` with `module fw-selector` and `go 1.22.1`.
        - [ ] Create `fw-selector/main.go` with CLI flags: `--config`, `--source-dir`, `--whence`, `--arch`, `--output`.
        - [ ] Verify the module compiles with `go build ./fw-selector/...`.
    - [ ] **Step 2: Implement kernel config parser (TDD)**
        - [ ] Write `config_test.go` (CONFIG_FOO=y, CONFIG_BAR=m, comments, blank lines, empty file).
        - [ ] Run test → verify RED.
        - [ ] Implement `config.go`: `ParseKernelConfig(reader io.Reader) (map[string]bool, error)`.
        - [ ] Run test → verify GREEN.
    - [ ] **Step 3: Implement Makefile parser (TDD)**
        - [ ] Write `makefile_test.go` (obj-$(CONFIG_FOO), tabs/spaces, continuation lines, obj-y/obj-m, no CONFIG_ entries).
        - [ ] Run test → verify RED.
        - [ ] Implement `makefile.go`: `ParseMakefile(reader io.Reader) (map[string][]string, error)`.
        - [ ] Run test → verify GREEN.
    - [ ] **Step 4: Implement MODULE_FIRMWARE extractor (TDD)**
        - [ ] Write `firmware_test.go` (single/multiple MODULE_FIRMWARE, macro-expanded paths, no MODULE_FIRMWARE).
        - [ ] Run test → verify RED.
        - [ ] Implement `firmware.go`: `ExtractModuleFirmware(reader io.Reader) ([]string, error)`.
        - [ ] Run test → verify GREEN.
    - [ ] **Step 5: Implement WHENCE manifest parser (TDD)**
        - [ ] Write `whence_test.go` (File: entries, Link: entries, quoted paths, ignored lines, empty file).
        - [ ] Run test → verify RED.
        - [ ] Implement `whence.go`: `ParseWhence(reader io.Reader) (map[string]bool, error)`.
        - [ ] Run test → verify GREEN.
    - [ ] **Step 6: Implement platform firmware lists (TDD)**
        - [ ] Write `platform_test.go` (arm64 RPi/Rockchip, amd64 i915/amdgpu, unknown arch).
        - [ ] Run test → verify RED.
        - [ ] Implement `platform.go`: `GetPlatformFirmware(arch string) []string`.
        - [ ] Run test → verify GREEN.
    - [ ] **Step 7: Implement selector orchestrator (TDD)**
        - [ ] Write `selector_test.go` (enabled/disabled drivers, false-positive scenario, WHENCE validation, missing source files).
        - [ ] Run test → verify RED.
        - [ ] Implement `selector.go`: `Selector.SelectFirmware() ([]string, error)`.
        - [ ] Run test → verify GREEN.
    - [ ] **Step 8: Implement CLI entrypoint**
        - [ ] Wire up `main.go` (parse flags, load WHENCE, call Selector, merge platform, deduplicate, output).
        - [ ] Verify `go build ./fw-selector` produces a working binary.
    - [ ] **Step 9: Integration validation**
        - [ ] Create golden-file test with synthetic kernel source layout + testdata/WHENCE.
        - [ ] Run test → compare output to golden file.
        - [ ] Compare output to current `select_firmware.sh` output within Docker.
    - [ ] **Step 10: Update Dockerfile**
        - [ ] Fetch WHENCE file during compile stage (`ADD` from linux-firmware.git).
        - [ ] Add build stage to compile `fw-selector` from Go source.
        - [ ] Replace `select_firmware.sh` invocation with Go binary + `--whence /tmp/WHENCE`.
    - [ ] **Step 11: Final validation**
        - [ ] Run full Docker build for amd64 (`./local-build.sh`).
        - [ ] Run full Docker build for arm64 (`./local-build.sh` or CI).
        - [ ] Compare `firmware-list.txt` — confirm fewer false-positive entries.
        - [ ] Verify boot tests pass (QEMU).
        - [ ] Run `task lint` and `task fmt`.
- **Acceptance Criteria**:
    - `firmware-list.txt` contains only firmware for enabled drivers, avoiding false positives from same-directory disabled drivers.
    - Extracted firmware paths are validated against the upstream WHENCE manifest.
    - Docker build succeeds for both amd64 and arm64; boot validation passes.
    - Build logs are clean — no spurious firmware warnings.
- **Complexity**: Medium
