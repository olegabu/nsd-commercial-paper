#!/usr/bin/env bash

newOrg=$1
newOrgIp=$2


###########################################################################
# Load chaincode inits
###########################################################################

# load ORGS var
if [[ -f ./env-external-orgs-list ]]; then
  source ./env-external-orgs-list;
else
  ORGS=""
fi
#------------------ init args depend on ORGS -----------------------
INSTRUCTION_INIT_JSON=$(cat ./instruction_init.json |tr -d '\n\r ' | sed 's/"/\\"/g' | envsubst )
: ${INSTRUCTION_INIT:='{"Args":["init","'$INSTRUCTION_INIT_JSON'"]}'}
: ${POSITION_INIT:='{"Args":["init"]}'}

###########################################################################
# Start
###########################################################################
channels="common"

#bilateral
biChannel="${MAIN_ORG}-${newOrg}"
channels="$channels ${biChannel}"
echo "Creating bilateral channel $biChannel, $channels"
network.sh -m create-channel $MAIN_ORG "$biChannel"
#${newOrg}

network.sh -m update-sign-policy -o $THIS_ORG -k "$biChannel"
network.sh -m instantiate-chaincode -o $THIS_ORG -k $biChannel -n position -I "${POSITION_INIT}"

#trilateral
ORGList=($ORGS)
echo "Create trilateral for orgs: $ORGS"

for org in ${ORGList[@]}; do
  if [[ "$org" != "$newOrg" ]]; then
    sortedChannelName=`echo "${org} ${newOrg}" | tr " " "\n" | sort |tr "\n" " " | sed 's/ /-/'`
    echo "Create channel: $sortedChannelName"
    channels="$channels ${sortedChannelName}"
    network.sh -m create-channel $MAIN_ORG "$sortedChannelName"

    network.sh -m update-sign-policy -o $THIS_ORG -k "$sortedChannelName"
    network.sh -m register-org-in-channel $MAIN_ORG "$sortedChannelName" ${org}
    network.sh -m instantiate-chaincode -o $THIS_ORG -k $sortedChannelName -n instruction -I "${INSTRUCTION_INIT}"
  fi
done

echo "***************************************************************************"
echo "Register new org in channels: $channels"
network.sh -m register-new-org -o ${newOrg} -M $MAIN_ORG -i ${newOrgIp} -k "$channels"

export ORGS="$ORGS $newOrg"
echo "export ORGS=\"$ORGS\"" > ./env-external-orgs-list