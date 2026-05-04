#!/bin/bash

set -euxo pipefail

if [ "${IN_CONTAINER:-false}" != "true" ]; then
    echo "FATAL: Not running in a docker container!"
    echo "This script modifies the system, and is not safe to run outside of a container!"
    exit 1
fi

if [ -z "${FULL_VERSION:-}" ]; then
    echo "FATAL: envvar FULL_VERSION not set!"
    exit 1
fi

: "${BUILD_ROOT=/tmp/build}"
: "${KERNEL_WORK=${BUILD_ROOT}/kernel-work}"
KERNEL_FLAVOUR="k3os"

if mountpoint -q "$(pwd)"; then
    git config --global --add safe.directory "$(pwd)"
fi

PROJECT_ROOT="$(git rev-parse --show-toplevel)"
DIST_DIR="${PROJECT_ROOT}/dist"
KERNEL_ROOT="${BUILD_ROOT}/kernel"

VERSION="${FULL_VERSION%.*}-${KERNEL_FLAVOUR}"

mkdir -p "${KERNEL_WORK}"

rsync -a "${PROJECT_ROOT}/overlay/" "${KERNEL_WORK}"
cp -a "${KERNEL_WORK}/debian.master/changelog" "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/"
cp -a "${KERNEL_WORK}/debian.master/control.stub.in" "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/"
cp -a "${KERNEL_WORK}"/debian.master/control.d/*.stub "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/control.d/"
cp -a "${KERNEL_WORK}"/debian.master/control.stub.in "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/"
cp -a "${KERNEL_WORK}"/debian.master/reconstruct "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/"

pushd "${KERNEL_WORK}"
debian/rules clean

if [ "${CONFIGMODE:-no}" != "no" ]; then
    apt-get update -q
    apt-get --assume-yes -q install --no-install-recommends \
        gcc-aarch64-linux-gnu gcc-x86-64-linux-gnu
    if ! debian/rules ${CONFIGMODE}configs
    then
        sed -i \
            -e "/^CONFIG_CC_CAN_LINK/d" \
            -e "/^CONFIG_CC_VERSION_TEXT/d" \
            -e "/^CONFIG_CC_HAS_MARCH_NATIVE/d" \
            -e "/^CONFIG_X86_NATIVE_CPU/d" \
            debian.k3os/config/annotations
        cp debian.k3os/config/annotations \
            "${PROJECT_ROOT}/overlay/debian.k3os/config/annotations"
        exit 1
    fi
fi
# see https://wiki.ubuntu.com/KernelTeam/KernelMaintenance#Overriding_module_check_failures
debian/rules binary-${KERNEL_FLAVOUR} \
    skipabi=true \
    skipmodule=true \
    skipretpoline=true \
    skipdbg=true
dpkg --unpack --no-triggers --force-depends \
    "../linux-image-unsigned-${VERSION}_${FULL_VERSION}_${TARGETARCH}.deb" \
    "../linux-modules-${VERSION}_${FULL_VERSION}_${TARGETARCH}.deb"
rm ../linux-*.deb
debian/rules clean
popd

# Setup initrd
echo "Generating initrd"
mkdir -p "${DIST_DIR}"
depmod "${VERSION}"
echo "RESUME=none" > /etc/initramfs-tools/conf.d/resume
mkinitramfs \
    -c gzip \
    -o /tmp/initrd \
    "${VERSION}"
cpio -id -D /tmp/initrd.old < /tmp/initrd
mkdir -p /tmp/initrd.new/lib
mv /tmp/initrd.old/usr/lib/modules /tmp/initrd.new/lib/
mv /tmp/initrd.old/usr/lib/firmware /tmp/initrd.new/lib/
rm -rf /tmp/initrd /tmp/initrd.old

# Build early microcode cpio (must be uncompressed, prepended to initrd)
MICROCODE_DIR="/tmp/early-microcode"
MICROCODE_CPIO="/tmp/microcode.cpio"
mkdir -p "${MICROCODE_DIR}/kernel/x86/microcode"
if [ "${TARGETARCH}" = "amd64" ]; then
    echo "Assembling early microcode for amd64"
    # AMD microcode
    if [ -d /lib/firmware/amd-ucode ]; then
        cat /lib/firmware/amd-ucode/microcode_amd*.bin \
            > "${MICROCODE_DIR}/kernel/x86/microcode/AuthenticAMD.bin" 2>/dev/null || true
    fi
    # Intel microcode
    if [ -d /lib/firmware/intel-ucode ]; then
        cat /lib/firmware/intel-ucode/* \
            > "${MICROCODE_DIR}/kernel/x86/microcode/GenuineIntel.bin" 2>/dev/null || true
    fi
    # Remove empty files
    find "${MICROCODE_DIR}" -empty -delete
fi
# Create the early microcode cpio (uncompressed, newc format)
pushd "${MICROCODE_DIR}"
if find . -type f | grep -q .; then
    find . | cpio -o -H newc > "${MICROCODE_CPIO}"
    echo "Early microcode cpio created ($(du -sh "${MICROCODE_CPIO}" | cut -f1))"
else
    echo "No early microcode found for ${TARGETARCH}, skipping"
    : > "${MICROCODE_CPIO}"
fi
popd

# Assemble final initrd: early microcode + main initrd
pushd /tmp/initrd.new
MAIN_INITRD="/tmp/main-initrd.cpio.gz"
find . | cpio -H newc -o | gzip -c -1 > "${MAIN_INITRD}"
popd
if [ -s "${MICROCODE_CPIO}" ]; then
    cat "${MICROCODE_CPIO}" "${MAIN_INITRD}" > "${DIST_DIR}/k3os-initrd-${TARGETARCH}.gz"
    echo "Initrd assembled with early microcode prepended"
else
    cp "${MAIN_INITRD}" "${DIST_DIR}/k3os-initrd-${TARGETARCH}.gz"
fi
rm -rf /tmp/initrd.new /tmp/early-microcode "${MICROCODE_CPIO}" "${MAIN_INITRD}"

# Assemble kernel
mkdir -p "${KERNEL_ROOT}/lib"
echo "${VERSION}" > "${KERNEL_ROOT}/version"
cp "${KERNEL_ROOT}/version" "${DIST_DIR}/k3os-kernel-version-${TARGETARCH}.txt"
cp "/boot/System.map-${VERSION}" "${KERNEL_ROOT}/System.map"
cp "/boot/config-${VERSION}" "${KERNEL_ROOT}/config"
cp "/boot/vmlinuz-${VERSION}" "${KERNEL_ROOT}/vmlinuz"
cp "${KERNEL_ROOT}/vmlinuz" "${DIST_DIR}/k3os-vmlinuz-${TARGETARCH}.img"
cp -r /lib/modules "${KERNEL_ROOT}/lib/"

# Assemble firmware (selective — only for enabled drivers)
echo "Building selective firmware inclusion list"
FIRMWARE_LIST="/tmp/firmware-include.txt"
: > "${FIRMWARE_LIST}"

# Step 1: Get list of enabled kernel modules/drivers from config
ENABLED_CONFIGS="/tmp/enabled-configs.txt"
grep -E '=y|=m' "${KERNEL_ROOT}/config" | sed 's/^CONFIG_//' | cut -d= -f1 \
    > "${ENABLED_CONFIGS}"

# Step 2: Extract MODULE_FIRMWARE declarations from kernel source and cross-reference
#   MODULE_FIRMWARE lines look like:  MODULE_FIRMWARE("firmware/path.bin");
#   The source file path tells us which config option controls that driver.
echo "Scanning kernel source for MODULE_FIRMWARE declarations..."
grep -rh "MODULE_FIRMWARE" "${KERNEL_WORK}" --include='*.c' 2>/dev/null \
    | sed -n 's/.*MODULE_FIRMWARE("\(.*\)").*/\1/p' \
    | sort -u > /tmp/all-module-firmware.txt

