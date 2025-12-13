#!/bin/bash

set -eu pipefail

source "$(dirname "$0")/envvars"

IMAGE_ID=$(docker image ls -q "${IMAGE_FQN}:${IMAGE_TAG}")
if [ "${IMAGE_ID}" == "" ]
then
    ./build-builder.sh
fi

docker run -it --rm \
    -v "$(git rev-parse --show-toplevel):/root/project" \
    "${IMAGE_FQN}:${IMAGE_TAG}"
