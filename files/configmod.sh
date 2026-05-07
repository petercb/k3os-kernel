#!/bin/bash
debian/rules clean
if ! debian/rules "${CONFIGMODE:-update}configs"
then
    sed -i \
        -e "/^CONFIG_CC_CAN_LINK/d" \
        -e "/^CONFIG_CC_VERSION_TEXT/d" \
        -e "/^CONFIG_CC_HAS_MARCH_NATIVE/d" \
        -e "/^CONFIG_X86_NATIVE_CPU/d" \
        "debian.${KERNEL_FLAVOUR:-k3os}/config/annotations"
    cp debian.k3os/config/annotations \
        /root/project/overlay/debian.k3os/config/annotations
    exit 1
fi
