KERNEL_VERSION ?= 5.15.0-89
KERNEL_PATCH ?= 99

ifdef CIRCLE_TAG
VERSION := $(CIRCLE_TAG)
else
VERSION := $(KERNEL_VERSION).$(KERNEL_PATCH)-next$(shell git rev-list $(shell git describe --always --tags --abbrev=0)..HEAD --count)
endif

ARCH ?= $(shell go env GOHOSTARCH)
DIST_DIR := ./dist
BUILD_DIR := ./build
CIRCLE_PROJECT_USERNAME ?= petercb
CIRCLE_PROJECT_REPONAME ?= k3os-kernel
REGISTRY ?= ghcr.io
IMAGE_FQN := ${REGISTRY}/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}


$(DIST_DIR)/kernel-$(ARCH).squashfs: $(BUILD_DIR)/CID
	mkdir -p "${DIST_DIR}"
	docker cp "$(file <$?):/output/kernel.squashfs" "${DIST_DIR}/kernel-${ARCH}.squashfs"
	docker rm -fv "$(file <$?)"
	rm $?


$(BUILD_DIR)/CID: docker-build
	docker create "${IMAGE_FQN}:${VERSION}" > $@


$(BUILD_DIR)/ubuntu-firmware.deb: $(BUILD_DIR)
	wget --no-verbose -cO $@ \
    http://launchpadlibrarian.net/698045658/linux-firmware_20220329.git681281e4-0ubuntu3.23_all.deb


$(BUILD_DIR)/ubuntu-kernel.deb: $(BUILD_DIR)
	wget --no-verbose -cO $@ \
    http://launchpadlibrarian.net/695331190/linux-source-5.15.0_${KERNEL_VERSION}.${KERNEL_PATCH}_all.deb


$(BUILD_DIR):
	mkdir -p $@

.PHONY: docker-build
docker-build: $(BUILD_DIR)/ubuntu-kernel.deb $(BUILD_DIR)/ubuntu-firmware.deb
	@echo "Building k3os kernel ${VERSION}"
	DOCKER_BUILDKIT=1 docker build \
		--tag "${IMAGE_FQN}:${VERSION}" \
		--build-arg BUILDKIT_INLINE_CACHE=1 \
		--build-arg "KERNEL_PATCH=${KERNEL_PATCH}" \
		--build-arg "KERNEL_VERSION=${KERNEL_VERSION}" \
		--target=kernel \
		.


clean:
	rm -rf $(BUILD_DIR) $(DIST_DIR)
