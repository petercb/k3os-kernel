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
- **Status**: In Progress
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
