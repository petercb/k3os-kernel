# syntax=docker/dockerfile:1

FROM buildpack-deps:jammy

ARG TARGETARCH
ARG KERNEL_VERSION="5.15.0"
ARG KERNEL_ORIG="/usr/src/linux-source-${KERNEL_VERSION}"

ENV TARGETARCH=${TARGETARCH}
ENV KERNEL_VERSION=${KERNEL_VERSION}
ENV IN_CONTAINER="true"
ENV BUILD_ROOT=/tmp/build
ENV KERNEL_WORK="${BUILD_ROOT}/kernel-work"

# hadolint ignore=DL3008
RUN <<-EOF
    apt-get update
    apt-get install -y --no-install-recommends \
        bc \
        bison \
        ccache \
        cpio \
        dwarfdump \
        dwarves \
        fakeroot \
        flex \
        gawk \
        kernel-wedge \
        kmod \
        libelf-dev \
        libiberty-dev \
        liblz4-tool \
        libncurses-dev \
        libpci-dev \
        libssl-dev \
        libudev-dev \
        linux-base \
        linux-firmware \
        linux-libc-dev \
        "linux-source-${KERNEL_VERSION}" \
        locales \
        rsync \
        squashfs-tools \
        zstd
    mkdir -p "${KERNEL_WORK}"
    cp -a "${KERNEL_ORIG}"/debian* "${KERNEL_WORK}/"
    chmod a+x "${KERNEL_WORK}"/debian*/rules
    chmod a+x "${KERNEL_WORK}"/debian*/scripts/*
    chmod a+x "${KERNEL_WORK}"/debian*/scripts/misc/*
    mkdir -p "${KERNEL_WORK}/debian/stamps"
    tar xf "${KERNEL_ORIG}/linux-source-${KERNEL_VERSION}.tar.bz2" \
        --strip-components=1 -C "${KERNEL_WORK}"
    rm -rf "${KERNEL_ORIG}"
    apt-get clean
    rm -rf /var/lib/apt/lists/*
EOF

WORKDIR /root/project
