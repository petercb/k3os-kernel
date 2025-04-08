#!/bin/bash

set -euxo pipefail

if [ "${IN_CONTAINER:-false}" != "true" ]; then
    echo "FATAL: Not running in a docker container!"
    echo "This script modifies the system, and is not safe to run outside of a container!"
    exit 1
fi

: "${FULL_VERSION=6.1.0-1036.36}"
: "${BUILD_ROOT=/tmp/build}"
: "${KERNEL_WORK=${BUILD_ROOT}/kernel-work}"
KERNEL_FLAVOUR="k3os"

if mountpoint -q "$(pwd)"; then
    git config --global --add safe.directory "$(pwd)"
fi

PROJECT_ROOT="$(git rev-parse --show-toplevel)"
DIST_DIR="${PROJECT_ROOT}/dist"
KERNEL_ROOT="${BUILD_ROOT}/kernel"

abi_suffix="-${CIRCLE_TAG:-$(git describe --tags 2>/dev/null)}"
export abi_suffix

VERSION="${FULL_VERSION%.*}${abi_suffix}-${KERNEL_FLAVOUR}"

mkdir -p "${KERNEL_WORK}"

rsync -a "${PROJECT_ROOT}/overlay/" "${KERNEL_WORK}"
cp -a "${KERNEL_WORK}/debian/changelog" "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/"
cp -a "${KERNEL_WORK}/debian.master/control.stub.in" "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/"
cp -a "${KERNEL_WORK}/debian.master/control.d/generic.inclusion-list" "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/control.d/k3os.inclusion-list"
cp -a "${KERNEL_WORK}"/debian.master/control.d/*.stub "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/control.d/"
cp -a "${KERNEL_WORK}"/debian.master/control.stub.in "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/"
cp -a "${KERNEL_WORK}"/debian.master/reconstruct "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/"

pushd "${KERNEL_WORK}"
debian/rules clean

if [ "${UPDATECONFIGS:-no}" == "yes" ]; then
    apt-get update -qq
    apt-get --assume-yes -qq install --no-install-recommends \
        gcc-12-aarch64-linux-gnu
    if ! debian/rules updateconfigs
    then
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
popd

pushd "${BUILD_ROOT}"
dpkg --install --no-triggers --force-depends \
    "linux-image-unsigned-${VERSION}_${FULL_VERSION}_${TARGETARCH}.deb" \
    "linux-modules-${VERSION}_${FULL_VERSION}_${TARGETARCH}.deb" \
    "linux-modules-extra-${VERSION}_${FULL_VERSION}_${TARGETARCH}.deb"
rm ./*.deb
popd

pushd "${KERNEL_WORK}"
debian/rules clean
popd

# Setup initrd
echo "Generating initrd"
mkdir -p "${DIST_DIR}"
depmod "${VERSION}"
mkinitramfs \
    -c gzip \
    -o "/tmp/initrd.gz" \
    "${VERSION}"
zcat /tmp/initrd.gz | cpio -id -D /tmp/initrd.old
mkdir -p /tmp/initrd.new/lib
mv /tmp/initrd.old/lib/modules /tmp/initrd.new/lib/
mv /tmp/initrd.old/lib/firmware /tmp/initrd.new/lib/
rm -rf /tmp/initrd.gz /tmp/initrd.old
pushd /tmp/initrd.new
find . | cpio -H newc -o | gzip -c -1 > "${DIST_DIR}/k3os-initrd-${TARGETARCH}.gz"
popd
rm -rf /tmp/initrd.new/lib/modules

# Assemble kernel
mkdir -p "${KERNEL_ROOT}/lib"
echo "${VERSION}" > "${KERNEL_ROOT}/version"
cp "${KERNEL_ROOT}/version" "${DIST_DIR}/k3os-kernel-version-${TARGETARCH}.txt"
mv "/boot/System.map-${VERSION}" "${KERNEL_ROOT}/System.map"
mv "/boot/config-${VERSION}" "${KERNEL_ROOT}/config"
mv "/boot/vmlinuz-${VERSION}" "${KERNEL_ROOT}/vmlinuz"
cp "${KERNEL_ROOT}/vmlinuz" "${DIST_DIR}/k3os-vmlinuz-${TARGETARCH}.img"
mv /lib/modules "${KERNEL_ROOT}/lib"

# Assemble firmware
mv /tmp/initrd.new/lib/firmware "${KERNEL_ROOT}/lib/"


pushd "${KERNEL_ROOT}"
OUTFILE="${DIST_DIR}/k3os-kernel-${TARGETARCH}.squashfs"
rm -f "${OUTFILE}"
mksquashfs . "${OUTFILE}" -no-progress
popd

# Cleanup
rm -rf "${KERNEL_ROOT}"
dpkg --remove \
    "linux-image-unsigned-${VERSION}" \
    "linux-modules-${VERSION}" \
    "linux-modules-extra-${VERSION}"
