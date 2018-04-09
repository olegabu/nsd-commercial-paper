#!/usr/bin/env bash

if [ -n "$1" ]; then
 cc_version=$1
else
  cc_version="1.0"
fi

echo "Install chaincode with version: ${cc_version}"

sleep 3
#if [ "$MAIN_ORG" == "$THIS_ORG" ]; then
    network.sh -m install-chaincode -o $THIS_ORG -v ${cc_version} -n book
#fi
sleep 1
network.sh -m install-chaincode -o $THIS_ORG -v ${cc_version} -n security
sleep 1
network.sh -m install-chaincode -o $THIS_ORG -v ${cc_version} -n instruction
sleep 1
network.sh -m install-chaincode -o $THIS_ORG -v ${cc_version} -n position
