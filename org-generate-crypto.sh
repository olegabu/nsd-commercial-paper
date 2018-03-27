#!/usr/bin/env bash

network.sh -m down
docker rm -f $(docker ps -aq)
docker volume rm $(docker volume ls -q -f "name=dockercompose_*")
docker volume prune -f
docker ps -a

###########################################################################
# Start
###########################################################################
network.sh -m removeArtifacts
network.sh -m generate-peer -o $THIS_ORG -R true

