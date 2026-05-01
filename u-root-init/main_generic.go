//go:build !amd64 && !arm64

package main

// currentArch is empty for unsupported architectures;
// only ArchAll tests will run.
const currentArch = ArchAll
