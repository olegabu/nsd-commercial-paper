#!/usr/bin/env bash

: ${FABRIC_STARTER_HOME:=../..}
source $FABRIC_STARTER_HOME/common.sh $1 $6

newOrg=$2
newOrgIp=$3
channel=$4
chaincode=$5


###########################################################################
# Start
###########################################################################

echo "Not Implemented" && exit 1;

#install cc

for channel in ${@:4}; do
  echo "Create channel $channel"
  network.sh -m create-channel $MAIN_ORG "$channel" ${newOrg}
#    instantiateWarmUp instruction ${CHANNEL_NAME} ${INSTRUCTION_INIT}
#    instantiateWarmUp position ${CHANNEL_NAME} ${POSITION_INIT}
done

network.sh -m register-new-org -o ${newOrg} -M $MAIN_ORG -i ${newOrgIp} -k common

