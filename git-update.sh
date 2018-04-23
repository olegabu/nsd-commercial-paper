#!/usr/bin/env bash

NSD_TAG_SUFFIX=$1

FABRIC_STARTER_TAG=2018_03-MAIN_ORG_DEPLOYMENT
NSD_TAG_PREFIX=2018_03-PRE_RELEASE_
###############################
# git clone NSD related packages
###############################
echo "--------------------------------------------"
echo "----  update nsd-commercial-paper"

git pull
git fetch --tags

if [ -n "$NSD_TAG_SUFFIX" ]; then
  TAG="${NSD_TAG_PREFIX}${NSD_TAG_SUFFIX}"
  echo "Force using tag $TAG"
  git checkout --force $TAG
  #git pull
fi


pushd ..

if [ ! -d "fabric-starter" ]; then
    echo "--------------------------------------------"
    echo "----  Clone fabric-starter"
    git clone --branch $FABRIC_STARTER_TAG https://github.com/olegabu/fabric-starter
fi

echo "--------------------------------------------"
echo "----  update fabric-starter"
cd fabric-starter
git fetch --tags
git checkout $FABRIC_STARTER_TAG --force
#git pull

popd
