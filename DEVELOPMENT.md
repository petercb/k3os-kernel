# Development Guide: K3OS Kernel

This repository contains the Linux kernel source and configuration for **k3OS**, a container-optimized OS for **k3s**.

## 🎯 Project Overview
- **Base Source**: Ubuntu Linux Kernel.
- **Targets**:
  - **AMD64**: Generic commodity hardware (NUCs, mini PCs).
  - **ARM64**: Raspberry Pi 4/5, Rockchip boards.
- **Configuration**: Managed via Ubuntu-style annotations in `overlay/debian.k3os/config/annotations`.

## 🧪 Testing Methodology
The kernel is validated by booting it in **QEMU** with a custom `initrd` built using **u-root**.

### Running Tests
Inside the build container (or matching environment):
```bash
./test-kernel.sh
```
This script:
1. Builds a custom `init` program from `u-root-init/`.
2. Boots the kernel in QEMU.
3. Verifies essential features (Cgroups, Namespaces, OverlayFS, Hardware Drivers).
4. Reports results in **JUnit XML** format.

## 🧹 Linting and Formatting
We maintain high code quality for the Go-based testing tools.

### Formatting
Always use **gofumpt** for stricter formatting:
```bash
gofumpt -w u-root-init
```

### Linting
Use **golangci-lint** with the appropriate target environment flags:
```bash
cd u-root-init
GOOS=linux GOARCH=arm64 golangci-lint run
```

### Build Verification
To ensure the initrd code builds for the target environment:
```bash
cd u-root-init
GOOS=linux GOARCH=arm64 go build -o /dev/null .
GOOS=linux GOARCH=amd64 go build -o /dev/null .
```

## 📁 Repository Structure
- `overlay/`: Kernel configuration overlays and patches.
- `u-root-init/`: Go source for the kernel validation suite.
  - `main.go`: Generic kernel feature tests.
  - `main_amd64.go`: x86_64 specific tests (e.g., HFS+ for ISO boot).
  - `main_arm64.go`: ARM64 specific tests (DRM, MMC, PINCTRL).
- `Dockerfile`: Build and test environment definition.
- `build.sh`: Main kernel build script.
- `test-kernel.sh`: QEMU boot validation script.
