#!/usr/bin/env bash

###########################################################################
# Start
###########################################################################
network.sh -m add-org-connectivity -o $THIS_ORG -M $MAIN_ORG -R $MAIN_ORG -i ${IP1}
network.sh -m up-one-org -o $THIS_ORG -M $MAIN_ORG


echo -e $separateLine
echo " >> Joining org $THIS_ORG to channel common"
network.sh -m  join-channel $THIS_ORG $MAIN_ORG common


#join bilateral
biChannel="${MAIN_ORG}-${THIS_ORG}"
echo " >> Joining org $THIS_ORG to channel $biChannel"
network.sh -m  join-channel $THIS_ORG $MAIN_ORG "$biChannel"


./install-cc.sh $1




