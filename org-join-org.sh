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
    echo "Join channel: $sortedChannelName"
    network.sh -m  join-channel $THIS_ORG $MAIN_ORG "$sortedChannelName"
  fi
done

#add IP to api's hosts
network.sh -m add-org-connectivity -o $THIS_ORG -M $externalOrg -i $externalOrgIp

