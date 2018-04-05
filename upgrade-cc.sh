# #########################################################################
# This script upgrades chaincodes in all NSD channels with the specified version
###########################################################################
#!/usr/bin/env bash

cc_version=$1

# #########################################################################
# Load chaincode init args
###########################################################################
INSTRUCTION_INIT_JSON=$(cat ./instruction_init.json |tr -d '\n\r ' | sed 's/"/\\"/g' | envsubst )
: ${INSTRUCTION_INIT:='{"Args":["init","'$INSTRUCTION_INIT_JSON'"]}'}

: ${POSITION_INIT:='{"Args":["init"]}'}

BOOK_INIT_JSON=$(cat ./book_init.json |sed 's/"/\\"/g' |tr -d '\n\r ' | envsubst )
: ${BOOK_INIT:='{"Args":["init","'$BOOK_INIT_JSON'"]}'}

SECURITY_INIT_JSON=$(cat ./security_init.json |tr -d '\n\r ' | sed 's/"/\\"/g' | envsubst )
: ${SECURITY_INIT:='{"Args":["init","'$SECURITY_INIT_JSON'"]}'}




# #########################################################################
# Start
###########################################################################


#common channels
network.sh -m upgrade-chaincode -o $THIS_ORG -v ${cc_version} -k common -n security -I $SECURITY_INIT
sleep 1
network.sh -m upgrade-chaincode -o $THIS_ORG -v ${cc_version} -k depository -n book -I $BOOK_INIT
sleep 1


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

ORGList=($ORGS)


#bilateral channel
echo " >> Upgrade bilatral channels for orgs: $ORGS"

for org in ${ORGList[@]}; do
  biChannel="${MAIN_ORG}-${org}"
  network.sh -m upgrade-chaincode -o $THIS_ORG -v ${cc_version} -k "$biChannel" -n position -I "${POSITION_INIT}"
done


#trilateral channels
echo " >> Upgrade trilateral channels for orgs: $ORGS"
subArrayStartIndex=1;
for org in ${ORGList[@]}; do
  pairingOrgs=(${ORGList[@]:subArrayStartIndex})
  echo ""
  echo "Orgs to Pair: ${pairingOrgs[@]}"
  for subOrg in ${pairingOrgs[@]}; do
      if [[ "$org" != "$subOrg" ]]; then
        sortedChannelName=`echo "${org} ${subOrg}" | tr " " "\n" | sort | tr "\n" " " | sed 's/ /-/'`
        echo " >> Upgrade on trilateral channel: $sortedChannelName"
        network.sh -m upgrade-chaincode -o $THIS_ORG  -v ${cc_version} -k $sortedChannelName -n instruction -I "${INSTRUCTION_INIT}"
      fi
  done
  subArrayStartIndex=$((subArrayStartIndex+1))
done