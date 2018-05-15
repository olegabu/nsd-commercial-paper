#!/usr/bin/env bash

###############################
# git clone\update NSD related packages
###############################
# ./git-update.sh

pushd ../fabric-starter

###############################
# (re)init docker
###############################
echo "--------------------------------------------"
echo "Current directory: $PWD"

if [ `command -v apt` ]; then
    echo "--------------------------------------------"
    echo "Using: init-docker.sh"
    export FABRIC_DOCKER_VERSION=docker-ce=18.03.0~ce-0~ubuntu
    ./init-docker.sh
else
    echo "--------------------------------------------"
    echo "Using: init-docker-centos.sh"
    export FABRIC_DOCKER_VERSION=docker-ce-18.03.0.ce
    ./init-docker-centos.sh
fi


###############################
# (re)init fabric
###############################
export FABRIC_VERSION=1.1.0
echo "--------------------------------------------"
echo "Init fabric: $FABRIC_VERSION"
./init-fabric.sh

popd