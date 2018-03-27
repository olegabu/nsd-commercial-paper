#!/usr/bin/env bash

externalOrg=$1
externalOrgIp=$2

externalOrgs=("$externalOrg")
###########################################################################
# Start
###########################################################################
#join trilateral channels
for org in ${externalOrgs[@]}; do
  if [[ "$org" != "$THIS_ORG" ]]; then
    sortedChannelName=`echo "${org} ${THIS_ORG}" | tr " " "\n" | sort |tr "\n" " " | sed 's/ /-/'`
    echo " >> Join channel: $sortedChannelName"
    network.sh -m  join-channel $THIS_ORG $MAIN_ORG "$sortedChannelName"
  fi
done

echo
echo "Download instruction_init.json"
f="dockercompose/docker-compose-$THIS_ORG.yaml"
GID=$(id -g)
c="wget http://www.${MAIN_ORG}.$DOMAIN:8080/instruction_init.json && chown -R $UID:$GID ."
docker-compose --file ${f} run --rm "cli.$THIS_ORG.$DOMAIN" bash -c "${c}"

mv -f artifacts/instruction_init.json ./instruction_init.json

#add IP to api's hosts
network.sh -m add-org-connectivity -o $THIS_ORG -M $MAIN_ORG -R $externalOrg -i $externalOrgIp

