#!/bin/bash

set -eux pipefail

source envvars

CI_ARGS=""
if [ "${CI:-}" = "true" ]; then
    CI_ARGS="--pull --build-arg BUILDKIT_INLINE_CACHE=1"
fi

export DOCKER_BUILDKIT=1

echo "Building ${IMAGE_FQN}:${IMAGE_TAG}"
docker build \
    ${CI_ARGS} \
    --tag "${IMAGE_FQN}:${IMAGE_TAG}" \
    --cache-from "${IMAGE_FQN}:${IMAGE_TAG}" \
    --progress=plain \
    .
