# Unit Testing Guideline: K3OS Kernel Validation Tools

## 1. Testing Philosophy
We test the **tools** that validate the kernel, as well as the **kernel** itself.

## 2. Testing `u-root-init` (Go)
The `u-root-init` program is the primary validation tool. It verifies essential features like Cgroups, Namespaces, OverlayFS, and Hardware Drivers.

### File Structure
- `main.go`: Generic kernel feature tests.
- `main_amd64.go`: x86_64 specific tests (e.g., HFS+ for ISO boot).
- `main_arm64.go`: ARM64 specific tests (MMC, PINCTRL).
- `main_generic.go`: Fallback for other architectures.

### Commands
- **Linting**: `task lint` (runs `golangci-lint` with `GOOS=linux`).
- **Formatting**: `task fmt` (runs `gofumpt`).
- **Unit Tests**: (Future) `GOOS=linux go test ./u-root-init/...`.

### Guidelines
- Use the `//go:build linux` tag for code that uses Linux-specific syscalls.
- Mock filesystem interactions if testing on macOS.
- Ensure all new tests in `main.go` or `main_arm64.go` have corresponding log output for the QEMU parser.

## 3. Kernel Boot Validation (Integration)
This is the primary way we ensure the kernel works.

### Commands (Run inside container)
- **Run Tests**: `./test-kernel.sh`.
- **Clean**: `task clean`.

### Success Criteria
- The output must contain: `SUCCESS: Kernel booted and validation completed`.
- A valid JUnit XML must be generated in `test-results/kernel-boot/results.xml`.

## 4. Hardware Regression Testing
- When a change affects a specific hardware target (e.g., RPi 4), manual verification on physical hardware is required before merging to `master`.
