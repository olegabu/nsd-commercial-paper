#!/usr/bin/env bash

network.sh -m down
docker rm -f $(docker ps -aq)
docker ps -a

###########################################################################
# Start
###########################################################################
network.sh -m removeArtifacts
network.sh -m generate-peer -o $THIS_ORG -R true

