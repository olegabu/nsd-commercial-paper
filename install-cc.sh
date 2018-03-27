cc_version="1.0"
sleep 3
network.sh -m install-chaincode -o $THIS_ORG -v ${cc_version} -n security
sleep 1
network.sh -m install-chaincode -o $THIS_ORG -v ${cc_version} -n book
sleep 1
network.sh -m install-chaincode -o $THIS_ORG -v ${cc_version} -n instruction
sleep 1
network.sh -m install-chaincode -o $THIS_ORG -v ${cc_version} -n position
