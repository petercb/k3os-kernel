#!/bin/bash

KERNEL_VERSION="5.15.0"
KERNEL_FLAVOUR="k3os"

: "${TARGETARCH=$(uname -m)}"

set -euxo pipefail

if [ "${IN_CONTAINER:-false}" != "true" ]; then
    echo "FATAL: Not running in a docker container!"
    echo "This script modifies the system, and is not safe to run outside of a container!"
    exit 1
fi

VERSION=${CIRCLE_TAG:-$(git describe --abbrev=0 --tags)-next$(git rev-list "$(git describe --always --tags --abbrev=0)..HEAD" --count)}

PROJECT_ROOT="$(git rev-parse --show-toplevel)"
DIST_DIR="${PROJECT_ROOT}/dist"
BUILD_ROOT="${PROJECT_ROOT}/build"
KERNEL_WORK="${BUILD_ROOT}/kernel-work"

apt-get --assume-yes -qq update

pushd "/tmp"
apt-get --assume-yes -q download linux-firmware linux-source-${KERNEL_VERSION}
ls -lFa
popd

mkdir -p "${BUILD_ROOT}/kernel"
dpkg-deb -x /tmp/linux-source-${KERNEL_VERSION}_*.deb "${BUILD_ROOT}/kernel"

apt-get --assume-yes -qq install --no-install-recommends \
    bc \
    bison \
    ccache \
    cpio \
    dkms \
    dwarves \
    fakeroot \
    flex \
    gawk  \
    gcc-9 \
    gnupg2 \
    initramfs-tools \
    kernel-wedge \
    kmod \
    less \
    libelf-dev \
    libiberty-dev \
    liblz4-tool \
    libncurses-dev \
    libpci-dev \
    libssl-dev \
    libudev-dev \
    linux-libc-dev \
    locales \
    rsync \
    squashfs-tools \
    vim \
    xz-utils \
    zstd

mkdir -p "${KERNEL_WORK}"
cp -a "${BUILD_ROOT}"/kernel/usr/src/linux-source-*/debian* "${KERNEL_WORK}/"
chmod a+x "${KERNEL_WORK}"/debian*/rules
chmod a+x "${KERNEL_WORK}"/debian*/scripts/*
chmod a+x "${KERNEL_WORK}"/debian*/scripts/misc/*
mkdir -p "${KERNEL_WORK}/debian/stamps"
tar xf "${BUILD_ROOT}"/kernel/usr/src/linux-source-*/linux-source*.tar.bz2 --strip-components=1 -C "${KERNEL_WORK}"
rsync -a "${PROJECT_ROOT}/overlay/" "${KERNEL_WORK}"
cp -a "${KERNEL_WORK}/debian/changelog" "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/"
cp -a "${KERNEL_WORK}/debian.master/control.stub.in" "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/"
cp -a "${KERNEL_WORK}/debian.master/rules.d/hooks.mk" "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/rules.d/"
cp -a "${KERNEL_WORK}/debian.master/control.d/generic.inclusion-list" "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/control.d/k3os.inclusion-list"

export CCACHE_DIR="${BUILD_ROOT}/ccache"
mkdir -p "${CCACHE_DIR}"
pushd "${KERNEL_WORK}"
debian/rules clean
# see https://wiki.ubuntu.com/KernelTeam/KernelMaintenance#Overriding_module_check_failures
debian/rules binary-headers binary-${KERNEL_FLAVOUR} \
    skipabi=true \
    skipmodule=true \
    skipretpoline=true \
    skipdbg=true
popd

SOURCE_ROOT=/usr/src/root
KERNEL_ROOT=/usr/src/kernel
mkdir -p "${SOURCE_ROOT}"

dpkg-deb -x "${KERNEL_DIR}"/../linux-headers-${KERNEL_VERSION}-*-${KERNEL_FLAVOUR}_*_${TARGETARCH}.deb "${SOURCE_ROOT}"
dpkg-deb -x "${KERNEL_DIR}"/../linux-headers-${KERNEL_VERSION}-*_all.deb "${SOURCE_ROOT}"
dpkg-deb -x "${KERNEL_DIR}"/../linux-image-unsigned-${KERNEL_VERSION}-*_${TARGETARCH}.deb "${SOURCE_ROOT}"
dpkg-deb -x "${KERNEL_DIR}"/../linux-modules-${KERNEL_VERSION}-*-${KERNEL_FLAVOUR}_*_${TARGETARCH}.deb "${SOURCE_ROOT}"
dpkg-deb -x /tmp/linux-firmware_*.deb "${SOURCE_ROOT}"
dpkg-deb -x "${KERNEL_DIR}"/../linux-modules-extra-${KERNEL_VERSION}-*-${KERNEL_FLAVOUR}_*_${TARGETARCH}.deb "${SOURCE_ROOT}"
{
    echo 'r8152'
    echo 'hfs'
    echo 'hfsplus'
    echo 'nls_utf8'
    echo 'nls_iso8859_1'
} >> /etc/initramfs-tools/modules
rsync -a ${SOURCE_ROOT}/lib/ /lib/

# Create initrd
mkdir -p ${KERNEL_ROOT}/lib
mkdir -p ${KERNEL_ROOT}/headers
INITRD_ROOT=/usr/src/initrd
pushd ${INITRD_ROOT}
echo "Generate initrd"
depmod "${VERSION}"
mkinitramfs -c gzip -o ${INITRD_ROOT}.tmp "${VERSION}"
zcat ${INITRD_ROOT}.tmp | cpio -idm
rm ${INITRD_ROOT}.tmp
echo "Generate firmware and module lists"
find lib/modules -name \*.ko > ${KERNEL_ROOT}/initrd-modules
echo "lib/modules/${VERSION}/modules.order" >> ${KERNEL_ROOT}/initrd-modules
echo "lib/modules/${VERSION}/modules.builtin" >> ${KERNEL_ROOT}/initrd-modules
find lib/firmware -type f > ${KERNEL_ROOT}/initrd-firmware
find usr/lib/firmware -type f | sed 's!usr/!!' >> ${KERNEL_ROOT}/initrd-firmware
popd
rm -rf ${INITRD_ROOT}

# Copy output assets
pushd ${SOURCE_ROOT}
cp -r usr/src/linux-headers* ${KERNEL_ROOT}/headers
cp -r lib/firmware ${KERNEL_ROOT}/lib/firmware
cp -r lib/modules ${KERNEL_ROOT}/lib/modules
cp boot/System.map* ${KERNEL_ROOT}/System.map
cp boot/config* ${KERNEL_ROOT}/config
cp boot/vmlinuz-* ${KERNEL_ROOT}/vmlinuz
echo "${VERSION}" > ${KERNEL_ROOT}/version
popd

mkdir -p ${INITRD_ROOT}/lib
pushd ${KERNEL_ROOT}
tar cf - -T initrd-modules -T initrd-firmware | tar xf - -C ${INITRD_ROOT}/
depmod -b ${INITRD_ROOT} "${VERSION}"
depmod -b . "${VERSION}"
mkdir -p "${DIST_DIR}"
mksquashfs . "${DIST_DIR}/kernel-${TARGETARCH}.squashfs"
popd