# For each source file with MODULE_FIRMWARE, try to match it to a Kconfig/Makefile
# to determine if its driver is enabled. We use a heuristic: find the Makefile in
# the same directory and extract the obj- config mappings.
grep -r "MODULE_FIRMWARE" "${KERNEL_WORK}" --include='*.c' -l 2>/dev/null \
    | while read -r srcfile; do
    srcdir="$(dirname "${srcfile}")"

    # Check if this source is referenced by an enabled config in its Makefile
    makefile="${srcdir}/Makefile"
    if [ -f "${makefile}" ]; then
        # Look for lines like: obj-$(CONFIG_FOO) += srcbase.o
        matched_configs=$(grep -oP 'CONFIG_\K[A-Z0-9_]+' "${makefile}" \
            | sort -u)
        for cfg in ${matched_configs}; do
            if grep -qx "${cfg}" "${ENABLED_CONFIGS}"; then
                # This driver is enabled — include its firmware
                grep -h "MODULE_FIRMWARE" "${srcfile}" 2>/dev/null \
                    | sed -n 's/.*MODULE_FIRMWARE("\(.*\)").*/\1/p' \
                    >> "${FIRMWARE_LIST}"
                break
            fi
        done
    fi
done

# Step 3: Add SoC and platform-specific firmware (always needed for target arch)
if [ "${TARGETARCH}" = "arm64" ]; then
    cat >> "${FIRMWARE_LIST}" <<'ARMFW'
