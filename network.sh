#!/usr/bin/env bash

DOMAIN=nsd.ru
ORG1=nsd
ORG2=a
ORG3=b
ORG4=c

CLI_TIMEOUT=10000
COMPOSE_FILE=ledger/docker-compose.yaml
COMPOSE_TEMPLATE=ledger/docker-compose-template.yaml
COMPOSE_FILE_DEV=ledger/docker-compose-dev.yaml

# Delete any images that were generated as a part of this setup
# specifically the following images are often left behind:
# TODO list generated image naming patterns
function removeUnwantedImages() {
  DOCKER_IMAGE_IDS=$(docker images | grep "dev\|none\|test-vp\|peer[0-9]-" | awk '{print $3}')
  if [ -z "$DOCKER_IMAGE_IDS" -o "$DOCKER_IMAGE_IDS" == " " ]; then
    echo "No images available for deletion"
  else
    echo "Removing docker images: $DOCKER_IMAGE_IDS"
    docker rmi -f ${DOCKER_IMAGE_IDS}
  fi
}

function removeArtifacts() {
  echo "Removing generated artifacts"
  rm ledger/docker-compose.yaml
  rm -rf artifacts/crypto-config
  rm -rf artifacts/channel
}

function removeDockersFromCompose() {
  if [ -f ${COMPOSE_FILE} ]; then
    echo "Removing docker containers listed in $COMPOSE_FILE"
    docker-compose -f ${COMPOSE_FILE} kill
    docker-compose -f ${COMPOSE_FILE} rm -f
  else
    echo "No generated $COMPOSE_FILE and no docker instances to remove"
  fi;
}

function removeDockersWithDomain() {
  SEARCH="$DOMAIN"
  DOCKER_IDS=$(docker ps -a | grep ${SEARCH} | awk '{print $1}')
  if [ -z "$DOCKER_IDS" -o "$DOCKER_IDS" == " " ]; then
    echo "No docker instances available for deletion with $SEARCH"
  else
    echo "Removing docker instances found with $SEARCH: $DOCKER_IDS"
    docker rm -f ${DOCKER_IDS}
  fi
}

