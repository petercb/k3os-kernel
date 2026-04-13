#!/bin/bash
set -e
set -o pipefail

# Detect architecture
if [ -z "${TARGETARCH:-}" ]; then
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            TARGETARCH=amd64
            ;;
        aarch64)
            TARGETARCH=arm64
            ;;
        *) echo "Unknown architecture: $ARCH"; exit 1 ;;
    esac
fi

PKGS=(
    golang-go
)

case "${TARGETARCH}" in
    amd64) PKGS+=(qemu-system-x86) ;;
    arm64) PKGS+=(qemu-system-arm ipxe-qemu qemu-efi-aarch64) ;;
    *) echo "Unknown architecture: ${TARGETARCH}"; exit 1 ;;
esac

# Install dependencies needed for QEMU test
echo "Installing packages: ${PKGS[*]}"
apt-get update -q
apt-get install -yq --no-install-recommends "${PKGS[@]}"

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
echo "Checking files before QEMU run"
chmod 644 "$KERNEL"

LOG_FILE="qemu.log"
echo "Booting $KERNEL in QEMU..."
set +e
case "${TARGETARCH}" in
    amd64)
        timeout 60s qemu-system-x86_64 \
            -machine q35 \
            -cpu max \
            -m 512 \
            -append "console=ttyS0 panic=-1" \
            -display none \
            -serial file:"$LOG_FILE" \
            -no-reboot \
            -kernel "$KERNEL" \
            -initrd "$INITRD"
        ;;
    arm64)
        timeout 60s qemu-system-aarch64 \
            -machine virt \
            -cpu max \
            -smp 1 \
            -m 512 \
            -append "console=ttyAMA0 panic=-1" \
            -display none \
            -serial file:"$LOG_FILE" \
            -no-reboot \
            -kernel "$KERNEL" \
            -initrd "$INITRD"
        ;;
    *) echo "Unknown architecture: ${TARGETARCH}"; exit 1 ;;
esac
QEMU_RC=$?
set -e

if [ $QEMU_RC -eq 124 ]; then
    echo "[WARN] QEMU timed out (60s). Logic may be incomplete."
elif [ $QEMU_RC -ne 0 ]; then
    echo "[WARN] QEMU exited with non-zero code: $QEMU_RC"
fi
cat "$LOG_FILE"
rm -rf "$INITRD_DIR"

# Verify output
RESULTS_DIR="test-results/kernel-boot"
rm -rf "$RESULTS_DIR"
mkdir -p "$RESULTS_DIR"
XML_REPORT="$RESULTS_DIR/results.xml"

echo "--- Analyzing Kernel Boot Log ---"

# Extract JUnit XML from log
# The Go program wraps the XML in markers for easy extraction
sed -n '/--- JUNIT START ---/,/--- JUNIT END ---/p' "$LOG_FILE" | grep -v -- "--- JUNIT" > "$XML_REPORT"

if [ ! -s "$XML_REPORT" ]; then
    echo "[FAIL] JUnit report not found in log!"
    exit 1
fi

# Check for overall success flag from the Go init
if grep -q "SUCCESS: Kernel booted and validation completed" "$LOG_FILE"; then
    echo "[PASS] Kernel boot validation successful."
    exit 0
else
    echo "[FAIL] Kernel boot validation failed."
    exit 1
fi
