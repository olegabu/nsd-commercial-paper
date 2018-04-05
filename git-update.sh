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

echo "--------------------------------------------"
echo "----  update nsd-commercial-paper"
cd ../nsd-commercial-paper
git pull
