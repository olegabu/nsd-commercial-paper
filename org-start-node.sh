#!/usr/bin/env bash

: ${FABRIC_STARTER_HOME:=../..}
source $FABRIC_STARTER_HOME/common.sh $1 $2

network.sh -m down
docker rm -f $(docker ps -aq)
docker ps -a

###########################################################################
# Start
###########################################################################
network.sh -m generate-peer -o $THIS_ORG

echo -e $separateLine
read -n1 -r -p "Peer material is generated. Now on node $MAIN_ORG  add org  $THIS_ORG then press any key in this console to start UP org $THIS_ORG ..."
network.sh -m add-org-connectivity -o $THIS_ORG -M $MAIN_ORG -i ${IP1}
network.sh -m up-one-org -o $THIS_ORG -M $MAIN_ORG


echo -e $separateLine
echo "Joining org $THIS_ORG to channel common"
network.sh -m  join-channel $THIS_ORG $MAIN_ORG common


#join bilateral
biChannel="${MAIN_ORG}-${THIS_ORG}"
network.sh -m  join-channel $THIS_ORG $MAIN_ORG "$biChannel"

#join threelateral
for org in ${ORGS}; do
  if [[ "$org" != "$THIS_ORG" ]]; then
    sortedChannelName=`echo "${org} ${THIS_ORG}" | tr " " "\n" | sort |tr "\n" " " | sed 's/ /-/'`
    echo "Join channel: $sortedChannelName"
    network.sh -m  join-channel $THIS_ORG $MAIN_ORG "$sortedChannelName"
  fi
done


./install-cc.sh $1 $2

network.sh -m add-org-connectivity -o $THIS_ORG -M $MAIN_ORG -i $IP1
network.sh -m restart-api -o $THIS_ORG


