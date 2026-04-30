# Activity Log
+
+## [2026-04-27 19:00] [TASK-005] Implemented dynamic module loading PoC.
+- **Action**: Integrated `pault.ag/go/modprobe`.
+- **Details**:
+    - Added `Module` field to `FeatureTest`.
+    - Implemented `tryLoadModule` with re-scanning of symbols.
+    - Updated test loop to attempt failover modprobe.
+    - Formatted code with `gofumpt`.

 ## [2026-04-27 18:25] [TASK-001] Completed boot validation fix. All 21 tests pass in QEMU (arm64).

## [2026-04-27 18:25] [TASK-001] Completed boot validation fix. All 21 tests pass in QEMU (arm64).
- **Action**: Completed boot validation fix.
- **Details**:
    - Refactored `u-root-init` into modular components.
    - Fixed IPTables/NFTables by enabling missing kernel tables.
    - Discovered and fixed Linux 7.0 symbols for BCM2835 MMC.
    - Updated documentation with `CONFIGMODE=update` workflow.

### Retrospective (TASK-001)
- **What went well**: Modular refactoring made debugging hardware-specific symbol failures much easier. Debug logging for symbols was crucial for identifying the `bcm2835_probe` rename.
- **What broke**: Initial `annotations` were too stripped down for K3s (missing IPTables filter/nat/mangle tables).
- **Next time**: Check for `NFTables` compatibility earlier when modernizing kernels.

## [2026-04-27 10:50] - Initial Onboarding
- **Action**: Created core documentation structure (`docs/`, `tasks/`).
- **Details**: Established PRD, Technical Docs, Architecture diagram, and initial task list for the Linux 7.0 modernization project.
- **Next**: Finalize onboarding and begin fixing boot tests.

## Retrospective
- Initial session setup complete. The project has a strong foundation with `u-root-init` but needs better documentation to manage the complexity of multi-arch kernel builds.