function generateArtifacts() {
    echo "Creating yaml files with $DOMAIN, $ORG1, $ORG2, $ORG3, $ORG4"
    # configtx and cryptogen
    sed -e "s/DOMAIN/$DOMAIN/g" -e "s/DOMAIN/$DOMAIN/g" -e "s/ORG1/$ORG1/g" -e "s/ORG2/$ORG2/g" -e "s/ORG3/$ORG3/g" -e "s/ORG4/$ORG4/g" artifacts/configtxtemplate.yaml > artifacts/configtx.yaml
    sed -e "s/DOMAIN/$DOMAIN/g" artifacts/cryptogentemplate-orderer.yaml > artifacts/"cryptogen-$DOMAIN.yaml"
    sed -e "s/DOMAIN/$DOMAIN/g" -e "s/ORG/$ORG1/g" artifacts/cryptogentemplate-peer.yaml > artifacts/"cryptogen-$ORG1.yaml"
    sed -e "s/DOMAIN/$DOMAIN/g" -e "s/ORG/$ORG2/g" artifacts/cryptogentemplate-peer.yaml > artifacts/"cryptogen-$ORG2.yaml"
    sed -e "s/DOMAIN/$DOMAIN/g" -e "s/ORG/$ORG3/g" artifacts/cryptogentemplate-peer.yaml > artifacts/"cryptogen-$ORG3.yaml"
    sed -e "s/DOMAIN/$DOMAIN/g" -e "s/ORG/$ORG4/g" artifacts/cryptogentemplate-peer.yaml > artifacts/"cryptogen-$ORG4.yaml"
    # docker-compose.yaml
    sed -e "s/DOMAIN/$DOMAIN/g" -e "s/ORG1/$ORG1/g" -e "s/ORG2/$ORG2/g" -e "s/ORG3/$ORG3/g" -e "s/ORG4/$ORG4/g" ${COMPOSE_TEMPLATE} > ${COMPOSE_FILE}
    # network-config.json
    #  fill environments                                                                                            |   remove comments
    sed -e "s/\DOMAIN/$DOMAIN/g" -e "s/\ORG1/$ORG1/g" -e "s/\ORG2/$ORG2/g" -e "s/\ORG3/$ORG3/g" -e "s/\ORG4/$ORG4/g" -e "s/^\s*\/\/.*$//g" artifacts/network-config-template.json > artifacts/network-config.json
    # fabric-ca-server-config.yaml
    sed -e "s/ORG/$ORG1/g" artifacts/fabric-ca-server-configtemplate.yaml > artifacts/"fabric-ca-server-config-$ORG1.yaml"
    sed -e "s/ORG/$ORG2/g" artifacts/fabric-ca-server-configtemplate.yaml > artifacts/"fabric-ca-server-config-$ORG2.yaml"
    sed -e "s/ORG/$ORG3/g" artifacts/fabric-ca-server-configtemplate.yaml > artifacts/"fabric-ca-server-config-$ORG3.yaml"
    sed -e "s/ORG/$ORG4/g" artifacts/fabric-ca-server-configtemplate.yaml > artifacts/"fabric-ca-server-config-$ORG4.yaml"

    echo "Generating crypto material with cryptogen"
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "cryptogen generate --config=cryptogen-$DOMAIN.yaml"
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "cryptogen generate --config=cryptogen-$ORG1.yaml"
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "cryptogen generate --config=cryptogen-$ORG2.yaml"
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "cryptogen generate --config=cryptogen-$ORG3.yaml"
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "cryptogen generate --config=cryptogen-$ORG4.yaml"

    echo "Change cryptomaterial ownership"
    GID=$(id -g)
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "chown -R $UID:$GID ."

    echo "Generating orderer genesis block with configtxgen"
    mkdir -p artifacts/channel
    docker-compose --file ${COMPOSE_FILE} run --rm -e FABRIC_CFG_PATH=/etc/hyperledger/artifacts "cli.$DOMAIN" configtxgen -profile OrdererGenesis -outputBlock ./channel/genesis.block

    CHANNEL_NAME=depository
    echo "Generating channel config transaction for $CHANNEL_NAME"
    docker-compose --file ${COMPOSE_FILE} run --rm -e FABRIC_CFG_PATH=/etc/hyperledger/artifacts "cli.$DOMAIN" configtxgen -profile "$CHANNEL_NAME" -outputCreateChannelTx ./channel/"$CHANNEL_NAME".tx -channelID "$CHANNEL_NAME"

    for CHANNEL_NAME in "$ORG2-$ORG3" "$ORG2-$ORG4" "$ORG3-$ORG4"
    do
        echo "Generating channel config transaction for $CHANNEL_NAME"
        docker-compose --file ${COMPOSE_FILE} run --rm -e FABRIC_CFG_PATH=/etc/hyperledger/artifacts "cli.$DOMAIN" configtxgen -profile "$CHANNEL_NAME" -outputCreateChannelTx ./channel/"$CHANNEL_NAME".tx -channelID "$CHANNEL_NAME"
    done

    echo "Adding generated CA private keys filenames to yaml"
    CA1_PRIVATE_KEY=$(basename `ls artifacts/crypto-config/peerOrganizations/"$ORG1.$DOMAIN"/ca/*_sk`)
    CA2_PRIVATE_KEY=$(basename `ls artifacts/crypto-config/peerOrganizations/"$ORG2.$DOMAIN"/ca/*_sk`)
    CA3_PRIVATE_KEY=$(basename `ls artifacts/crypto-config/peerOrganizations/"$ORG3.$DOMAIN"/ca/*_sk`)
    CA4_PRIVATE_KEY=$(basename `ls artifacts/crypto-config/peerOrganizations/"$ORG4.$DOMAIN"/ca/*_sk`)
    [[ -z  ${CA1_PRIVATE_KEY}  ]] && echo "empty CA1 private key" && exit 1
    [[ -z  ${CA2_PRIVATE_KEY}  ]] && echo "empty CA2 private key" && exit 1
    [[ -z  ${CA3_PRIVATE_KEY}  ]] && echo "empty CA3 private key" && exit 1
    [[ -z  ${CA4_PRIVATE_KEY}  ]] && echo "empty CA4 private key" && exit 1
    sed -i -e "s/CA1_PRIVATE_KEY/${CA1_PRIVATE_KEY}/g" -e "s/CA2_PRIVATE_KEY/${CA2_PRIVATE_KEY}/g" -e "s/CA3_PRIVATE_KEY/${CA3_PRIVATE_KEY}/g" -e "s/CA4_PRIVATE_KEY/${CA4_PRIVATE_KEY}/g" ${COMPOSE_FILE}
}

