#!/usr/bin/env bash

newOrg=$1
newOrgIp=$2


###########################################################################
# Load list of existing ORGS
###########################################################################
# ORGS variable
if [[ -f ./env-external-orgs-list ]]; then
  source ./env-external-orgs-list;
else
  ORGS=""
fi
ORGS=`echo ${ORGS} | tr -d '\r\n'` #remove windows-EOL

# #########################################################################
# Load chaincode init args
###########################################################################
INSTRUCTION_INIT_JSON=$(cat ./instruction_init.json |tr -d '\n\r ' | sed 's/"/\\"/g' | envsubst )
: ${INSTRUCTION_INIT:='{"Args":["init","'$INSTRUCTION_INIT_JSON'"]}'}
: ${POSITION_INIT:='{"Args":["init"]}'}

cp -f instruction_init.json www/artifacts/
###########################################################################
# Start
###########################################################################

#bilateral
biChannel="${MAIN_ORG}-${newOrg}"
echo " >> Creating bilateral channel $biChannel, $channels"
network.sh -m create-channel $MAIN_ORG "$biChannel"
network.sh -m update-sign-policy -o $THIS_ORG -k "$biChannel"
#network.sh -m instantiate-chaincode -o $THIS_ORG -k $biChannel -n position -I "${POSITION_INIT}"

#trilateral
ORGList=($ORGS)
echo " >> Create trilateral for orgs: $ORGS"

trilateralChannels=""
for org in ${ORGList[@]}; do
  if [[ "$org" != "$newOrg" ]]; then
    sortedChannelName=`echo "${org} ${newOrg}" | tr " " "\n" | sort |tr "\n" " " | sed 's/ /-/'`
    echo " >> Create channel: $sortedChannelName"
    network.sh -m create-channel $MAIN_ORG "$sortedChannelName"
    network.sh -m update-sign-policy -o $THIS_ORG -k "$sortedChannelName"
    network.sh -m register-org-in-channel $MAIN_ORG "$sortedChannelName" ${org}
#    network.sh -m instantiate-chaincode -o $THIS_ORG -k $sortedChannelName -n instruction -I "${INSTRUCTION_INIT}"

    #track trilateral channels list
    trilateralChannels="$trilateralChannels ${sortedChannelName}"
  fi
done

echo "***************************************************************************"
echo " >> Register new org in channels: $channels"
network.sh -m register-new-org -o ${newOrg} -M $MAIN_ORG -i ${newOrgIp} -k "common $biChannel $trilateralChannels"

echo " >> Instantiate chaincode position in bilateral channel: $biChannel"
network.sh -m instantiate-chaincode -o $THIS_ORG -k "$biChannel" -n position -I "${POSITION_INIT}"
network.sh -m warmup-chaincode -o $THIS_ORG -k "$biChannel" -n position -I '{"Args":["query",""]}'

if [ -n "$trilateralChannels" ]; then
  echo " >> Instantiate chaincode instruction in trilateral channels: $trilateralChannels"
  network.sh -m instantiate-chaincode -o $THIS_ORG -k "$trilateralChannels" -n instruction -I "${INSTRUCTION_INIT}"
  network.sh -m warmup-chaincode -o $THIS_ORG -k "$trilateralChannels" -n instruction -I '{"Args":["query",""]}'
fi

export ORGS="$ORGS $newOrg"
echo "export ORGS=\"$ORGS\"" > ./env-external-orgs-list