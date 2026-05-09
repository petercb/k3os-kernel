# syntax=docker/dockerfile:1
# hadolint global ignore=DL3008

ARG UBUNTU_NAME=resolute
ARG KERNEL_VERSION="7.0.0"
ARG UBUNTU_BUILD="14"
ARG UBUNTU_FLAVOUR="generic"
ARG ABI_VERSION="14"
ARG KERNEL_FULL_VERSION="${KERNEL_VERSION}-${UBUNTU_BUILD}.${ABI_VERSION}"
ARG KERNEL_FLAVOUR="k3os"
ARG KVER="${KERNEL_VERSION}-${UBUNTU_BUILD}-${KERNEL_FLAVOUR}"
ARG KERNEL_WORK="/usr/src/linux"

############################################################
FROM ubuntu:${UBUNTU_NAME} AS base
############################################################

ARG TARGETARCH
ARG UBUNTU_NAME
ARG KERNEL_VERSION
ARG UBUNTU_BUILD
ARG UBUNTU_FLAVOUR
ARG ABI_VERSION
ARG KERNEL_FULL_VERSION
ARG KERNEL_FLAVOUR
ARG KVER
ARG KERNEL_WORK

ENV TARGETARCH=${TARGETARCH}
ENV KERNEL_FULL_VERSION=${KERNEL_FULL_VERSION}
ENV KERNEL_VERSION=${KERNEL_VERSION}
ENV KERNEL_FLAVOUR=${KERNEL_FLAVOUR}
ENV KVER=${KVER}
ENV KERNEL_WORK=${KERNEL_WORK}

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Etc/UTC

SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    <<-EOF
    rm -f /etc/apt/apt.conf.d/docker-clean
    echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache
    sed -i 's/^Types:.*$/Types: deb deb-src/' /etc/apt/sources.list.d/ubuntu.sources
    apt-get update
    apt-get install -y --no-install-recommends \
        ca-certificates
EOF


############################################################
FROM base AS fw-selector-builder
############################################################

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get install -y --no-install-recommends golang-go

WORKDIR /src
COPY fw-selector/ .
RUN go build -o /bin/fw-selector .


############################################################
FROM base AS buildpack
############################################################

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    <<-EOF
    apt-get build-dep -y --no-install-recommends linux
    apt-get install -y --no-install-recommends \
        dwarves \
        libncurses-dev \
        llvm
EOF


############################################################
FROM base AS source
############################################################

WORKDIR "${KERNEL_WORK}"
# hadolint ignore=DL3020
ADD --link \
    "git://git.launchpad.net/~ubuntu-kernel/ubuntu/+source/linux/+git/${UBUNTU_NAME}#Ubuntu-${KERNEL_FULL_VERSION}" \
    .


############################################################
FROM buildpack AS builder
############################################################

WORKDIR "${KERNEL_WORK}"
COPY --from=source "${KERNEL_WORK}" .
COPY /overlay .

WORKDIR "${KERNEL_WORK}/debian.${KERNEL_FLAVOUR}"
RUN <<-EOF
    cp -a ../debian.master/changelog .
    cp -a ../debian.master/control.stub.in .
    cp -a ../debian.master/control.d/*.stub control.d/
    cp -a ../debian.master/control.stub.in .
    cp -a ../debian.master/reconstruct .
EOF


############################################################
FROM builder AS configmod
############################################################

ENV CONFIGMODE=update

COPY --chmod=+x files/configmod.sh /configmod.sh

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    <<-EOF
    apt-get install -y --no-install-recommends \
        gcc-aarch64-linux-gnu \
        gcc-x86-64-linux-gnu
EOF

WORKDIR "${KERNEL_WORK}"

CMD ["/configmod.sh"]


############################################################
FROM builder AS compile
############################################################

WORKDIR "${KERNEL_WORK}"
RUN <<-EOF
    debian/rules clean
    # see https://wiki.ubuntu.com/KernelTeam/KernelMaintenance#Overriding_module_check_failures
    debian/rules "binary-${KERNEL_FLAVOUR}" \
        skipabi=true \
        skipmodule=true \
        skipretpoline=true \
        skipdbg=true
    dpkg --unpack --no-triggers --force-depends \
        "../linux-image-unsigned-${KVER}_${KERNEL_FULL_VERSION}_${TARGETARCH}.deb" \
        "../linux-modules-${KVER}_${KERNEL_FULL_VERSION}_${TARGETARCH}.deb"
    depmod "${KVER}"
    rm ../linux-*.deb
    debian/rules clean
EOF

WORKDIR /boot
RUN <<-EOF
    mv "System.map-${KVER}" System.map
    mv "config-${KVER}" config
    mv "vmlinuz-${KVER}" vmlinuz
    echo "${KVER}" > kversion
EOF

COPY --from=fw-selector-builder /bin/fw-selector /usr/local/bin/fw-selector
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    <<-EOF
    apt-get install -y --no-install-recommends curl
    # Fetch WHENCE file from kernel.org. If it fails, touch an empty file to fallback to extracting everything.
    curl -sfL -o /tmp/WHENCE https://git.kernel.org/pub/scm/linux/kernel/git/firmware/linux-firmware.git/plain/WHENCE || touch /tmp/WHENCE
    fw-selector --config /boot/config --source-dir "${KERNEL_WORK}" --whence /tmp/WHENCE --arch "${TARGETARCH}" --output /boot/firmware-list.txt
    rm /tmp/WHENCE
EOF


############################################################
FROM base AS test
############################################################

ENV GOPATH=/root/go
ENV PATH=${GOPATH}/bin:${PATH}
ENV INITRD="/tmp/test-initrd.cpio"
ENV KERNEL="/tmp/vmlinuz"
ENV LOG_FILE="/tmp/qemu.log"

# hadolint ignore=SC3054
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    <<-EOF
    #!/bin/bash
    PKGS=(
        cpio golang-go
    )
    case "${TARGETARCH}" in
        amd64) PKGS+=(qemu-system-x86) ;;
        arm64) PKGS+=(qemu-system-arm ipxe-qemu qemu-efi-aarch64) ;;
        *) echo "Unknown architecture: ${TARGETARCH}"; exit 1 ;;
    esac
    echo "Installing packages: ${PKGS[*]}"
    apt-get install -y --no-install-recommends "${PKGS[@]}"
EOF

WORKDIR /tmp/initrd/u-root
ADD --link https://github.com/u-root/u-root.git#v0.16.0 .
RUN go install

WORKDIR /tmp/initrd
COPY --parents u-root-init ./
COPY --from=compile /lib/modules ./lib/
RUN <<-EOF
    go work init ./u-root
    go work use ./u-root-init
    u-root -o "${INITRD}" \
        -build=binary \
        -defaultsh="" \
        -initcmd test-init \
        -files ./lib \
        test-init
    cpio -ivt < "${INITRD}"
EOF

COPY --from=compile --chmod=644 "/boot/vmlinuz" "${KERNEL}"

COPY --chmod=+x files/test_kernel.sh /bin/test_kernel.sh
WORKDIR /output
RUN /bin/test_kernel.sh


############################################################
FROM base AS output
############################################################

LABEL org.opencontainers.image.source = "https://github.com/petercb/k3os-kernel"

COPY --from=compile --parents /boot /
COPY --from=compile --parents /usr/lib/modules /
COPY --from=test /output/results.xml /output/
