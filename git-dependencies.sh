#!/usr/bin/env bash

force=$1

cd ..

if [ "$force" == "-f" ]; then
  rm -rf fabric-starter
  rm -rf nsd-commercial-paper-client
fi

git clone --depth=1 --branch 2018_03-MAIN_ORG_DEPLOYMENT https://github.com/olegabu/fabric-starter
git clone --depth=1 --branch develop https://github.com/olegabu/nsd-commercial-paper-client