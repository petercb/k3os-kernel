# Tasks: K3OS Kernel Modernization

## Current Progress
- **Status**: Onboarding / Initializing Linux 7.0 Update.

---

## [TASK-001] Fix failing boot validation tests for Linux 7.0
- **Priority**: High
- **Status**: Planned
- **PRD Reference**: Success Criteria - All boot tests pass.
- **Implementation Checklist**:
    - [ ] Run `./test-kernel.sh` inside the container to identify specific failures.
    - [ ] Analyze `qemu.log` for symbol or driver mismatches in Linux 7.0.
    - [ ] Fix symbol names in `u-root-init/main.go` and `main_arm64.go` if they changed in 7.0.
    - [ ] Update `overlay/annotations` if specific modules failed to load.
    - [ ] Verify fix with `./test-kernel.sh` for both `amd64` and `arm64`.
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
