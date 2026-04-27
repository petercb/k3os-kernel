# PRD: K3OS Kernel Modernization (Linux 7.0)

## 1. Product Vision
Maintain a lean, stable, and high-performance Linux kernel specifically tailored for K3OS nodes. The focus is on providing only the necessary drivers and subsystems required for Kubernetes workloads, with a strong emphasis on automation and broad ARM64 hardware support.

## 2. Goals & Success Criteria
- **Modernization**: Successfully update the kernel base from Linux 6.8 to Linux 7.0.
- **Stability**: Ensure the kernel is rock-solid for long-running Kubernetes clusters.
- **Size Reduction**: Minimize the kernel image and module footprint by stripping unused features.
- **Hardware Support**: Maintain first-class support for AMD64 and ARM64 (RPi 4), and expand to RPi 5 and Rockchip-based boards (Rock 4/5).
- **Automation**: Fully migrate build and test workflows to `Taskfile.yml`.

## 3. User Personas/Stakeholders
- **K3OS Developers**: Need a reliable build system to iterate on kernel configs.
- **Cluster Operators**: Need a stable kernel that supports modern container features (Cgroup v2, OverlayFS, etc.).
- **Edge Hardware Users**: Need support for specific SBCs (RPi, Rockchip).

## 4. User Flow
1. **Configure**: Developer modifies `overlay/debian.k3os/config/annotations`.
2. **Build**: Developer runs `task build` (likely within a container).
3. **Validate**: Automated tests run `u-root-init` via QEMU for both AMD64 and ARM64.
4. **Deploy**: Artifacts (kernel, initrd, squashfs) are uploaded for K3OS consumption.

## 5. Features & Requirements
- **Kernel Config Management**: Use Ubuntu-style annotations for maintainability.
- **Boot Validation Suite**: Go-based init program (`u-root-init`) to check:
    - Filesystems (OverlayFS, SquashFS)
    - Container primitives (Cgroup v2, Namespaces)
    - Hardware drivers (USB, NVMe, GPU, MMC)
- **Multi-Arch Support**: Unified build system for `amd64` and `arm64`.
- **JUnit Reporting**: Automated test results compatible with CI pipelines.

## 6. Out of Scope
- Support for non-Kubernetes workloads.
- Legacy hardware drivers (32-bit, obscure NICs).
- Desktop-specific features (Audio, complex Desktop GPUs).

## 7. Constraints & Assumptions
- Built within a Debian-based Docker container.
- Uses `u-root` for minimal initrd generation.
- Assumes QEMU is available for CI validation.

## 8. Acceptance Criteria
- `task build` completes successfully for all target architectures.
- All tests in `u-root-init` pass in QEMU.
- Kernel image size does not regress significantly compared to 6.8.
- Validated boot on at least one physical ARM64 device (RPi 4).

## 9. Metrics/KPIs
- Build time (target < 20 mins in CI).
- Artifact size (SquashFS size).
- Test coverage (number of kernel subsystems validated).
