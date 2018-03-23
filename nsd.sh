#!/usr/bin/env bash

#make sure on OR32 and ORG3 ./org-generate-crypto.sh is executed before start this script on main org (ORG1)

./main-start-node.sh
./main-register-new-org.sh $ORG2 $IP2
./main-register-new-org.sh $ORG3 $IP3

