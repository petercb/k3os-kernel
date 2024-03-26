#!/bin/bash

docker run -it --rm \
    -e IN_CONTAINER="true" \
    -e TARGETARCH="$(go env GOHOSTARCH)" \
    -v "$(git rev-parse --show-toplevel):/root/project" \
    -w /root/project \
    buildpack-deps:jammy
