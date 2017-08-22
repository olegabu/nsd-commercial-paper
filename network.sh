#!/usr/bin/env bash

STARTTIME=$(date +%s)

DOMAIN=nsd.ru
ORG1=nsd
ORG2=a
ORG3=b
ORG4=c

CLI_TIMEOUT=10000
COMPOSE_TEMPLATE=ledger/docker-composetemplate.yaml
COMPOSE_FILE_DEV=ledger/docker-composedev.yaml
HTTP_PORT=8080

GID=$(id -g)

# Delete any images that were generated as a part of this setup
# specifically the following images are often left behind:
# TODO list generated image naming patterns

function removeUnwantedContainers() {
  docker ps -a -q -f "name=dev-*"|xargs docker rm -f
}

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
  rm ledger/docker-compose-*.yaml
  rm -rf artifacts/crypto-config
  rm -rf artifacts/channel
}

function removeDockersFromCompose() {
    for O in ${DOMAIN} ${ORG1} ${ORG2} ${ORG3} ${ORG4}
    do
      COMPOSE_FILE="ledger/docker-compose-$O.yaml"

      if [ -f ${COMPOSE_FILE} ]; then
        echo "Removing docker containers listed in $COMPOSE_FILE"
        docker-compose -f ${COMPOSE_FILE} kill
        docker-compose -f ${COMPOSE_FILE} rm -f
      fi;
    done
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

function generateOrdererArtifacts() {
    echo "Creating orderer yaml files with $DOMAIN, $ORG1, $ORG2, $ORG3, $ORG4"

    COMPOSE_FILE="ledger/docker-compose-$DOMAIN.yaml"
    COMPOSE_TEMPLATE=ledger/docker-composetemplate-orderer.yaml

    # configtx and cryptogen
    sed -e "s/DOMAIN/$DOMAIN/g" -e "s/ORG1/$ORG1/g" -e "s/ORG2/$ORG2/g" -e "s/ORG3/$ORG3/g" -e "s/ORG4/$ORG4/g" artifacts/configtxtemplate.yaml > artifacts/configtx.yaml
    sed -e "s/DOMAIN/$DOMAIN/g" artifacts/cryptogentemplate-orderer.yaml > artifacts/"cryptogen-$DOMAIN.yaml"
    # docker-compose.yaml
    sed -e "s/DOMAIN/$DOMAIN/g" -e "s/ORG1/$ORG1/g" -e "s/ORG2/$ORG2/g" -e "s/ORG3/$ORG3/g" -e "s/ORG4/$ORG4/g" ${COMPOSE_TEMPLATE} > ${COMPOSE_FILE}
    # network-config.json
    #  fill environments                                                                                            |   remove comments
    sed -e "s/\DOMAIN/$DOMAIN/g" -e "s/\ORG1/$ORG1/g" -e "s/\ORG2/$ORG2/g" -e "s/\ORG3/$ORG3/g" -e "s/\ORG4/$ORG4/g" -e "s/^\s*\/\/.*$//g" artifacts/network-config-template.json > artifacts/network-config.json

    echo "Generating crypto material with cryptogen"
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "cryptogen generate --config=cryptogen-$DOMAIN.yaml"

    echo "Change artifacts file ownership"
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "chown -R $UID:$GID ."

    echo "Generating orderer genesis block with configtxgen"
    mkdir -p artifacts/channel
    docker-compose --file ${COMPOSE_FILE} run --rm -e FABRIC_CFG_PATH=/etc/hyperledger/artifacts "cli.$DOMAIN" configtxgen -profile OrdererGenesis -outputBlock ./channel/genesis.block

    for CHANNEL_NAME in depository common "$ORG2-$ORG3" "$ORG2-$ORG4" "$ORG3-$ORG4" "$ORG1-$ORG2" "$ORG1-$ORG3" "$ORG1-$ORG4"
    do
        echo "Generating channel config transaction for $CHANNEL_NAME"
        docker-compose --file ${COMPOSE_FILE} run --rm -e FABRIC_CFG_PATH=/etc/hyperledger/artifacts "cli.$DOMAIN" configtxgen -profile "$CHANNEL_NAME" -outputCreateChannelTx ./channel/"$CHANNEL_NAME".tx -channelID "$CHANNEL_NAME"
    done
}

function generatePeerArtifacts() {
    ORG=$1
    API_PORT=$2

    echo "Creating peer yaml files with $DOMAIN, $ORG, $API_PORT"

    COMPOSE_FILE="ledger/docker-compose-$ORG.yaml"
    COMPOSE_TEMPLATE=ledger/docker-composetemplate-peer.yaml

    # cryptogen
    sed -e "s/DOMAIN/$DOMAIN/g" -e "s/ORG/$ORG/g" artifacts/cryptogentemplate-peer.yaml > artifacts/"cryptogen-$ORG.yaml"

    # docker-compose.yaml
    sed -e "s/DOMAIN/$DOMAIN/g" -e "s/\([^ ]\)ORG/\1$ORG/g" -e "s/API_PORT/$API_PORT/g" ${COMPOSE_TEMPLATE} > ${COMPOSE_FILE}

    # fabric-ca-server-config.yaml
    sed -e "s/ORG/$ORG/g" artifacts/fabric-ca-server-configtemplate.yaml > artifacts/"fabric-ca-server-config-$ORG.yaml"

    echo "Generating crypto material with cryptogen"
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "cryptogen generate --config=cryptogen-$ORG.yaml"

    echo "Change artifacts ownership"
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "chown -R $UID:$GID ."

    echo "Adding generated CA private keys filenames to $COMPOSE_FILE"
    CA_PRIVATE_KEY=$(basename `ls artifacts/crypto-config/peerOrganizations/"$ORG.$DOMAIN"/ca/*_sk`)
    [[ -z  ${CA_PRIVATE_KEY}  ]] && echo "empty CA private key" && exit 1
    sed -i -e "s/CA_PRIVATE_KEY/${CA_PRIVATE_KEY}/g" ${COMPOSE_FILE}
}

function servePeerArtifacts() {
    ORG=$1
    COMPOSE_FILE="ledger/docker-compose-$ORG.yaml"
    
    echo "Copying generated TLS cert files to be served by www.$ORG.$DOMAIN"
    D="artifacts/crypto-config/peerOrganizations/$ORG.$DOMAIN/peers/peer0.$ORG.$DOMAIN/tls"
    mkdir -p "www/${D}"
    cp "${D}/ca.crt" "www/${D}"

    echo "Copying generated MSP cert files to be served by www.$ORG.$DOMAIN"
    D="artifacts/crypto-config/peerOrganizations/$ORG.$DOMAIN"
    cp -r "${D}/msp" "www/${D}"

    docker-compose --file ${COMPOSE_FILE} up -d "www.$ORG.$DOMAIN"
}

function serveOrdererArtifacts() {
    COMPOSE_FILE="ledger/docker-compose-$ORG1.yaml"

    D="artifacts/crypto-config/ordererOrganizations/$DOMAIN/orderers/orderer.$DOMAIN/tls"
    echo "Copying generated orderer TLS cert files from $D to be served by www.$ORG1.$DOMAIN"
    mkdir -p "www/${D}"
    cp "${D}/ca.crt" "www/${D}"

    D="artifacts"
    echo "Copying generated network config file from $D to be served by www.$ORG1.$DOMAIN"
    cp "${D}/network-config.json" "www/${D}"

    echo "Copying channel block files from $D to be served by www.$ORG1.$DOMAIN"
    cp ${D}/*.block "www/${D}"

    docker-compose --file ${COMPOSE_FILE} up -d "www.$ORG1.$DOMAIN"
}

function createChannel () {
    CHANNEL_NAME=$1
    F="ledger/docker-compose-${ORG1}.yaml"

    info "creating channel $CHANNEL_NAME by $ORG1 using $F"

    docker-compose --file ${F} run --rm "cli.$ORG1.$DOMAIN" bash -c "peer channel create -o orderer.$DOMAIN:7050 -c $CHANNEL_NAME -f /etc/hyperledger/artifacts/channel/$CHANNEL_NAME.tx --tls --cafile /etc/hyperledger/crypto/orderer/tls/ca.crt"

    echo "Change channel block file ownership"
    docker-compose --file ${F} run --rm "cli.$DOMAIN" bash -c "chown -R $UID:$GID ."
}

function joinChannel() {
    ORG=$1
    CHANNEL_NAME=$2
    F="ledger/docker-compose-${ORG}.yaml"

    info "joining channel $CHANNEL_NAME by $ORG using $F"

    docker-compose --file ${F} run --rm "cli.$ORG.$DOMAIN" bash -c "CORE_PEER_ADDRESS=peer0.$ORG.$DOMAIN:7051 peer channel join -b $CHANNEL_NAME.block      && CORE_PEER_ADDRESS=peer1.$ORG.$DOMAIN:7051 peer channel join -b $CHANNEL_NAME.block"
}

function instantiateChaincode () {
    ORG=$1
    CHANNEL_NAME=$2
    N=$3
    I=$4
    F="ledger/docker-compose-${ORG}.yaml"

    info "instantiating chaincode $N on $CHANNEL_NAME by $ORG with $I using $F"

    docker-compose --file ${F} run --rm "cli.$ORG.$DOMAIN" bash -c "CORE_PEER_ADDRESS=peer0.$ORG.$DOMAIN:7051 peer chaincode instantiate -n $N -v 1.0 -c '$I' -o orderer.$DOMAIN:7050 -C $CHANNEL_NAME --tls --cafile /etc/hyperledger/crypto/orderer/tls/ca.crt"
}

function warmUpChaincode () {
    ORG=$1
    CHANNEL_NAME=$2
    N=$3
    F="ledger/docker-compose-${ORG}.yaml"

    info "warming up chaincode $N on $CHANNEL_NAME on all peers of $ORG with query using $F"

    docker-compose --file ${F} run --rm "cli.$ORG.$DOMAIN" bash -c "CORE_PEER_ADDRESS=peer0.$ORG.$DOMAIN:7051 peer chaincode query -n $N -v 1.0 -c '{\"Args\":[\"query\"]}' -C $CHANNEL_NAME"
    docker-compose --file ${F} run --rm "cli.$ORG.$DOMAIN" bash -c "CORE_PEER_ADDRESS=peer1.$ORG.$DOMAIN:7051 peer chaincode query -n $N -v 1.0 -c '{\"Args\":[\"query\"]}' -C $CHANNEL_NAME"
}

function installChaincode() {
    ORG=$1
    N=$2
    # chaincode path is the same as chaincode name by convention: code of chaincode instruction lives in ./chaincode/go/instruction mapped to docker path /opt/gopath/src/instruction
    P=${N}
    F="ledger/docker-compose-${ORG}.yaml"

    info "installing chaincode $N to peers of $ORG from ./chaincode/go/$P using $F"

    docker-compose --file ${F} run --rm "cli.$ORG.$DOMAIN" bash -c "CORE_PEER_ADDRESS=peer0.$ORG.$DOMAIN:7051 peer chaincode install -n $N -v 1.0 -p $P && CORE_PEER_ADDRESS=peer1.$ORG.$DOMAIN:7051 peer chaincode install -n $N -v 1.0 -p $P"
}

function dockerComposeUp () {
  COMPOSE_FILE="ledger/docker-compose-$1.yaml"

  info "starting docker instances from $COMPOSE_FILE"

  TIMEOUT=${CLI_TIMEOUT} docker-compose -f ${COMPOSE_FILE} up -d 2>&1
  if [ $? -ne 0 ]; then
    echo "ERROR !!!! Unable to start network"
    logs ${1}
    exit 1
  fi
}

function dockerComposeDown () {
  COMPOSE_FILE="ledger/docker-compose-$1.yaml"

  if [ -f ${COMPOSE_FILE} ]; then
      info "stopping docker instances from $COMPOSE_FILE"
      docker-compose -f ${COMPOSE_FILE} down
  fi;

}

function installInstantiateWarmUp() {
  CHAINCODE_NAME=$1
  CHANNEL_NAME=$2
  CHAINCODE_INIT=$3

  installChaincode ${ORG1} ${CHAINCODE_NAME}
  instantiateWarmUp ${CHAINCODE_NAME} ${CHANNEL_NAME} ${CHAINCODE_INIT}
}

function instantiateWarmUp() {
  CHAINCODE_NAME=$1
  CHANNEL_NAME=$2
  CHAINCODE_INIT=$3

  instantiateChaincode ${ORG1} ${CHANNEL_NAME} ${CHAINCODE_NAME} ${CHAINCODE_INIT}
  sleep 7
  warmUpChaincode ${ORG1} ${CHANNEL_NAME} ${CHAINCODE_NAME}
}

function joinWarmUp() {
  ORG=$1
  CHAINCODE_NAME=$2
  CHANNEL_NAME=$3

  joinChannel ${ORG} ${CHANNEL_NAME}
  sleep 7
  warmUpChaincode ${ORG} ${CHANNEL_NAME} ${CHAINCODE_NAME}
}

function startDepository () {
  for CHANNEL_NAME in depository common "$ORG2-$ORG3" "$ORG2-$ORG4" "$ORG3-$ORG4" "$ORG1-$ORG2" "$ORG1-$ORG3" "$ORG1-$ORG4"
    do
      createChannel ${CHANNEL_NAME}
      joinChannel ${ORG1} ${CHANNEL_NAME}
    done

  installInstantiateWarmUp book depository '{"Args":["init","AC0689654902","87680000045800005","RU000ABC0001","100"]}'

  installInstantiateWarmUp security common '{"Args":["init","RU000ABC0001","active"]}'

  installChaincode ${ORG1} instruction

  for CHANNEL_NAME in "$ORG2-$ORG3" "$ORG2-$ORG4" "$ORG3-$ORG4"
    do
      instantiateWarmUp instruction ${CHANNEL_NAME} '{"Args":["init","[{\"organization\":\"a.nsd.ru\",\"balances\":[{\"account\":\"AC0689654902\",\"division\":\"87680000045800005\"},{\"account\":\"AC0689654902\",\"division\":\"69070000982300006\"},{\"account\":\"AC0191654904\",\"division\":\"80120002322000007\"},{\"account\":\"AC0191654904\",\"division\":\"36060003558300008\"}]},{\"organization\":\"b.nsd.ru\",\"balances\":[{\"account\":\"WD0D00654903\",\"division\":\"58680002816000009\"},{\"account\":\"WD0D00654903\",\"division\":\"11560007930600010\"},{\"account\":\"WD0H7B654905\",\"division\":\"51630003768000011\"},{\"account\":\"WD0H7B654905\",\"division\":\"36090008645500012\"}]},{\"organization\":\"c.nsd.ru\",\"balances\":[{\"account\":\"YN0000654906\",\"division\":\"6294000472000013\"},{\"account\":\"YN0000654906\",\"division\":\"57680007190700014\"},{\"account\":\"YN0927654908\",\"division\":\"9384000328700015\"},{\"account\":\"YN0927654908\",\"division\":\"37800007360900016\"}]}]"]}'
    done

  installChaincode ${ORG1} position

  for CHANNEL_NAME in "$ORG1-$ORG2" "$ORG1-$ORG3" "$ORG1-$ORG4"
    do
      instantiateWarmUp position ${CHANNEL_NAME} '{"Args":["init"]}'
    done
}

function startMember () {
  ORG=$1

  installChaincode ${ORG} security
  installChaincode ${ORG} instruction
  installChaincode ${ORG} position

  joinWarmUp ${ORG} security common

  joinWarmUp ${ORG} position "${ORG1}-${ORG}"

  for CHANNEL_NAME in ${@:2}
    do
      joinWarmUp ${ORG} instruction ${CHANNEL_NAME}
    done
}

function makeCertDirs() {
  # crypto-config/ordererOrganizations/nsd.ru/orderers/orderer.nsd.ru/tls/ca.crt"
  mkdir -p "artifacts/crypto-config/ordererOrganizations/$DOMAIN/orderers/orderer.$DOMAIN/tls"

  for ORG in ${ORG1} ${ORG2} ${ORG3} ${ORG4}
    do
        # crypto-config/peerOrganizations/nsd.nsd.ru/peers/peer0.nsd.nsd.ru/tls/ca.crt
        D="artifacts/crypto-config/peerOrganizations/$ORG.$DOMAIN/peers/peer0.$ORG.$DOMAIN/tls"
        echo "mkdir -p ${D}"
        mkdir -p ${D}
    done
}

function copyMemberMSP() {
  for ORG in ${ORG2} ${ORG3} ${ORG4}
    do
        # cp ../a/artifacts/crypto-config/peerOrganizations/a.nsd.ru/msp/ artifacts/crypto-config/peerOrganizations/a.nsd.ru
        S="../$ORG/artifacts/crypto-config/peerOrganizations/$ORG.$DOMAIN/msp/"
        D="artifacts/crypto-config/peerOrganizations/$ORG.$DOMAIN/"
        echo "cp -r $S $D"
        cp -r ${S} ${D}
    done
}

function downloadMemberMSP() {
    COMPOSE_FILE="ledger/docker-compose-$ORG1.yaml"

    C="for ORG in ${ORG1} ${ORG2} ${ORG3} ${ORG4}; do wget --verbose --directory-prefix crypto-config/peerOrganizations/\$ORG.$DOMAIN/msp/admincerts http://www.\$ORG.$DOMAIN:$HTTP_PORT/crypto-config/peerOrganizations/\$ORG.$DOMAIN/msp/admincerts/Admin@\$ORG.$DOMAIN-cert.pem && wget --verbose --directory-prefix crypto-config/peerOrganizations/\$ORG.$DOMAIN/msp/cacerts http://www.\$ORG.$DOMAIN:$HTTP_PORT/crypto-config/peerOrganizations/\$ORG.$DOMAIN/msp/cacerts/ca.\$ORG.$DOMAIN-cert.pem && wget --verbose --directory-prefix crypto-config/peerOrganizations/\$ORG.$DOMAIN/msp/tlscacerts http://www.\$ORG.$DOMAIN:$HTTP_PORT/crypto-config/peerOrganizations/\$ORG.$DOMAIN/msp/tlscacerts/tlsca.\$ORG.$DOMAIN-cert.pem; done"
    echo ${C}
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "${C}"

    echo "Change artifacts file ownership"
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "chown -R $UID:$GID ."
}

function copyNetworkConfig() {
    S="../$ORG1/artifacts/network-config.json"
    D="artifacts"
    echo "cp $S $D"
    cp ${S} ${D}
}

function downloadNetworkConfig() {
    COMPOSE_FILE="ledger/docker-compose-$1.yaml"

    C="wget --verbose http://www.$ORG1.$DOMAIN:$HTTP_PORT/network-config.json && chown -R $UID:$GID ."
    echo ${C}
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "${C}"
}

function copyChannelBlockFiles() {
    ORG=$1

    for CHANNEL_NAME in common "$ORG1-$ORG" ${@:2}
    do
      S="../$ORG1/artifacts/$CHANNEL_NAME.block"
      D="artifacts"
      echo "cp $S $D"
      cp ${S} ${D}
    done
}

function downloadChannelBlockFiles() {
    ORG=$1
    COMPOSE_FILE="ledger/docker-compose-$ORG.yaml"

    for CHANNEL_NAME in common "$ORG1-$ORG" ${@:2}
    do
      C="wget --verbose http://www.$ORG1.$DOMAIN:$HTTP_PORT/$CHANNEL_NAME.block && chown -R $UID:$GID ."
      echo ${C}
      docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "${C}"
    done
}

function startMemberWithCopy() {
    copyArtifactsMember ${@}
    dockerComposeUp ${1}
    startMember ${@}
}

function startMemberWithDownload() {
    downloadArtifactsMember ${@}
    dockerComposeUp ${1}
    startMember ${@}
}

function copyCerts() {
    S="../$ORG1/artifacts/crypto-config/ordererOrganizations/$DOMAIN/orderers/orderer.$DOMAIN/tls/ca.crt"
    D="artifacts/crypto-config/ordererOrganizations/$DOMAIN/orderers/orderer.$DOMAIN/tls/"
    echo "cp $S $D"
    cp ${S} ${D}

  for ORG in ${ORG1} ${ORG2} ${ORG3} ${ORG4}
    do
	S="../$ORG/artifacts/crypto-config/peerOrganizations/$ORG.$DOMAIN/peers/peer0.$ORG.$DOMAIN/tls/ca.crt"
	D="artifacts/crypto-config/peerOrganizations/$ORG.$DOMAIN/peers/peer0.$ORG.$DOMAIN/tls"
	echo "cp $S $D"
        cp ${S} ${D}
    done
}

function downloadCerts() {
    COMPOSE_FILE="ledger/docker-compose-$1.yaml"

    C="wget --verbose --directory-prefix crypto-config/ordererOrganizations/$DOMAIN/orderers/orderer.$DOMAIN/tls http://www.$ORG1.$DOMAIN:$HTTP_PORT/crypto-config/ordererOrganizations/$DOMAIN/orderers/orderer.$DOMAIN/tls/ca.crt"
    echo ${C}
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "${C}"

    C="for ORG in ${ORG1} ${ORG2} ${ORG3} ${ORG4}; do wget --verbose --directory-prefix crypto-config/peerOrganizations/\${ORG}.$DOMAIN/peers/peer0.\${ORG}.$DOMAIN/tls http://www.\${ORG}.$DOMAIN:$HTTP_PORT/crypto-config/peerOrganizations/\${ORG}.$DOMAIN/peers/peer0.\${ORG}.$DOMAIN/tls/ca.crt; done"
    echo ${C}
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "${C}"

    echo "Change artifacts file ownership"
    docker-compose --file ${COMPOSE_FILE} run --rm "cli.$DOMAIN" bash -c "chown -R $UID:$GID ."
}

function copyArtifactsMember() {
  makeCertDirs
  copyCerts
  copyNetworkConfig
  copyChannelBlockFiles ${@}
}

function downloadArtifactsMember() {
  makeCertDirs
  downloadCerts ${1}
  downloadNetworkConfig ${1}
  downloadChannelBlockFiles ${@}
}

function copyArtifactsDepository() {
  for ORG in ${ORG2} ${ORG3} ${ORG4}
    do
      rm -rf "artifacts/crypto-config/peerOrganizations/$ORG.$DOMAIN"
    done

  makeCertDirs
  copyCerts
  copyMemberMSP
}

function downloadArtifactsDepository() {
  for ORG in ${ORG2} ${ORG3} ${ORG4}
    do
      rm -rf "artifacts/crypto-config/peerOrganizations/$ORG.$DOMAIN"
    done

  makeCertDirs
  downloadCerts ${ORG1}
  downloadMemberMSP
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
# docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode install -p book -n book -v 0"
 docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode instantiate -n book -v 0 -C myc -c '{\"Args\":[\"init\",\"AC0689654902\",\"87680000045800005\",\"RU000ABC0001\",\"100\"]}'"

 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode install -p instruction -n instruction -v 0 && peer chaincode instantiate -n instruction -v 0 -C myc -c '{\"Args\":[\"init\"]}'"
 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode instantiate -n instruction -v 0 -C myc -c '{\"Args\":[\"init\"]}'"

 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode install -p security -n security -v 0 && peer chaincode instantiate -n security -v 0 -C myc -c '{\"Args\":[\"init\",\"RU000ABC0001\",\"active\"]}'"
 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode instantiate -n security -v 0 -C myc -c '{\"Args\":[\"init\",\"RU000ABC0001\",\"active\"]}'"

# docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode install -p position -n position -v 0 && peer chaincode instantiate -n position -v 0 -C myc -c '{\"Args\":[\"init\"]}'"
}

function devInvoke () {
 # ["AC0689654902","87680000045800005","WD0D00654903","58680002816000009","RU000ABC0001","10"]
 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode invoke -n book -v 0 -C myc -c '{\"Args\":[\"move\",\"AC0689654902\",\"87680000045800005\",\"WD0D00654903\",\"58680002816000009\",\"RU000ABC0001\",\"10\"]}'"
 docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode invoke -n book -v 0 -C myc -c '{\"Args\":[\"move\",\"AC0689654902\",\"87680000045800005\",\"WD0D00654903\",\"58680002816000009\",\"RU000ABC0001\",\"10\",\"a\",\"2017-08-12\",\"2017-08-12\"]}'"

 # ["DE000DB7HWY7","AC0689654902","87680000045800005","CA9861913023","WD0D00654903","58680002816000009","RU000ABC0001","10","reference1000","2017-08-08","2017-08-07","reason"]
 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode invoke -n instruction -v 0 -C myc -c '{\"Args\":[\"receive\",\"aDeponent\",\"aEmissionAccount\",\"aActiveDivision\",\"bInvestmentAccount\",\"bActiveDivision\",\"RU000ABC0001\",\"10\",\"reference1000\",\"2017-08-08\",\"2017-08-07\",\"reason\"]}'"
 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode invoke -n instruction -v 0 -C myc -c '{\"Args\":[\"check\",\"aEmissionAccount\",\"aActiveDivision\",\"RU000ABC0001\",\"10\"]}'"

 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode invoke -n security -v 0 -C myc -c '{\"Args\":[\"put\",\"RU000ABC0001\",\"redeemed\"]}'"

 # ["AC0689654902","87680000045800005","RU000ABC0001","10"]
 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode invoke -n position -v 0 -C myc -c '{\"Args\":[\"put\",\"AC0689654902\",\"87680000045800005\",\"RU000ABC0001\",\"10\"]}'"
}

function devQuery () {
 docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode query -n book -v 0 -C myc -c '{\"Args\":[\"query\"]}'"
 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode query -n book -v 0 -C myc -c '{\"Args\":[\"history\",\"aEmissionAccount\",\"aActiveDivision\",\"RU000ABC0001\"]}'"

 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode query -n instruction -v 0 -C myc -c '{\"Args\":[\"query\"]}'"

 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode query -n security -v 0 -C myc -c '{\"Args\":[\"query\"]}'"

 #docker-compose -f ${COMPOSE_FILE_DEV} run cli bash -c "peer chaincode query -n position -v 0 -C myc -c '{\"Args\":[\"query\"]}'"
}

function info() {
    #figlet $1
    echo "*************************************************************************************************************"
    echo "$1"
    echo "*************************************************************************************************************"
}

function logs () {
  COMPOSE_FILE="ledger/docker-compose-$1.yaml"

  TIMEOUT=${CLI_TIMEOUT} COMPOSE_HTTP_TIMEOUT=${CLI_TIMEOUT} docker-compose -f ${COMPOSE_FILE} logs -f
}

function devLogs () {
    TIMEOUT=${CLI_TIMEOUT} COMPOSE_HTTP_TIMEOUT=${CLI_TIMEOUT} docker-compose -f ${COMPOSE_FILE_DEV} logs -f
}

function clean() {
  removeDockersFromCompose
#  removeDockersWithDomain
  removeUnwantedImages
}

function generateWait() {
  echo "$(date --rfc-3339='seconds' -u) *** Wait for 7 minutes to make sure the certificates become active ***"
  sleep 7m
  beep
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
while getopts "h?m:o:a:" opt; do
  case "$opt" in
    h|\?)
      printHelp
      exit 0
    ;;
    m)  MODE=$OPTARG
    ;;
    o)  ORG=$OPTARG
    ;;
    a)  API_PORT=$OPTARG
    ;;
  esac
done

if [ "${MODE}" == "up" -a "${ORG}" == "" ]; then
  dockerComposeUp ${DOMAIN}
  dockerComposeUp ${ORG1}
  dockerComposeUp ${ORG2}
  dockerComposeUp ${ORG3}
  dockerComposeUp ${ORG4}
  startDepository
  startMember ${ORG2} "${ORG2}-${ORG3}" "${ORG2}-${ORG4}"
  startMember ${ORG3} "${ORG2}-${ORG3}" "${ORG3}-${ORG4}"
  startMember ${ORG4} "${ORG2}-${ORG4}" "${ORG3}-${ORG4}"
elif [ "${MODE}" == "up" ]; then
  dockerComposeUp ${ORG}
elif [ "${MODE}" == "down" ]; then
  dockerComposeDown ${DOMAIN}
  dockerComposeDown ${ORG1}
  dockerComposeDown ${ORG2}
  dockerComposeDown ${ORG3}
  dockerComposeDown ${ORG4}
  removeUnwantedContainers
  removeUnwantedImages
elif [ "${MODE}" == "clean" ]; then
  clean
elif [ "${MODE}" == "generate" ]; then
  clean
  removeArtifacts
  generatePeerArtifacts ${ORG1} 4000
  generatePeerArtifacts ${ORG2} 4001
  generatePeerArtifacts ${ORG3} 4002
  generatePeerArtifacts ${ORG4} 4003
  generateOrdererArtifacts
  generateWait
elif [ "${MODE}" == "generate-orderer" ]; then
  downloadArtifactsDepository
  generateOrdererArtifacts
elif [ "${MODE}" == "generate-peer" ]; then
  clean
  removeArtifacts
  generatePeerArtifacts ${ORG} ${API_PORT}
  servePeerArtifacts ${ORG}
elif [ "${MODE}" == "download-artifacts-depository" ]; then
  downloadArtifactsDepository
elif [ "${MODE}" == "download-certs" ]; then
  downloadCerts ${ORG}
elif [ "${MODE}" == "serve-orderer-artifacts" ]; then
  serveOrdererArtifacts
elif [ "${MODE}" == "up-depository" ]; then
  dockerComposeUp ${DOMAIN}
  dockerComposeUp ${ORG1}
  startDepository
  serveOrdererArtifacts
elif [ "${MODE}" == "up-2" ]; then
  startMemberWithDownload ${ORG2} "${ORG2}-${ORG3}" "${ORG2}-${ORG4}"
elif [ "${MODE}" == "up-3" ]; then
  startMemberWithDownload ${ORG3} "${ORG2}-${ORG3}" "${ORG3}-${ORG4}"
elif [ "${MODE}" == "up-4" ]; then
  startMemberWithDownload ${ORG4} "${ORG2}-${ORG4}" "${ORG3}-${ORG4}"
elif [ "${MODE}" == "logs" ]; then
  logs ${ORG}
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

# print spent time
ENDTIME=$(date +%s)
echo "Finished in $(($ENDTIME - $STARTTIME)) seconds"
