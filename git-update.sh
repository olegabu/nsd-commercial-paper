#!/usr/bin/env bash

TAG_SUFFIX=$1

###############################
# git clone NSD related packages
###############################
echo "--------------------------------------------"
echo "----  update nsd-commercial-paper"

git pull
if [ -n "$TAG_SUFFIX" ]; then
  TAG="2018_03-PRE_RELEASE_${TAG_SUFFIX}"
  echo "Force using tag $TAG"
  git checkout --force $TAG
  git pull
fi


cd ..

echo "--------------------------------------------"
echo "----  Clone fabric-starter"
git clone --depth=1 --branch 2018_03-MAIN_ORG_DEPLOYMENT https://github.com/olegabu/fabric-starter

echo "--------------------------------------------"
echo "----  Clone nsd-commercial-paper-client"
git clone --depth=1 --branch develop https://github.com/Altoros/nsd-commercial-paper-client


###############################
# Update repos
###############################
echo "--------------------------------------------"
echo "----  update nsd-commercial-paper-client"
cd nsd-commercial-paper-client
git pull

echo "--------------------------------------------"
echo "----  update fabric-starter"
cd ../fabric-starter
git pull

cd ../nsd-commercial-paper