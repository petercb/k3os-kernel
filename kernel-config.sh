#!/bin/bash

set -euo pipefail

source "$(dirname "$0")/envvars"

export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=plain

echo "Building ${IMAGE_FQN}-configmod:${IMAGE_TAG}"
docker build \
    --platform "linux/${TARGETARCH}" \
    --tag "${IMAGE_FQN}-configmod:${IMAGE_TAG}" \
    --cache-from "${IMAGE_FQN}:${IMAGE_TAG},${IMAGE_FQN}-configmod:${IMAGE_TAG}" \
    --build-arg "KERNEL_VERSION=${KERNEL_VERSION}" \
    --build-arg "UBUNTU_BUILD=${UBUNTU_BUILD}" \
    --build-arg "UBUNTU_FLAVOUR=${UBUNTU_FLAVOUR}" \
    --build-arg "UBUNTU_NAME=${UBUNTU_NAME}" \
    --build-arg "ABI_VERSION=${ABI_VERSION}" \
    --target=configmod \
    .

docker run -it --rm \
    --platform "linux/${TARGETARCH}" \
    --env CONFIGMODE="${1:-edit}" \
    -v "$(git rev-parse --show-toplevel):/root/project" \
    "${IMAGE_FQN}-configmod:${IMAGE_TAG}"
