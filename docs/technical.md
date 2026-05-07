# Technical Documentation: K3OS Kernel Build & Test System

## 1. Core Principles
- **Base Source**: Ubuntu Linux Kernel.
- **Minimalism**: Disable all kernel features not required for Kubernetes or target hardware.
- **Reproducibility**: Use Dockerized environments for builds
- **Automated Validation**: Every build must pass a suite of boot tests in QEMU.
- **Configuration as Code**: All kernel config changes are tracked in `overlay/debian.k3os/config/annotations`.

## 2. Engineering Patterns
- **Ubuntu-Style Annotations**: We use the `debian/rules` mechanism from Ubuntu/Debian kernels to manage configurations across multiple architectures.
- **u-root-init**: A Go-based init program that replaces standard init during testing. It performs direct syscalls and inspects `/proc` and `/sys` to verify kernel state.

## 3. Technology Stack
- **Language**: Go (for `u-root-init`), Bash (primary scripts), YAML (Taskfile - WIP).
- **Kernel Base**: Linux 7.0 (modernized from 6.8).
- **Tools**:
    - `u-root`: For creating minimal initramfs.
    - `qemu-system-*`: For multi-arch boot validation.
    - `golangci-lint`: For Go code quality.
    - `gofumpt`: For strict Go formatting.
    - `docker`: For build environment isolation.

## 4. Directory Structure
- `/docs`: Documentation (PRD, Technical, Status).
- `/tasks`: Implementation plans and task summaries.
- `/u-root-init`: Go source for the boot validation init program.
- `/overlay`: Kernel configuration and Debian build files.
- `kernel-config.sh`: Script to enter the kernel configuration container.
- `local-build.sh`: Main kernel build script (for local testing).

## 5. Development Workflow
1. **Configure Kernel**: `./kernel-config.sh edit`
2. **Updatwe Kernel**: (after an upstream update) `./kernel-config.sh update`
1. **Build Kernel**: `./local-build.sh`.

## 6. Target Hardware Specifics
- **AMD64**: Generic commodity hardware support (NUCs, mini PCs). Includes HFS+ support for ISO boot validation.
- **ARM64**:
    - RPi 4/5: Uses `v3d`, `bcm2835_mmc`, `bcm2835_pinctrl`.
    - **Rockchip**: Uses `rockchip_pinctrl`.

## 7. Development Conventions

### Git Flow
- **Feature Branches**: `feature/[task-id]-[slug]`
- **Main Branch**: `master` (for current modernization effort)
- **Merging**: Fast-forward merges (`--ff-only`) are preferred into the main branch.

### Commit Messages
- **Convention**: [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
- **Format**: `<type>[optional scope]: <description>`
- **Types**: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`.
- **Example**: `fix(boot): [TASK-001] resolve failing BCM2835 MMC symbols`
