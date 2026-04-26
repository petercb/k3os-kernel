# syntax=docker/dockerfile:1

ARG UBUNTU_NAME=resolute

FROM buildpack-deps:${UBUNTU_NAME}

ARG TARGETARCH
ARG KERNEL_VERSION
ARG UBUNTU_BUILD
ARG UBUNTU_FLAVOUR
ARG UBUNTU_NAME
ARG ABI_VERSION
ARG IMAGE_VERSION="${KERNEL_VERSION}-${UBUNTU_BUILD}-${UBUNTU_FLAVOUR}"
ARG FULL_VERSION="${KERNEL_VERSION}-${UBUNTU_BUILD}.${ABI_VERSION}"

ENV TARGETARCH=${TARGETARCH}
ENV FULL_VERSION=${FULL_VERSION}
ENV IN_CONTAINER="true"
ENV KERNEL_WORK="/usr/src/linux"
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Etc/UTC

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# hadolint ignore=DL3008
RUN <<-EOF
    sed -i 's/^Types:.*$/Types: deb deb-src/' /etc/apt/sources.list.d/ubuntu.sources
    apt-get update
    apt-get build-dep -y --no-install-recommends linux
    apt-get install -y --no-install-recommends \
        cpio \
        dwarves \
        git \
        initramfs-tools-core \
        libncurses-dev \
        linux-firmware-misc \
        linux-firmware-realtek \
        llvm \
        rsync \
        squashfs-tools
    [ "${TARGETARCH}" == "arm64" ] && apt-get install -y --no-install-recommends \
        linux-firmware-raspi
    [ "${TARGETARCH}" == "amd64" ] && apt-get install -y --no-install-recommends \
        linux-firmware-amd-misc \
        linux-firmware-amd-graphics \
        linux-firmware-intel-misc \
        linux-firmware-intel-graphics
    apt-get clean
    rm -rf /var/lib/apt/lists/*
EOF

WORKDIR /tmp
RUN <<-EOF
    wget -q https://launchpad.net/ubuntu/+archive/primary/+sourcefiles/linux/${FULL_VERSION}/linux_${KERNEL_VERSION}.orig.tar.gz
    wget -q https://launchpad.net/ubuntu/+archive/primary/+sourcefiles/linux/${FULL_VERSION}/linux_${FULL_VERSION}.diff.gz
    wget -q https://launchpad.net/ubuntu/+archive/primary/+sourcefiles/linux/${FULL_VERSION}/linux_${FULL_VERSION}.dsc
    dpkg-source -x linux_${FULL_VERSION}.dsc
    rm *.gz *.dsc
    dirs=( linux-${KERNEL_VERSION%.*}* )
    if [ -d "${dirs[0]}" ]; then
        mv "${dirs[0]}" "${KERNEL_WORK}"
    else
        echo "Error: No directory matching linux-${KERNEL_VERSION%.*} found."
        exit 1
    fi
EOF

WORKDIR "${KERNEL_WORK}"
RUN <<-EOF
    ls -lFah
    chmod a+x debian/rules
    chmod a+x debian/scripts/*
    chmod a+x debian/scripts/misc/*
EOF

WORKDIR /root/project
