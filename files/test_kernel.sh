#!/bin/bash

set -euxo pipefail

: ${INITRD="/tmp/test-initrd.cpio"}
: ${KERNEL="/tmp/vmlinuz"}
: ${LOG_FILE="/tmp/qemu.log"}
: ${XML_REPORT="results.xml"}
: ${TARGETARCH=arm64}

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

# Verify output
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
