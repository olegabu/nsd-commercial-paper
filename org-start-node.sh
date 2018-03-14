#!/usr/bin/env bash

: ${FABRIC_STARTER_HOME:=../..}
source $FABRIC_STARTER_HOME/common.sh $1 $2

###########################################################################
# Start
###########################################################################
network.sh -m generate-peer -o $THIS_ORG

echo -e $separateLine
read -n1 -r -p "Peer material is generated. Now on node $MAIN_ORG  add org  $THIS_ORG then press any key in this console to start UP org $THIS_ORG ..."
network.sh -m add-org-connectivity -o $THIS_ORG -M $MAIN_ORG -i ${IP1}
network.sh -m up-one-org -o $THIS_ORG -M $MAIN_ORG

echo -e $separateLine
echo "Installing Chaincode 'security' "
network.sh -m install-chaincode -o $THIS_ORG -v 1.0 -n security


echo -e $separateLine
echo "Joining org $THIS_ORG to channel common"
network.sh -m  join-channel $THIS_ORG $MAIN_ORG common





