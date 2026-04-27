# The Kernel of K3OS

![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/petercb/k3os-kernel?label=release&sort=semver)
[![CircleCI](https://dl.circleci.com/status-badge/img/gh/petercb/k3os-kernel/tree/master.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/petercb/k3os-kernel/tree/master)

This repo can build the kernel and package the firmware.

## 🤖 AI Agent Instructions
When working on this repository, please keep the following in mind:
- **Tools Path**: Most development tools (go, gofumpt, golangci-lint) are installed via Homebrew and reside in `/opt/homebrew/bin`. You may need to add this to your `PATH` or use absolute paths.
- **Go Code**: The code in `u-root-init/` is intended for `linux`. When running tests, linters, or builds on macOS, you **must** set `GOOS=linux` and a target `GOARCH` (e.g., `arm64` or `amd64`).
- **Configuration**: The single source of truth for the kernel configuration is the Ubuntu-style annotations file at `overlay/debian.k3os/config/annotations`.
- **Testing**: Kernel validation is done via QEMU using the scripts provided. See [technical.md](docs/technical.md) and [unit_testing_guideline.md](docs/unit_testing_guideline.md) for detailed workflows.
