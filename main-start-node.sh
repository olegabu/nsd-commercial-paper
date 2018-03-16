#!/usr/bin/env bash

: ${FABRIC_STARTER_HOME:=../..}
source $FABRIC_STARTER_HOME/common.sh $1 $2

network.sh -m down
docker rm -f $(docker ps -aq)
docker ps -a

###########################################################################
# Start
###########################################################################
network.sh -m removeArtifacts

echo "THIS_ORG: $THIS_ORG"
network.sh -m generate-peer -o $THIS_ORG -a 4000 -w 8081

network.sh -m generate-orderer -o $THIS_ORG
network.sh -m up-orderer

network.sh -m up-one-org -o $THIS_ORG -M $THIS_ORG -k common
network.sh -m update-sign-policy -o $THIS_ORG -k common

network.sh -m create-channel $MAIN_ORG  "depository"


echo -e $separateLine
echo "Megafon is registered in channel common. Now chaincode 'security, book' will be installed and instantiated "

./install-cc.sh $1 $2

network.sh -m instantiate-chaincode -o $THIS_ORG -k common -n security -I "${SECURITY_INIT}"
network.sh -m instantiate-chaincode -o $THIS_ORG -k depository -n book -I "${BOOK_INIT}"


echo -e $separateLine
echo "Org 'nsd' is up. Channels 'common', 'depository' are created. New organizations may be added by using 'main-register-new-org.sh'"