function createChannel () {
    CHANNEL_NAME=$1
    info "creating channel $CHANNEL_NAME by $ORG1"

    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$ORG1.$DOMAIN" bash -c "peer channel create -o orderer.$DOMAIN:7050 -c $CHANNEL_NAME -f /etc/hyperledger/artifacts/channel/$CHANNEL_NAME.tx --tls --cafile /etc/hyperledger/crypto/orderer/tls/ca.crt"
}

function joinChannel() {
    ORG=$1
    CHANNEL_NAME=$2
    info "joining channel $CHANNEL_NAME by $ORG"

    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$ORG.$DOMAIN" bash -c "CORE_PEER_ADDRESS=peer0.$ORG.$DOMAIN:7051 peer channel join -b $CHANNEL_NAME.block      && CORE_PEER_ADDRESS=peer1.$ORG.$DOMAIN:7051 peer channel join -b $CHANNEL_NAME.block"
}

function instantiateChaincode () {
    ORG=$1
    CHANNEL_NAME=$2
    N=$3
    I=$4
    info "instantiating chaincode $N on $CHANNEL_NAME by $ORG with $I"

    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$ORG.$DOMAIN" bash -c "CORE_PEER_ADDRESS=peer0.$ORG.$DOMAIN:7051 peer chaincode instantiate -n $N -v 1.0 -c '$I' -o orderer.$DOMAIN:7050 -C $CHANNEL_NAME --tls --cafile /etc/hyperledger/crypto/orderer/tls/ca.crt"
}

function warmUpChaincode () {
    ORG=$1
    CHANNEL_NAME=$2
    N=$3
    info "warming up chaincode $N on $CHANNEL_NAME on all peers of $ORG with query"

    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$ORG.$DOMAIN" bash -c "CORE_PEER_ADDRESS=peer0.$ORG.$DOMAIN:7051 peer chaincode query -n $N -v 1.0 -c '{\"Args\":[\"query\"]}' -C $CHANNEL_NAME"
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$ORG.$DOMAIN" bash -c "CORE_PEER_ADDRESS=peer1.$ORG.$DOMAIN:7051 peer chaincode query -n $N -v 1.0 -c '{\"Args\":[\"query\"]}' -C $CHANNEL_NAME"
}

function installChaincode() {
    ORG=$1
    N=$2
    P=$3
    info "installing chaincode $N to peers of $ORG from $P"

    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$ORG.$DOMAIN" bash -c "CORE_PEER_ADDRESS=peer0.$ORG.$DOMAIN:7051 peer chaincode install -n $N -v 1.0 -p $P      && CORE_PEER_ADDRESS=peer1.$ORG.$DOMAIN:7051 peer chaincode install -n $N -v 1.0 -p $P"
}

