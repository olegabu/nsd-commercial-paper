#!/usr/bin/env bash

forceRemoveRepos=$1 # "-f"

###############################
# git clone NSD related packages
###############################
cd ..

if [ "$forceRemoveRepos" == "-f" ]; then
  rm -rf fabric-starter
  rm -rf nsd-commercial-paper-client
fi

git clone --depth=1 --branch 2018_03-MAIN_ORG_DEPLOYMENT https://github.com/olegabu/fabric-starter
git clone --depth=1 --branch develop https://github.com/olegabu/nsd-commercial-paper-client


###############################
# (re)init docker
###############################
export FABRIC_DOCKER_VERSION=docker-ce-18.03.0
[ `command -v apt` ] && echo "Using: init-docker.sh" || echo "Using: init-docker-centos.sh"
[ `command -v apt` ] && ../fabric-starter/init-docker.sh || ../fabric-starter/init-docker-centos.sh

###############################
# (re)init fabric
###############################
export FABRIC_VERSION=1.1.0
../fabric-starter/init-fabric.sh