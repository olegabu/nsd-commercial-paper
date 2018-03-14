#!/usr/bin/env bash

: ${FABRIC_STARTER_HOME:=../..}
source $FABRIC_STARTER_HOME/common.sh $1 $2

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

echo -e $separateLine
read -n1 -r -p "Org 'nsd' is up and joined to channel 'common'. Now on node 'megafon' generate crypto material (start script ./org-start-node.sh ) and press any key to register 'megafon' channel 'common'"
network.sh -m register-new-org -o megafon -i ${IP2} -k common

echo -e $separateLine
echo "Megafon is registered in channel common. Now chaincode 'chaincode_example02' will be installed and instantiated "
network.sh -m install-chaincode -o $THIS_ORG -v 1.0 -n security
network.sh -m instantiate-chaincode -o $THIS_ORG -k common -n security -I "${SECURITY_INIT}"