function networkUp () {
  # generate artifacts if they don't exist
  if [ ! -d "artifacts/crypto-config" ]; then
    generateArtifacts
  fi

  TIMEOUT=${CLI_TIMEOUT} docker-compose -f ${COMPOSE_FILE} up -d 2>&1
  if [ $? -ne 0 ]; then
    echo "ERROR !!!! Unable to start network"
    logs
    exit 1
  fi

  createChannel depository
  joinChannel ${ORG1} depository

  installChaincode ${ORG1} book book
  instantiateChaincode ${ORG1} depository book '{"Args":["init","aEmissionAccount","aActiveDivision","RU000ABC0001","1000"]}'
  warmUpChaincode ${ORG1} depository book

  installChaincode ${ORG1} security security
  instantiateChaincode ${ORG1} depository security '{"Args":["init","RU000ABC0001","active"]}'
  warmUpChaincode ${ORG1} depository security

  createChannel "$ORG2-$ORG3"
  joinChannel ${ORG1} "$ORG2-$ORG3"
  joinChannel ${ORG2} "$ORG2-$ORG3"
  joinChannel ${ORG3} "$ORG2-$ORG3"

  createChannel "$ORG2-$ORG4"
  joinChannel ${ORG1} "$ORG2-$ORG4"
  joinChannel ${ORG2} "$ORG2-$ORG4"
  joinChannel ${ORG4} "$ORG2-$ORG4"

  createChannel "$ORG3-$ORG4"
  joinChannel ${ORG1} "$ORG3-$ORG4"
  joinChannel ${ORG3} "$ORG3-$ORG4"
  joinChannel ${ORG4} "$ORG3-$ORG4"

  CHAINCODE_NAME=instruction
  CHAINCODE_PATH=instruction
  CHAINCODE_INIT='{"Args":["init","depository"]}'

  installChaincode ${ORG1} ${CHAINCODE_NAME} ${CHAINCODE_PATH}
  installChaincode ${ORG2} ${CHAINCODE_NAME} ${CHAINCODE_PATH}
  installChaincode ${ORG3} ${CHAINCODE_NAME} ${CHAINCODE_PATH}
  installChaincode ${ORG4} ${CHAINCODE_NAME} ${CHAINCODE_PATH}

  instantiateChaincode ${ORG1} "$ORG2-$ORG3" ${CHAINCODE_NAME} ${CHAINCODE_INIT}
  instantiateChaincode ${ORG1} "$ORG2-$ORG4" ${CHAINCODE_NAME} ${CHAINCODE_INIT}
  instantiateChaincode ${ORG1} "$ORG3-$ORG4" ${CHAINCODE_NAME} ${CHAINCODE_INIT}

  warmUpChaincode ${ORG1} "$ORG2-$ORG3" ${CHAINCODE_NAME}
  warmUpChaincode ${ORG2} "$ORG2-$ORG3" ${CHAINCODE_NAME}
  warmUpChaincode ${ORG3} "$ORG2-$ORG3" ${CHAINCODE_NAME}

  warmUpChaincode ${ORG1} "$ORG2-$ORG4" ${CHAINCODE_NAME}
  warmUpChaincode ${ORG2} "$ORG2-$ORG4" ${CHAINCODE_NAME}
  warmUpChaincode ${ORG4} "$ORG2-$ORG4" ${CHAINCODE_NAME}

  warmUpChaincode ${ORG1} "$ORG3-$ORG4" ${CHAINCODE_NAME}
  warmUpChaincode ${ORG3} "$ORG3-$ORG4" ${CHAINCODE_NAME}
  warmUpChaincode ${ORG4} "$ORG3-$ORG4" ${CHAINCODE_NAME}

  #logs
}

function devNetworkUp () {
  docker-compose -f ${COMPOSE_FILE_DEV} up -d 2>&1
  if [ $? -ne 0 ]; then
    echo "ERROR !!!! Unable to start network"
    logs
    exit 1
  fi
}

function devNetworkDown () {
  docker-compose -f ${COMPOSE_FILE_DEV} down
}

function devInstallInstantiate () {
 docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode install -p book -n book -v 0 && peer chaincode instantiate -n book -v 0 -C myc -c '{\"Args\":[\"init\",\"aEmissionAccount\",\"aActiveDivision\",\"RU000ABC0001\",\"1000\"]}'"
 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode instantiate -n book -v 0 -C myc -c '{\"Args\":[\"init\",\"aEmissionAccount\",\"aActiveDivision\",\"RU000ABC0001\",\"1000\"]}'"

 docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode install -p instruction -n instruction -v 0 && peer chaincode instantiate -n instruction -v 0 -C myc -c '{\"Args\":[\"myc\",\"myc\"]}'"
 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode instantiate -n instruction -v 0 -C myc -c '{\"Args\":[\"init\",\"myc\"]}'"
 
 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode install -p security -n security -v 0 && peer chaincode instantiate -n security -v 0 -C myc -c '{\"Args\":[\"init\",\"RU000ABC0001\",\"active\"]}'"
 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode instantiate -n security -v 0 -C myc -c '{\"Args\":[\"init\",\"RU000ABC0001\",\"active\"]}'"
}

function devInvoke () {
 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode invoke -n book -v 0 -C myc -c '{\"Args\":[\"move\",\"aEmissionAccount\",\"aActiveDivision\",\"bInvestmentAccount\",\"bActiveDivision\",\"RU000ABC0001\",\"10\"]}'"

 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode invoke -n instruction -v 0 -C myc -c '{\"Args\":[\"receive\",\"aDeponent\",\"aEmissionAccount\",\"aActiveDivision\",\"bInvestmentAccount\",\"bActiveDivision\",\"RU000ABC0001\",\"10\",\"reference1000\",\"2017-08-08\",\"2017-08-07\",\"reason\"]}'"
 docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode invoke -n instruction -v 0 -C myc -c '{\"Args\":[\"check\",\"aEmissionAccount\",\"aActiveDivision\",\"RU000ABC0001\",\"10\"]}'"

 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode invoke -n security -v 0 -C myc -c '{\"Args\":[\"put\",\"RU000ABC0001\",\"redeemed\"]}'"
}

