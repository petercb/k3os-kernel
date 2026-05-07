#!/bin/bash

set -euxo pipefail

: "${KERNEL_WORK=/usr/src/linux}"

# Select firmware (only for enabled drivers)
echo "Building selective firmware inclusion list"
FIRMWARE_LIST="/tmp/firmware-include.txt"
: > "${FIRMWARE_LIST}"

# Step 1: Get list of enabled kernel modules/drivers from config
ENABLED_CONFIGS="/tmp/enabled-configs.txt"
grep -E '=y|=m' "/boot/config" | sed 's/^CONFIG_//' | cut -d= -f1 \
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
            | sort -u || true)
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

sort -u "${FIRMWARE_LIST}" > /boot/firmware-list.txt
