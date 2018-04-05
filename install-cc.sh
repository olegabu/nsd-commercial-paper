#!/usr/bin/env bash

[ -n "$1" ] && cc_version=$1 || cc_version="1.0"

sleep 2
if [ "$MAIN_ORG" == "$THIS_ORG" ]; then
    network.sh -m install-chaincode -o $THIS_ORG -v ${cc_version} -n book
fi
sleep 1
network.sh -m install-chaincode -o $THIS_ORG -v ${cc_version} -n security
sleep 1
network.sh -m install-chaincode -o $THIS_ORG -v ${cc_version} -n instruction
sleep 1
network.sh -m install-chaincode -o $THIS_ORG -v ${cc_version} -n position
