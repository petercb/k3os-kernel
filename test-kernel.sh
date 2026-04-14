#!/bin/bash
set -e
set -o pipefail

PKGS=(
    golang-go
)

# Detect architecture
if [ -z "${TARGETARCH:-}" ]; then
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            TARGETARCH=amd64
            PKGS+=(qemu-system-x86)
            ;;
        aarch64)
            TARGETARCH=arm64
            PKGS+=(qemu-system-arm ipxe-qemu qemu-efi-aarch64)
            ;;
        *) echo "Unknown architecture: $ARCH"; exit 1 ;;
    esac
fi

# Install dependencies needed for QEMU test
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
echo "Checking files before QEMU run:"
chmod 644 "$KERNEL"
ls -lh "$KERNEL" "$INITRD"

LOG_FILE="qemu.log"
echo "Booting $KERNEL in QEMU..."
set +e
if [ "$TARGETARCH" == "amd64" ]; then
    timeout 60s qemu-system-x86_64 -machine q35 -cpu max -m 512 -append "console=ttyS0 panic=-1" -display none -serial file:"$LOG_FILE" -no-reboot -kernel "$KERNEL" -initrd "$INITRD"
elif [ "$TARGETARCH" == "arm64" ]; then
    timeout 60s qemu-system-aarch64 -machine virt -cpu max -smp 1 -m 512 -append "console=ttyAMA0 panic=-1" -display none -serial file:"$LOG_FILE" -no-reboot -kernel "$KERNEL" -initrd "$INITRD"
fi
set -e
cat "$LOG_FILE"
rm -rf "$INITRD_DIR"

# Verify output
RESULTS_DIR="test-results/kernel-boot"
mkdir -p "$RESULTS_DIR"
XML_REPORT="$RESULTS_DIR/results.xml"

echo "--- Analyzing Kernel Boot Log ---"

# Helper to write JUnit XML testcase
write_testcase() {
    local name=$1
    local status=$2 # 0 for pass, 1 for fail
    local msg=$3

    if [ "$status" -eq 0 ]; then
        echo "  <testcase name=\"$name\"/>" >> "$XML_REPORT"
    else
        echo "  <testcase name=\"$name\">" >> "$XML_REPORT"
        echo "    <failure message=\"$msg\"/>" >> "$XML_REPORT"
        echo "  </testcase>" >> "$XML_REPORT"
    fi
}

echo "<?xml version=\"1.0\" encoding=\"UTF-8\"?>" > "$XML_REPORT"
echo "<testsuite name=\"kernel-boot-$TARGETARCH\" tests=\"4\">" >> "$XML_REPORT"

# 1. Check for basic boot success
if grep -q "SUCCESS: Kernel booted and validation completed" "$LOG_FILE"; then
    echo "[PASS] Basic boot and init execution successful."
    write_testcase "Boot and Init Execution" 0
else
    echo "[FAIL] Kernel failed to execute init process correctly."
    write_testcase "Boot and Init Execution" 1 "Init process did not report completion"
    FINAL_RC=1
fi

# 2. Check for OverlayFS
if grep -q "\[PASS\] OverlayFS support detected" "$LOG_FILE"; then
    echo "[PASS] OverlayFS verified by init."
    write_testcase "OverlayFS Support" 0
else
    echo "[FAIL] OverlayFS verification failed."
    write_testcase "OverlayFS Support" 1 "OverlayFS not found in /proc/filesystems"
    FINAL_RC=1
fi

# 3. Check for Cgroup v2
if grep -q "\[PASS\] Cgroup v2 support detected" "$LOG_FILE"; then
    echo "[PASS] Cgroup v2 verified by init."
    write_testcase "Cgroup v2 Support" 0
else
    echo "[FAIL] Cgroup v2 verification failed."
    write_testcase "Cgroup v2 Support" 1 "Cgroup v2 controllers not found/mounted"
    FINAL_RC=1
fi

# 4. Check for Namespaces
if grep -q "\[PASS\] Namespace isolation (UTS) successfully tested" "$LOG_FILE"; then
    echo "[PASS] Namespaces verified by init."
    write_testcase "Namespace Support" 0
else
    echo "[FAIL] Namespace verification failed."
    write_testcase "Namespace Support" 1 "Namespace unshare test failed"
    FINAL_RC=1
fi

echo "</testsuite>" >> "$XML_REPORT"

echo "--- Analysis Complete: JUnit report generated at $XML_REPORT ---"
exit ${FINAL_RC:-0}
