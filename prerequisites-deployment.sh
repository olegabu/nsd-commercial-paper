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

echo "--------------------------------------------"
echo "----  Clone fabric-starter"
git clone --depth=1 --branch 2018_03-MAIN_ORG_DEPLOYMENT https://github.com/olegabu/fabric-starter

echo "--------------------------------------------"
echo "----  Clone nsd-commercial-paper-client"
git clone --depth=1 --branch develop https://github.com/olegabu/nsd-commercial-paper-client

echo "--------------------------------------------"
echo "----  update nsd-commercial-paper-client"
cd nsd-commercial-paper-client
git pull

echo "--------------------------------------------"
echo "----  update fabric-starter"
cd ../fabric-starter
git pull


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