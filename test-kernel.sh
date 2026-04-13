#!/bin/bash
set -ex
set -o pipefail

# Detect architecture
if [ -z "${TARGETARCH:-}" ]; then
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            TARGETARCH=amd64
            PKGS="qemu-system-x86"
            ;;
        aarch64)
            TARGETARCH=arm64
            PKGS="qemu-system-arm ipxe-qemu qemu-efi-aarch64"
            ;;
        *) echo "Unknown architecture: $ARCH"; exit 1 ;;
    esac
fi

# Install dependencies needed for QEMU test
apt-get update -q
apt-get install -yq --no-install-recommends golang-go ${PKGS}

PROJECT_ROOT=$(pwd)
KERNEL="${PROJECT_ROOT}/dist/k3os-vmlinuz-${TARGETARCH}.img"
INITRD_DIR=$(mktemp -d)
INITRD="${INITRD_DIR}/test-initrd.gz"

if [ ! -f "$KERNEL" ]; then
    echo "Kernel $KERNEL not found!"
    exit 1
fi

export GOPATH=/root/go
export PATH=${GOPATH}/bin:${PATH}

# Install u-root
go install github.com/u-root/u-root@v0.16.0

# build u-root initramfs for testing
cp -r "${PROJECT_ROOT}/u-root-init" "$INITRD_DIR"
pushd "$INITRD_DIR"

# u-root now requires a Go workspace for multi-module builds
go work init
go work use ./u-root-init
u-root -o "$(basename "$INITRD")" \
       -defaultsh="" \
       -initcmd test-init \
       test-init
popd

if [ ! -f "$INITRD" ]; then
    echo "Test initrd $INITRD not found!"
    exit 1
fi

# Run QEMU
echo "Checking files before QEMU run:"
chmod 644 "$KERNEL"
ls -lh "$KERNEL" "$INITRD"

echo "Booting $KERNEL in QEMU..."
set +e
if [ "$TARGETARCH" == "amd64" ]; then
    timeout 60s qemu-system-x86_64 -machine q35 -cpu max -m 512 -append "console=ttyS0 panic=-1" -display none -serial file:qemu.log -no-reboot -kernel "$KERNEL" -initrd "$INITRD"
elif [ "$TARGETARCH" == "arm64" ]; then
    timeout 60s qemu-system-aarch64 -machine virt -cpu max -smp 1 -m 512 -append "console=ttyAMA0 panic=-1" -display none -serial file:qemu.log -no-reboot -kernel "$KERNEL" -initrd "$INITRD"
fi
rc=$?
set -e
cat qemu.log
rm -rf "$INITRD_DIR"

# Verify output
if grep -q "SUCCESS: Kernel booted and executed init" qemu.log; then
    echo "Kernel boot test passed for $TARGETARCH!"
    exit 0
else
    echo "Kernel boot test failed with exit code $rc for $TARGETARCH!"
    exit 1
fi
