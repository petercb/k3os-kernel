# syntax=docker/dockerfile:1.7.0

FROM buildpack-deps:jammy AS kernel-stage1

ARG KERNEL_VERSION
ARG KERNEL_PATCH
ARG KERNEL_FLAVOUR="generic"
ARG TARGETARCH
ARG DOWNLOADS=/usr/src/downloads
ARG KERNEL_DIR=/tmp/build/kernel

ENV VERSION="${KERNEL_VERSION}-${KERNEL_FLAVOUR}"

WORKDIR "${DOWNLOADS}"

SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

COPY --link build/ubuntu-firmware.deb .
COPY --link build/ubuntu-kernel.deb .

WORKDIR "${DOWNLOADS}/kernel"

RUN dpkg-deb -x ${DOWNLOADS}/ubuntu-kernel.deb .

WORKDIR "${KERNEL_DIR}"

# hadolint ignore=DL3008
RUN <<-EOF
   apt-get --assume-yes -qq update
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
      vim \
      xz-utils \
      zstd
   apt-get -qq clean
   rm -rf /var/lib/apt/lists/*
   rsync -aq ${DOWNLOADS}/kernel/usr/src/linux-source-*/debian* ./
   tar xf ${DOWNLOADS}/kernel/usr/src/linux-source-*/linux-source*.tar.bz2 --strip-components=1
EOF

COPY patches /tmp/patches

RUN <<-EOF
   shopt -s globstar nullglob
   for p in /tmp/patches/*.patch; do
      echo "applying $p"
      patch -p1 -i "$p"
   done
EOF

WORKDIR "${KERNEL_DIR}/debian/stamps"

# some hacking
# hadolint ignore=DL4005
RUN <<-EOF
   chmod a+x ${KERNEL_DIR}/debian*/scripts/*
   chmod a+x ${KERNEL_DIR}/debian*/scripts/misc/*
   rm -f /bin/sh && ln -s /bin/bash /bin/sh
EOF

WORKDIR "${KERNEL_DIR}"

ENV CCACHE_DIR=/ccache

RUN --mount=type=cache,target=/ccache/ <<-EOF
   unset -v KERNEL_DIR
   debian/rules clean
   # see https://wiki.ubuntu.com/KernelTeam/KernelMaintenance#Overriding_module_check_failures
   debian/rules binary-headers binary-${KERNEL_FLAVOUR} \
      do_zfs=false \
      do_dkms_nvidia=false \
      do_dkms_nvidia_server=false \
      skipabi=true \
      skipmodule=true \
      skipretpoline=true
   ccache -s
EOF

WORKDIR /usr/src/root

RUN <<-EOF
   dpkg-deb -x ${KERNEL_DIR}/../linux-headers-5.*generic*.deb .
   dpkg-deb -x ${KERNEL_DIR}/../linux-headers-5.*all.deb .
   dpkg-deb -x ${KERNEL_DIR}/../linux-image-unsigned-5.*.deb .
   dpkg-deb -x ${KERNEL_DIR}/../linux-modules-5.*.deb .
   dpkg-deb -x ${DOWNLOADS}/ubuntu-firmware.deb .
   dpkg-deb -x ${KERNEL_DIR}/../linux-modules-extra-5.*.deb .
   {
      echo 'r8152'
      echo 'hfs'
      echo 'hfsplus'
      echo 'nls_utf8'
      echo 'nls_iso8859_1'
   } >> /etc/initramfs-tools/modules
   rsync -aq /usr/src/root/lib/ /lib/
EOF

# Create initrd
WORKDIR /output/lib
WORKDIR /output/headers
WORKDIR /usr/src/initrd
RUN <<-EOF
    echo "Generate initrd"
    depmod "${VERSION}"
    mkinitramfs -c gzip -o /usr/src/initrd.tmp "${VERSION}"
    zcat /usr/src/initrd.tmp | cpio -idm
    rm /usr/src/initrd.tmp
    echo "Generate firmware and module lists"
    find lib/modules -name \*.ko > /output/initrd-modules
    echo "lib/modules/${VERSION}/modules.order" >> /output/initrd-modules
    echo "lib/modules/${VERSION}/modules.builtin" >> /output/initrd-modules
    find lib/firmware -type f > /output/initrd-firmware
    find usr/lib/firmware -type f | sed 's!usr/!!' >> /output/initrd-firmware
EOF

# Copy output assets
WORKDIR /usr/src/root
RUN <<-EOF
    cp -r usr/src/linux-headers* /output/headers
    cp -r lib/firmware /output/lib/firmware
    cp -r lib/modules /output/lib/modules
    cp boot/System.map* /output/System.map
    cp boot/config* /output/config
    cp boot/vmlinuz-* /output/vmlinuz
    echo "${VERSION}" > /output/version
EOF


FROM alpine:3.17.7 AS kernel

SHELL ["/bin/ash", "-euo", "pipefail", "-c"]

# hadolint ignore=DL3018
RUN apk add --no-cache squashfs-tools
COPY --link --from=kernel-stage1 /output/ /usr/src/kernel/

WORKDIR /usr/src/initrd/lib
WORKDIR /usr/src/kernel
# hadolint ignore=DL4006
RUN <<-EOF
    tar cf - -T initrd-modules -T initrd-firmware \
        | tar xf - -C /usr/src/initrd/
    depmod -b /usr/src/initrd "$(cat /usr/src/kernel/version)"
EOF

WORKDIR /output
WORKDIR /usr/src/kernel
RUN <<-EOF
    depmod -b . "$(cat /usr/src/kernel/version)"
    mksquashfs . "/output/kernel.squashfs" -no-progress
EOF
