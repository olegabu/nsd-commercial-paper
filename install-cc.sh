
version=$1
network.sh -m install-chaincode -o $THIS_ORG -v ${version} -n security
network.sh -m install-chaincode -o $THIS_ORG -v ${version} -n book
network.sh -m install-chaincode -o $THIS_ORG -v ${version} -n instruction
network.sh -m install-chaincode -o $THIS_ORG -v ${version} -n position



