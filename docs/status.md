# Project Status

## Completed Features
- Initial modernization to Linux 7.0 (Build system update).
- Taskfile.yml base implementation.
- **[TASK-001] Fix failing boot validation tests for Linux 7.0**: Refactored test suite to modular Go structure, added unit tests, fixed IPTables/NFTables support, and identified Linux 7.0 specific hardware symbols for BCM2835 MMC. (2026-04-27)

## In Progress
- **Onboarding and Documentation** (TASK-000)
    - ✅ PRD.md
    - ✅ technical.md
    - ✅ architecture.mermaid
    - ✅ unit_testing_guideline.md
    - ✅ tasks.md
    - 🏗️ Final Review
- **[TASK-002] Bash to Taskfile conversion**: Migrating build and test logic from `build.sh` and `test-kernel.sh` to a unified `Taskfile.yml`.

## Pending
- [TASK-XXX] Transition to Dev Containers extension (replacing `localbuildenv.sh`).

## Known Issues
- Boot validation tests failing on Linux 7.0 update.

## Decision History
- **2026-04-27** — Initialized PROJECT_ONBOARDING_MODE to establish engineering rigor for the Linux 7.0 update.

## Next Steps
- Finalize onboarding and start [TASK-001].
