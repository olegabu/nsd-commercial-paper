# #########################################################################
# This script updates project source code then upgrades chaincodes
###########################################################################
#!/usr/bin/env bash

cc_version=$1
TAG_SUFFIX=$2

[ -z "$cc_version" ] && echo "Specify new chaincode version in the first argument, e.g: 2.0" && exit 1
[ -z "$TAG_SUFFIX" ] && echo "Specify tag suffix in second argument to update source code by tag, e.g: '02' will produce tag '2018_03-PRE_RELEASE_02'" && exit 1

# #########################################################################
# Update source codes
###########################################################################
#./git-update.sh $TAG_SUFFIX


echo "Upgrade chaincode to version: $cc_version"

# #########################################################################
# Install chaincode
###########################################################################
./install-cc.sh $cc_version

# #########################################################################
# Upgrade
###########################################################################
if [ "$THIS_ORG" == "$MAIN_ORG" ]; then
  ./upgrade-cc.sh $cc_version
fi