# syntax=docker/dockerfile:1

FROM buildpack-deps:jammy

ARG TARGETARCH
ARG KERNEL_VERSION
ARG UBUNTU_BUILD
ARG UBUNTU_FLAVOUR
ARG ABI_VERSION
ARG IMAGE_VERSION="${KERNEL_VERSION}-${UBUNTU_BUILD}-${UBUNTU_FLAVOUR}"
ARG FULL_VERSION="${KERNEL_VERSION}-${UBUNTU_BUILD}.${ABI_VERSION}"

ENV TARGETARCH=${TARGETARCH}
ENV KERNEL_VERSION=${KERNEL_VERSION}
ENV FULL_VERSION=${FULL_VERSION}
ENV IN_CONTAINER="true"
ENV BUILD_ROOT=/usr/src
ENV KERNEL_WORK="${BUILD_ROOT}/linux-${UBUNTU_FLAVOUR}-6.1-${KERNEL_VERSION}"
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Etc/UTC

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /usr/src
# hadolint ignore=DL3008
RUN <<-EOF
    sed -i 's/^#[[:space:]]\+deb-src[[:space:]]/deb-src /' /etc/apt/sources.list
    # sed -i 's/^Types:.*$/Types: deb deb-src/' /etc/apt/sources.list.d/ubuntu.sources
    apt-get update
    apt-get build-dep -y --no-install-recommends \
        linux \
        linux-image-unsigned-${IMAGE_VERSION}
    apt-get install -y --no-install-recommends \
        cpio \
        dwarves \
        git \
        initramfs-tools-core \
        libncurses-dev \
        linux-firmware \
        llvm \
        rsync \
        squashfs-tools
    [ "${TARGETARCH}" == "arm64" ] && apt-get install -y linux-firmware-raspi
    apt-get source --no-install-recommends \
        linux-image-unsigned-${IMAGE_VERSION}
    apt-get clean
    rm -rf /var/lib/apt/lists/*
EOF

WORKDIR "${KERNEL_WORK}"
RUN <<-EOF
    chmod a+x debian/rules
    chmod a+x debian/scripts/*
    chmod a+x debian/scripts/misc/*
EOF

WORKDIR /root/project
