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

**Note** you'll need to wait about 6 minutes until the timing of the generated certs lines up. 
Needed temporarily until the issue is resolved. This may come handy if you need to regenerate frequently:

`./network.sh -m generate && sleep 6m && beep && ./network.sh -m up`

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

Generate artifacts for the depository and each member. Pass organization name with `-o` and api server port as mapped to
the host machine with `-a`.

```bash
cd nsd
./network.sh -m generate-peer -o nsd -a 4000
cd ../a
./network.sh -m generate-peer -o a -a 4001
cd ../b
./network.sh -m generate-peer -o b -a 4002
cd ../c
./network.sh -m generate-peer -o c -a 4003
```

Now each member has generated their crypto material. The orderer can gather member certificates and use them to generate
genesis block files `artifacts/*.block` and channel config transaction files `artifacts/channel/*.tx`. The script will
copy cert files from respective member directories `a`, `b`, `c` into the folder `nsd` shared by depository nsd
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

Open terminal windows and start member instances. The script will copy cert files, channel block files received 
at creation of channels in the previous step and network-config from `nsd/artifacts` folder into its own `artifacts`. 
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





