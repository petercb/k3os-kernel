# Project Status

## Completed Features
- **[TASK-001] Fix failing boot validation tests for Linux 7.0**: Refactored test suite and fixed BCM2835 symbols.
- **[TASK-005] Dynamic Module Loading**: Integrated `pault.ag/go/modprobe` for automated driver loading during validation.
- **[TASK-007] Optimize firmware selection tool**: Replaced fragile bash script with precise Go-based `fw-selector` tool, validating against upstream WHENCE manifest.

## In Progress
- **Onboarding and Documentation** (TASK-000)
    - ✅ PRD.md
    - ✅ technical.md
    - ✅ architecture.mermaid
    - ✅ unit_testing_guideline.md
    - ✅ tasks.md
    - ✅ Final Review
- **[TASK-002] Bash to Taskfile conversion**: Migrating build and test logic from `build.sh` and `test-kernel.sh` to a unified `Taskfile.yml`.

## Pending
- [TASK-XXX] Transition to Dev Containers extension (replacing `localbuildenv.sh`).
- [TASK-003] Add support for RPi 5
- [TASK-004] Add support for Rock 4/5

## Known Issues
- None (TASK-007 resolved build log noise).

## Decision History
- **2026-05-08** [TASK-007] — Adopted precise Go-based mapping of CONFIG symbols to source files to eliminate false-positive firmware inclusions.
- **2026-04-27** — Initialized PROJECT_ONBOARDING_MODE to establish engineering rigor for the Linux 7.0 update.

## Next Steps
- Begin [TASK-002] Bash to Taskfile migration.
