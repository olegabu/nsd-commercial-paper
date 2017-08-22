# Commercial Paper on Blockchain Pilot for NSD

Decentralized application manages instructions to transfer securities between members of NSD.
See [Functional Specification Google Doc](https://docs.google.com/document/d/1N2PjBoSN_M2hXXtBFyUv9HACu0Q-6WWqCv_TRcdIS8Y/edit?usp=sharing).

# Chaincode development

Use docker instances to support chaincode development and debugging in an IDE.

Chaincode name, path and version are currently hardcoded in [network.sh](network.sh) to `book`, `chaincode/go/book`, `0`.
Feel free to improve the script to accept parameters or simply edit it locally with your own chaincode parameters.   

Start dev network of orderer, peer and cli with pre-generated [genesis block](ledger/dev-genesis.block), 
channel config transaction and identities in [dev-msp](ledger/dev-msp):

`./network.sh -m devup` 

Start chaincode in your IDE's debugger with env variables

```
CORE_CHAINCODE_ID_NAME=book:0
CORE_CHAINCODE_LOGGING_LEVEL=debug
CORE_PEER_ADDRESS=0.0.0.0:7051
```

You should see your chaincode connect to the dev peer. 

Now install and instantiate the chaincode on your dev peer:

`./network.sh -m devinit`

Invoke and query the chaincode with payload hardcoded in [network.sh](network.sh):

`./network.sh -m devinvoke`

Watch the logs and shutdown:

```
./network.sh -m devlogs
./network.sh -m devdown
```

# Local deployment with all dockers run from the same folder

Use docker instances to support four members on a local host.

Generate crypto material, genesis block, config transactions and start a network for a consortium of four organizations:

1. depository nsd
1. member a issuer
1. member b investor
1. member c investor

Each organization starts several docker instances:

- certificate authority [fabric-ca](https://github.com/hyperledger/fabric-ca)
- peer 0
- peer 1
- api server [fabric-rest](https://github.com/Altoros/fabric-rest)

Following channels are created, peers join them and chaincodes are installed and instantiated:

1. common (*members*: nsd, a, b, c; *chaincodes*: [security](chaincode/go/security))
1. depository (*members*: nsd; *chaincodes*: [book](chaincode/go/book))
1. a-b (*members*: nsd, a, b; *chaincodes*: [instruction](chaincode/go/instruction))
1. a-c (*members*: nsd, a, c; *chaincodes*: [instruction](chaincode/go/instruction))
1. b-c (*members*: nsd, b, c; *chaincodes*: [instruction](chaincode/go/instruction))
1. nsd-a (*members*: nsd, a; *chaincodes*: [position](chaincode/go/position))
1. nsd-b (*members*: nsd, b; *chaincodes*: [position](chaincode/go/position))
1. nsd-c (*members*: nsd, c; *chaincodes*: [position](chaincode/go/position))

Generate artifacts for the network and network-config file for the API server:

`./network.sh -m generate`

**Note** you'll need to wait about 7 minutes until the timing of the generated certs lines up. 
Needed temporarily until the issue is resolved. This may come handy if you need to regenerate frequently:

`./network.sh -m generate && sleep 7m && beep && ./network.sh -m up`

Start the network, watch the logs, shutdown.

```
./network.sh -m up
./network.sh -m logs -o nsd
./network.sh -m logs -o a
./network.sh -m logs -o b
./network.sh -m down
```

Navigate to web interfaces of respective organizations at ports 4000-4003:

1. [depository nsd](http://localhost:4000)
1. [issuer a](http://localhost:4001)
1. [investor b](http://localhost:4002)
1. [investor c](http://localhost:4003)

# Local deployment with dockers run from separate folders

Use to test artifacts generation and certificate exchange. Clone repo into separate directories imitating separate
host servers. Docker instances still operate within `ledger_default` network on the developer's host machine.

Clone repo and copy into 4 folders representing host machines for orderer and depository and for each member:

```bash
mkdir tmp
cd tmp
git clone https://github.com/olegabu/nsd-commercial-paper
mv nsd-commercial-paper nsd
cp -r nsd a
cp -r nsd b
cp -r nsd c
```

Generate artifacts for the depository and each member.

```bash
cd nsd
./network.sh -m generate-1
cd ../a
./network.sh -m generate-2
cd ../b
./network.sh -m generate-3
cd ../c
./network.sh -m generate-4
```

These will create docker-compose files for four orgs nsd, a, b, c with sets of mapped ports that don't conflict 
with each other on one host:

1. 4000 8080 7054 7051 7053 7056 7058
2. 4001 8081 8054 8051 8053 8056 8058
3. 4002 8082 9054 9051 9053 9056 9058
4. 4003 8083 10054 10051 10053 10056 10058

Also each member has generated their crypto material. The orderer can gather member certificates and use them to generate
genesis block files `artifacts/*.block` and channel config transaction files `artifacts/channel/*.tx`. The script will
download cert files from members `a`, `b`, `c` www servers into the folder `nsd` shared by depository nsd
and the orderer.

```bash
cd ../nsd
./network.sh -m generate-orderer && sleep 7m
```

Note the `sleep` above: generated certificates have their start time off so we need to wait at least 7 minutes before
starting up.

Start the orderer and the depository peers: nsd. Will create and join channels, install and instantiate chaincodes
on nsd peers:

```bash
./network.sh -m up-depository
``` 
Ignore `WARNING: Found orphan containers`.

Open terminal windows and start member instances. The script will download cert files from each member, 
channel block files received at creation of channels in the previous step and network-config from nsd www server 
into its own `artifacts`. 
Will start the ca server, peers and api server and tail their logs.

```bash
cd tmp/a
./network.sh -m up-2
``` 
```bash
cd tmp/b
./network.sh -m up-3
``` 
```bash
cd tmp/c
./network.sh -m up-4
```
 
# Deployment with dockers run on separate hosts

Real world deployment scenario with members deploying their CA server, peer, api and web servers as docker instances on
one host. Members' host servers connect to each other over internet. 

Each member clones the repo and generates artifacts. Pass organization name with `-o`. You can pass other ports as args
`-a`, `-w`, `-c`, `-0`, `-1`, `-2`, `-3`, `-4` for the api, web, ca and peer ports. If omitted,default ones are used.

```bash
git clone https://github.com/olegabu/nsd-commercial-paper
cd nsd-commercial-paper
./network.sh -m generate -o nsd
```

```bash
./network.sh -m generate -o a
```
```bash
./network.sh -m generate -o b
```
```bash
./network.sh -m generate -o c
```

These will create docker-compose files with default mapped ports that don't have to be different for each member as they
run on separate hosts: 

4000 8080 7054 7051 7053 7056 7058

Each member has generated their crypto material and is now serving their cert files to be gathered during the orderer's
generation process on depository host nsd:

```bash
./network.sh -m generate-orderer && sleep 7m
```
After that you can start the orderer and the depository peers: nsd. Will create and join channels, 
install and instantiate chaincodes on nsd peers:

```bash
./network.sh -m up-depository
``` 

Now the orderer has created channels, nsd peers instantiated chaincodes and other members can join channels by 
downloading channel *.block files.

Each member starts the ca server, peers and api servers:

```bash
./network.sh -m up-2
``` 
```bash
./network.sh -m up-3
``` 
```bash
./network.sh -m up-4
```



