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
echo "Checking files before QEMU run:"
chmod 644 "$KERNEL"
ls -lh "$KERNEL" "$INITRD"

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
set -e
cat "$LOG_FILE"
rm -rf "$INITRD_DIR"

# Verify output
RESULTS_DIR="test-results/kernel-boot"
mkdir -p "$RESULTS_DIR"
XML_REPORT="$RESULTS_DIR/results.xml"

echo "--- Analyzing Kernel Boot Log ---"

# Define tests: "DisplayName|SearchPattern|FailureMessage"
TESTS=(
    "Boot and Init Execution|SUCCESS: Kernel booted and validation completed|Init process did not report completion"
    "OverlayFS Support|\[PASS\] OverlayFS support detected|OverlayFS not found in /proc/filesystems"
    "Cgroup v2 Support|\[PASS\] Cgroup v2 support detected|Cgroup v2 controllers not found/mounted"
    "Namespace Support|\[PASS\] Namespace isolation (UTS) successfully tested|Namespace unshare test failed"
    "USB Storage Support|\[PASS\] USB Storage support detected|USB storage driver not found in /sys"
    "Veth Support|\[PASS\] Veth support detected|Veth driver not found in /sys or kallsyms"
    "Bridge Support|\[PASS\] Bridge support detected|Bridge driver not found in /sys or kallsyms"
)

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
echo "<testsuite name=\"kernel-boot-$TARGETARCH\" tests=\"${#TESTS[@]}\">" >> "$XML_REPORT"

FINAL_RC=0
for test_def in "${TESTS[@]}"; do
    IFS="|" read -r name pattern fail_msg <<< "$test_def"

    if grep -q "$pattern" "$LOG_FILE"; then
        echo "[PASS] $name successful."
        write_testcase "$name" 0
    else
        echo "[FAIL] $name failed."
        write_testcase "$name" 1 "$fail_msg"
        FINAL_RC=1
    fi
done

echo "</testsuite>" >> "$XML_REPORT"

echo "--- Analysis Complete: JUnit report generated at $XML_REPORT ---"
exit ${FINAL_RC}