function devQuery () {
 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode query -n book -v 0 -C myc -c '{\"Args\":[\"query\"]}'"
 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode query -n book -v 0 -C myc -c '{\"Args\":[\"history\",\"aEmissionAccount\",\"aActiveDivision\",\"RU000ABC0001\"]}'"

 docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode query -n instruction -v 0 -C myc -c '{\"Args\":[\"query\"]}'"

 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode query -n security -v 0 -C myc -c '{\"Args\":[\"query\"]}'"
}

function info() {
    #figlet $1
    echo "*************************************************************************************************************"
    echo "$1"
    echo "*************************************************************************************************************"
    sleep 2
}

function logs () {
    TIMEOUT=${CLI_TIMEOUT} COMPOSE_HTTP_TIMEOUT=${CLI_TIMEOUT} docker-compose -f ${COMPOSE_FILE} logs -f
}

function devLogs () {
    TIMEOUT=${CLI_TIMEOUT} COMPOSE_HTTP_TIMEOUT=${CLI_TIMEOUT} docker-compose -f ${COMPOSE_FILE_DEV} logs -f
}

function networkDown () {
  docker-compose -f ${COMPOSE_FILE} down
  # Don't remove containers, images, etc if restarting
  if [ "$MODE" != "restart" ]; then
    clean
  fi
}

function clean() {
  removeDockersFromCompose
  removeDockersWithDomain
  removeUnwantedImages
}

# Print the usage message
function printHelp () {
  echo "Usage: "
  echo "  network.sh -m up|down|restart|generate"
  echo "  network.sh -h|--help (print this message)"
  echo "    -m <mode> - one of 'up', 'down', 'restart' or 'generate'"
  echo "      - 'up' - bring up the network with docker-compose up"
  echo "      - 'down' - clear the network with docker-compose down"
  echo "      - 'restart' - restart the network"
  echo "      - 'generate' - generate required certificates and genesis block"
  echo "      - 'logs' - print and follow all docker instances log files"
  echo
  echo "Typically, one would first generate the required certificates and "
  echo "genesis block, then bring up the network. e.g.:"
  echo
  echo "	sudo network.sh -m generate"
  echo "	network.sh -m up"
  echo "	network.sh -m logs"
  echo "	network.sh -m down"
}

# Parse commandline args
while getopts "h?m:" opt; do
  case "$opt" in
    h|\?)
      printHelp
      exit 0
    ;;
    m)  MODE=$OPTARG
    ;;
  esac
done

if [ "${MODE}" == "up" ]; then
  networkUp
elif [ "${MODE}" == "down" ]; then
  networkDown
elif [ "${MODE}" == "clean" ]; then
  clean
elif [ "${MODE}" == "generate" ]; then
  clean
  removeArtifacts
  generateArtifacts
elif [ "${MODE}" == "restart" ]; then
  networkDown
  networkUp
elif [ "${MODE}" == "logs" ]; then
  logs
elif [ "${MODE}" == "channel" ]; then
  createChannel "$ORG2-$ORG3"
elif [ "${MODE}" == "join" ]; then
  joinChannel ${ORG1} "$ORG2-$ORG3"
  joinChannel ${ORG2} "$ORG2-$ORG3"
  joinChannel ${ORG3} "$ORG2-$ORG3"
elif [ "${MODE}" == "install" ]; then
  installChaincode ${ORG1} ${CHAINCODE_NAME} ${CHAINCODE_PATH}
elif [ "${MODE}" == "instantiate" ]; then
  instantiateChaincode ${ORG1} ${CHAINCODE_NAME} ${CHAINCODE_INIT}
elif [ "${MODE}" == "devup" ]; then
  devNetworkUp
elif [ "${MODE}" == "devinit" ]; then
  devInstallInstantiate
elif [ "${MODE}" == "devinvoke" ]; then
  devInvoke
elif [ "${MODE}" == "devquery" ]; then
  devQuery
elif [ "${MODE}" == "devlogs" ]; then
  devLogs
elif [ "${MODE}" == "devdown" ]; then
  devNetworkDown
else
  printHelp
  exit 1
fi
