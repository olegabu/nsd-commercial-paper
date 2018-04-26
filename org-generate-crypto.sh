#!/usr/bin/env bash

network.sh -m down
docker rm -f $(docker ps -aq)
docker volume rm $(docker volume ls -q -f "name=dockercompose_*")
docker volume prune -f
docker ps -a

###########################################################################
# Start
###########################################################################
if [ "$DEBUG_NOT_REMOVE_OLD_ARTIFACTS" == "" ]; then #sometimes in debug it need not to remove old artifacts
    echo "Removing old artifacts"
    network.sh -m removeArtifacts
fi
network.sh -m generate-peer -o $THIS_ORG -R true