brcm/brcmfmac43455-sdio.bin
brcm/brcmfmac43455-sdio.txt
brcm/brcmfmac43455-sdio.clm_blob
brcm/brcmfmac43456-sdio.bin
brcm/brcmfmac43456-sdio.txt
brcm/brcmfmac43456-sdio.clm_blob
cypress/cyfmac43455-sdio.bin
cypress/cyfmac43455-sdio.clm_blob
raspberrypi/bootloader-2711/latest/pieeprom-2711-latest.bin
raspberrypi/bootloader-2712/latest/pieeprom-2712-latest.bin
rockchip/dptx.bin
ARMFW
fi

if [ "${TARGETARCH}" = "amd64" ]; then
    cat >> "${FIRMWARE_LIST}" <<'X86FW'
i915/
amdgpu/
X86FW
fi

# Step 4: Copy only the firmware we need
sort -u "${FIRMWARE_LIST}" > /tmp/firmware-sorted.txt
mkdir -p "${KERNEL_ROOT}/lib/firmware"
firmware_count=0
while IFS= read -r fw_path; do
    [ -z "${fw_path}" ] && continue
    src="/lib/firmware/${fw_path}"
    dst="${KERNEL_ROOT}/lib/firmware/${fw_path}"
    if [ -d "${src}" ]; then
        # Directory entry (e.g., i915/, amdgpu/) — copy whole dir
        mkdir -p "${dst}"
        cp -a "${src}/." "${dst}/"
        firmware_count=$((firmware_count + 1))
    elif [ -f "${src}" ]; then
        mkdir -p "$(dirname "${dst}")"
        cp -a "${src}" "${dst}"
        firmware_count=$((firmware_count + 1))
    elif ls "${src}"* 2>/dev/null | head -1 > /dev/null; then
        # Glob match (e.g., firmware with version suffixes)
        mkdir -p "$(dirname "${dst}")"
        cp -a "${src}"* "$(dirname "${dst}")/"
        firmware_count=$((firmware_count + 1))
    else
        echo "[WARN] Firmware not found: ${fw_path}"
    fi
done < /tmp/firmware-sorted.txt
echo "Copied ${firmware_count} firmware entries (selective)"
rm -f "${FIRMWARE_LIST}" "${ENABLED_CONFIGS}" /tmp/all-module-firmware.txt /tmp/firmware-sorted.txt

pushd "${KERNEL_ROOT}"
OUTFILE="${DIST_DIR}/k3os-kernel-${TARGETARCH}.squashfs"
rm -f "${OUTFILE}"
echo "Writing squashfs to ${OUTFILE}:"
mksquashfs . "${OUTFILE}" -no-progress -info
popd

# Cleanup
rm -rf "${KERNEL_ROOT}"
dpkg --remove \
    "linux-image-unsigned-${VERSION}" \
    "linux-modules-${VERSION}"
