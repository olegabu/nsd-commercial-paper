#!/usr/bin/env bash

#network.sh -m down
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

./install-cc.sh

BOOK_INIT_JSON=$(cat ./book_init.json |sed 's/"/\\"/g' |tr -d '\n\r ' | envsubst )
: ${BOOK_INIT:='{"Args":["init","'$BOOK_INIT_JSON'"]}'}

: ${SECURITY_INIT:='{"Args":["init","RU000A0JVVB5","active","MZ0987654321","22000000000000000"]}'}


network.sh -m instantiate-chaincode -o $THIS_ORG -k common -n security -I "${SECURITY_INIT}"
network.sh -m instantiate-chaincode -o $THIS_ORG -k depository -n book -I "${BOOK_INIT}"


echo -e $separateLine
echo "Org 'nsd' is up. Channels 'common', 'depository' are created. New organizations may be added by using 'main-register-new-org.sh'"

export ORGS=""
echo "export ORGS=\"\"" > ./env-external-orgs-list