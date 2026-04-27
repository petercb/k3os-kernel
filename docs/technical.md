# Technical Documentation: K3OS Kernel Build & Test System

## 1. Core Principles
- **Base Source**: Ubuntu Linux Kernel.
- **Minimalism**: Disable all kernel features not required for Kubernetes or target hardware.
- **Reproducibility**: Use Dockerized environments for builds (WIP: transitioning from `localbuildenv.sh` to Dev Containers).
- **Automated Validation**: Every build must pass a suite of boot tests in QEMU.
- **Configuration as Code**: All kernel config changes are tracked in `overlay/debian.k3os/config/annotations`.

## 2. Engineering Patterns
- **Ubuntu-Style Annotations**: We use the `debian/rules` mechanism from Ubuntu/Debian kernels to manage configurations across multiple architectures.
- **u-root-init**: A Go-based init program that replaces standard init during testing. It performs direct syscalls and inspects `/proc` and `/sys` to verify kernel state.
- **Taskfile Automation**: We are moving away from monolithic bash scripts to modular `Taskfile.yml` tasks for clarity and dependency management.
- **Dynamic Module Validation (Planned)**: Investigation into using `pault.ag/go/modprobe` to load and validate drivers dynamically, allowing for smaller kernels with more modular components.

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
- `/src/[module]`: (Future) Proposed structure for core logic.
- `/u-root-init`: Go source for the boot validation init program.
- `/overlay`: Kernel configuration and Debian build files.
- `/tests`: (Proposed) Location for integration and unit tests.
- `localbuildenv.sh`: Script to enter the build container.
- `build.sh`: Main kernel build script (runs inside container).
- `test-kernel.sh`: Main test script (runs inside container).

## 5. Development Workflow
1. **Enter Environment**: Run `./localbuildenv.sh` to enter the Dockerized build environment.
2. **Lint/Format**: (Inside container) `task lint` and `task fmt`.
3. **Build Kernel**: (Inside container) `./build.sh`.
4. **Boot Test**: (Inside container) `./test-kernel.sh`.

## 6. Target Hardware Specifics
- **AMD64**: Generic commodity hardware support (NUCs, mini PCs). Includes HFS+ support for ISO boot validation (`u-root-init/main_amd64.go`).
- **ARM64**:
    - RPi 4/5: Uses `v3d`, `bcm2835_mmc`, `bcm2835_pinctrl`.
    - **Rockchip**: Uses `rockchip_pinctrl`.

## 7. Development Conventions

### Git Flow
- **Feature Branches**: `feature/[task-id]-[slug]`
- **Main Branch**: `7.0` (for current modernization effort)
- **Merging**: Fast-forward merges (`--ff-only`) are preferred into the main branch.

### Commit Messages
- **Convention**: [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
- **Format**: `<type>[optional scope]: <description>`
- **Types**: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`.
- **Example**: `fix(boot): [TASK-001] resolve failing BCM2835 MMC symbols`
