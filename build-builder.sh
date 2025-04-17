#!/bin/bash

set -eux pipefail

source envvars

CI_ARGS=""
if [ "${CI:-}" = "true" ]; then
    CI_ARGS="--pull --build-arg BUILDKIT_INLINE_CACHE=1"
fi

export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=plain

echo "Building ${IMAGE_FQN}:${IMAGE_TAG}"
docker build \
    ${CI_ARGS} \
    --tag "${IMAGE_FQN}:${IMAGE_TAG}" \
    --cache-from "${IMAGE_FQN}:${IMAGE_TAG}" \
    --build-arg "KERNEL_VERSION=${KERNEL_VERSION}" \
    --build-arg "UBUNTU_BUILD=${UBUNTU_BUILD}" \
    --build-arg "UBUNTU_FLAVOUR=${UBUNTU_FLAVOUR}" \
    --build-arg "ABI_VERSION=${ABI_VERSION}" \
    .
