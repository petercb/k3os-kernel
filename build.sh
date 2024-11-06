#!/bin/bash

set -euxo pipefail

if [ "${IN_CONTAINER:-false}" != "true" ]; then
    echo "FATAL: Not running in a docker container!"
    echo "This script modifies the system, and is not safe to run outside of a container!"
    exit 1
fi

KERNEL_VERSION="5.15.0"
KERNEL_FLAVOUR="k3os"

PACKAGE_LIST=(
    "bc"
    "bison"
    "ccache"
    "cpio"
    "dwarves"
    "fakeroot"
    "flex"
    "gawk"
    "initramfs-tools"
    "kernel-wedge"
    "kmod"
    "libelf-dev"
    "libiberty-dev"
    "liblz4-tool"
    "libncurses-dev"
    "libpci-dev"
    "libssl-dev"
    "libudev-dev"
    "linux-libc-dev"
    "locales"
    "rsync"
    "squashfs-tools"
)

case "${TARGETARCH=$(uname -m)}" in
    amd64)
        PACKAGE_LIST+=("gcc-aarch64-linux-gnu")
        ;;
    arm64)
        PACKAGE_LIST+=("gcc-x86-64-linux-gnu")
        ;;
    *)
        echo "ERROR: Unsupported TARGETARCH '${TARGETARCH}' !!"
        exit 1
esac

if mountpoint -q "$(pwd)"; then
    git config --global --add safe.directory "$(pwd)"
fi

PROJECT_ROOT="$(git rev-parse --show-toplevel)"
DIST_DIR="${PROJECT_ROOT}/dist"
BUILD_ROOT="${PROJECT_ROOT}/build"
KERNEL_ORIG="${BUILD_ROOT}/kernel-orig"
KERNEL_WORK="${BUILD_ROOT}/kernel-work"
DOWNLOAD_DIR="${BUILD_ROOT}/artifacts"
SOURCE_ROOT="${BUILD_ROOT}/root"
KERNEL_ROOT="${BUILD_ROOT}/kernel"
INITRD_ROOT="${BUILD_ROOT}/initrd"
INITRD_CONFDIR="${BUILD_ROOT}/initrd-conf"

apt-get --assume-yes -qq update

apt-get --assume-yes -qq install --no-install-recommends "${PACKAGE_LIST[@]}"

rm -rf "${DOWNLOAD_DIR}"
mkdir -p "${DOWNLOAD_DIR}"
pushd "${DOWNLOAD_DIR}"
apt-get --assume-yes -q download linux-firmware linux-source-${KERNEL_VERSION}
ls -lFa
VERSION=$(echo linux-source-${KERNEL_VERSION}_*_all.deb | sed -e "s/^linux-source-${KERNEL_VERSION}_//" -e "s/\.[[:digit:]]\+_all\.deb$//")-${KERNEL_FLAVOUR}
popd

rm -rf "${KERNEL_ORIG}"
mkdir -p "${KERNEL_ORIG}"
dpkg-deb -x "${DOWNLOAD_DIR}"/linux-source-${KERNEL_VERSION}_*.deb "${KERNEL_ORIG}"

rm -rf "${KERNEL_WORK}"
mkdir -p "${KERNEL_WORK}"
cp -a "${KERNEL_ORIG}"/usr/src/linux-source-*/debian* "${KERNEL_WORK}/"
chmod a+x "${KERNEL_WORK}"/debian*/rules
chmod a+x "${KERNEL_WORK}"/debian*/scripts/*
chmod a+x "${KERNEL_WORK}"/debian*/scripts/misc/*
mkdir -p "${KERNEL_WORK}/debian/stamps"
tar xf "${KERNEL_ORIG}"/usr/src/linux-source-*/linux-source*.tar.bz2 \
    --strip-components=1 -C "${KERNEL_WORK}"
rsync -a "${PROJECT_ROOT}/overlay/" "${KERNEL_WORK}"
cp -a "${KERNEL_WORK}/debian/changelog" "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/"
cp -a "${KERNEL_WORK}/debian.master/control.stub.in" "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/"
cp -a "${KERNEL_WORK}/debian.master/rules.d/hooks.mk" "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/rules.d/"
cp -a "${KERNEL_WORK}/debian.master/control.d/generic.inclusion-list" "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/control.d/k3os.inclusion-list"
cp -a "${KERNEL_WORK}"/debian.master/control.d/*.stub "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/control.d/"
cp -a "${KERNEL_WORK}"/debian.master/control.stub.in "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/"
cp -a "${KERNEL_WORK}"/debian.master/reconstruct "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}/"

pushd "${KERNEL_WORK}"
debian/rules clean

if [ "${UPDATECONFIGS:-no}" == "yes" ]; then
    debian/rules updateconfigs
fi
# see https://wiki.ubuntu.com/KernelTeam/KernelMaintenance#Overriding_module_check_failures
debian/rules binary-headers binary-${KERNEL_FLAVOUR} \
    skipabi=true \
    skipmodule=true \
    skipretpoline=true \
    skipdbg=true
popd

rm -rf "${SOURCE_ROOT}"
mkdir -p "${SOURCE_ROOT}"
mkdir -p "${DIST_DIR}"

pushd "${KERNEL_WORK}/.."
for deb in \
    linux-image-unsigned-${KERNEL_VERSION}-*_${TARGETARCH}.deb \
    linux-modules-${KERNEL_VERSION}-*-${KERNEL_FLAVOUR}_*_${TARGETARCH}.deb \
    linux-modules-extra-${KERNEL_VERSION}-*-${KERNEL_FLAVOUR}_*_${TARGETARCH}.deb
do
    dpkg-deb -x "${deb}" "${SOURCE_ROOT}"
    rm "${deb}"
done
dpkg-deb -x "${DOWNLOAD_DIR}"/linux-firmware_*.deb "${SOURCE_ROOT}"
popd

# Setup initrd
mkdir -p "${INITRD_CONFDIR}/scripts"
cp /etc/initramfs-tools/initramfs.conf "${INITRD_CONFDIR}/"
cat <<EOF > "${INITRD_CONFDIR}/modules"
r8152
EOF

rm -rf "/lib/modules/${VERSION}"
rsync -a "${SOURCE_ROOT}/lib/" /lib/

# Create initrd packing lists
rm -rf "${INITRD_ROOT}"
mkdir -p "${INITRD_ROOT}"
pushd "${INITRD_ROOT}"
echo "Generate initrd"
depmod "${VERSION}"
mkinitramfs -d "${INITRD_CONFDIR}" -c gzip \
    -o "${DIST_DIR}/k3os-initrd-${TARGETARCH}.gz" "${VERSION}"
popd

# Assemble kernel
mkdir -p "${KERNEL_ROOT}/lib"
mkdir -p "${KERNEL_ROOT}/headers"
pushd "${SOURCE_ROOT}"
mv lib/firmware "${KERNEL_ROOT}/lib/firmware"
mv lib/modules "${KERNEL_ROOT}/lib/modules"
mv boot/System.map* "${KERNEL_ROOT}/System.map"
mv boot/config* "${KERNEL_ROOT}/config"
mv boot/vmlinuz-* "${KERNEL_ROOT}/vmlinuz"
echo "${VERSION}" > "${KERNEL_ROOT}/version"
popd

pushd "${KERNEL_ROOT}"
depmod -b . "${VERSION}"
OUTFILE="${DIST_DIR}/k3os-kernel-${TARGETARCH}.squashfs"
rm -f "${OUTFILE}"
mksquashfs . "${OUTFILE}"
popd
rm -rf "${KERNEL_ROOT}"
