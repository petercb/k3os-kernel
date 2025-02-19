#!/bin/bash

set -euxo pipefail

if [ "${IN_CONTAINER:-false}" != "true" ]; then
    echo "FATAL: Not running in a docker container!"
    echo "This script modifies the system, and is not safe to run outside of a container!"
    exit 1
fi

: "${KERNEL_VERSION=5.15.0}"
: "${BUILD_ROOT=/tmp/build}"
: "${KERNEL_WORK=${BUILD_ROOT}/kernel-work}"
KERNEL_FLAVOUR="k3os"

if mountpoint -q "$(pwd)"; then
    git config --global --add safe.directory "$(pwd)"
fi

PROJECT_ROOT="$(git rev-parse --show-toplevel)"
DIST_DIR="${PROJECT_ROOT}/dist"
KERNEL_ROOT="${BUILD_ROOT}/kernel"

FULL_VERSION=$(dpkg-query --show --showformat='${Version}' "linux-source-${KERNEL_VERSION}")
VERSION="${FULL_VERSION%.*}-${KERNEL_FLAVOUR}"

mkdir -p "${KERNEL_WORK}"

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
    apt-get update -qq
    apt-get --assume-yes -qq install --no-install-recommends \
        gcc-aarch64-linux-gnu gcc-x86-64-linux-gnu
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

# Assemble kernel
mkdir -p "${KERNEL_ROOT}/lib" "${DIST_DIR}"
echo "${VERSION}" > "${KERNEL_ROOT}/version"
cp "${KERNEL_ROOT}/version" "${DIST_DIR}/k3os-kernel-version-${TARGETARCH}.txt"
mv "/boot/System.map-${VERSION}" "${KERNEL_ROOT}/System.map"
mv "/boot/config-${VERSION}" "${KERNEL_ROOT}/config"
mv "/boot/vmlinuz-${VERSION}" "${KERNEL_ROOT}/vmlinuz"
cp "${KERNEL_ROOT}/vmlinuz" "${DIST_DIR}/k3os-vmlinuz-${TARGETARCH}.img"
mv /lib/modules "${KERNEL_ROOT}/lib"

# Assemble firmware
mkdir -p "${KERNEL_ROOT}/lib/firmware/intel/ice"
cp -a /lib/firmware/intel/ice/ddp "${KERNEL_ROOT}/lib/firmware/intel/ice"
case "${TARGETARCH}" in
    arm64)
        cp -a /lib/firmware/rtl_nic "${KERNEL_ROOT}/lib/firmware"
        mkdir -p "${KERNEL_ROOT}/lib/firmware/nvidia"
        cp -a /lib/firmware/nvidia/tegra210 "${KERNEL_ROOT}/lib/firmware/nvidia"
        ;;
    amd64)
        cp -a /lib/firmware/bnx2x "${KERNEL_ROOT}/lib/firmware"
        cp -a /lib/firmware/bnx2 "${KERNEL_ROOT}/lib/firmware"
        ;;
esac

pushd "${KERNEL_ROOT}"
OUTFILE="${DIST_DIR}/k3os-kernel-${TARGETARCH}.squashfs"
rm -f "${OUTFILE}"
mksquashfs . "${OUTFILE}" -no-progress -comp zstd
popd

# Cleanup
rm -rf "${KERNEL_ROOT}"
dpkg --remove \
    "linux-image-unsigned-${VERSION}" \
    "linux-modules-${VERSION}" \
    "linux-modules-extra-${VERSION}"
