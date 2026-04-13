#!/bin/bash
set -ex
set -o pipefail

# Install dependencies needed for QEMU test
apt-get update -q
apt-get install -yq --no-install-recommends busybox-static cpio
if [ "${TARGETARCH=amd64}" == "amd64" ]; then
    apt-get install -yq --no-install-recommends qemu-system-x86
    KERNEL=dist/k3os-vmlinuz-amd64.img
    QPROC=(qemu-system-x86_64 -machine q35 -cpu max -m 512 -append "console=ttyS0 panic=-1" -nographic -no-reboot)
elif [ "${TARGETARCH}" == "arm64" ]; then
    apt-get install -yq --no-install-recommends qemu-system-arm ipxe-qemu qemu-efi-aarch64
    KERNEL=dist/k3os-vmlinuz-arm64.img
    QPROC=(qemu-system-aarch64 -machine virt -cpu max -smp 1 -m 512 -append "console=ttyAMA0 panic=-1" -nographic -no-reboot)
else
    echo "Unsupported arch: $TARGETARCH"
    exit 1
fi

if [ ! -f "$KERNEL" ]; then
    echo "Kernel $KERNEL not found!"
    exit 1
fi

# Prepare a minimal busybox initrd to verify user space executes
rm -rf /tmp/test-initramfs
mkdir -p /tmp/test-initramfs/bin
cp /bin/busybox /tmp/test-initramfs/bin/busybox

cat << 'EOF' > /tmp/test-initramfs/init
#!/bin/busybox sh
/bin/busybox echo "SUCCESS: Kernel booted and executed init"
/bin/busybox poweroff -f
EOF
chmod +x /tmp/test-initramfs/init

# Compress it
pushd /tmp/test-initramfs
find . -print0 | cpio --null -o -H newc | gzip -c -1 > /tmp/test-initrd.gz
popd
mkdir -p dist
mv /tmp/test-initrd.gz dist/test-initrd.gz

# Run QEMU
echo "Booting $KERNEL in QEMU..."
if ! timeout 60s "${QPROC[@]}" -kernel "$KERNEL" -initrd dist/test-initrd.gz 2>&1 | tee qemu.log; then
    rc=${PIPESTATUS[0]}
    echo "Kernel boot test failed with exit code $rc for $TARGETARCH!"
    exit $rc
fi

# Verify output
if grep -q "SUCCESS: Kernel booted and executed init" qemu.log; then
    echo "Kernel boot test passed for $TARGETARCH!"
    exit 0
else
    echo "Kernel boot test failed for $TARGETARCH!"
    exit 1
fi
