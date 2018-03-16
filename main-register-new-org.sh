#!/usr/bin/env bash

: ${FABRIC_STARTER_HOME:=../..}
source $FABRIC_STARTER_HOME/common.sh $1 $4

newOrg=$2
newOrgIp=$3


###########################################################################
# Start
###########################################################################
channels="common"

#bilateral
biChannel="${MAIN_ORG}-${newOrg}"
channels="$channels ${biChannel}"
network.sh -m create-channel $MAIN_ORG "$biChannel" ${newOrg}

network.sh -m instantiate-chaincode -o $THIS_ORG -k $biChannel -n position -I "${POSITION_INIT}"

#threelateral
for org in ${ORGS}; do
  if [[ "$org" != "$newOrg" ]]; then
    sortedChannelName=`echo "${org} ${newOrg}" | tr " " "\n" | sort |tr "\n" " " | sed 's/ /-/'`
    echo "Create channel: $sortedChannelName"
    channels="$channels ${sortedChannelName}"
    network.sh -m create-channel $MAIN_ORG "$sortedChannelName" ${org} ${newOrg}
    network.sh -m instantiate-chaincode -o $THIS_ORG -k $sortedChannelName -n instruction -I "${INSTRUCTION_INIT}"
  fi
done

network.sh -m register-new-org -o ${newOrg} -M $MAIN_ORG -i ${newOrgIp} -k "$channels"

