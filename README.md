# Commercial Paper on Blockchain Pilot for NSD

Decentralized application manages instructions to transfer securities between members of NSD.
See [Functional Specification Google Doc](https://docs.google.com/document/d/1N2PjBoSN_M2hXXtBFyUv9HACu0Q-6WWqCv_TRcdIS8Y/edit?usp=sharing).

# Chaincode Development

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

# Local Deployment

Use docker instances to support four members on a local host.

Generate crypto material, genesis block, config transactions and start a network for a consortium of four organizations:

1. nsd
1. issuer
1. investor a
1. investor b

Each organization starts several docker instances:

- certificate authority [fabric-ca](https://github.com/hyperledger/fabric-ca)
- peer 0
- peer 1
- api server [fabric-rest](https://github.com/Altoros/fabric-rest)

Following channels are created, peers join them and chaincodes are installed and instantiated:

1. depository (*members*: nsd, chaincodes: [book](chaincode/go/book))
1. issuer-a (*members*: nsd, issuer, investor a, chaincode: instruction)
1. issuer-b (*members*: nsd, issuer, investor b, chaincode: instruction)
1. a-b (*members*: nsd, investor b, investor a, chaincode: instruction)

Generate artifacts for the network and network-config file for the API server:

`./network.sh -m generate`

**Note** you'll need to wait about 6 minutes until the timing of the generated certs lines up. 
Needed temporarily until the issue is resolved. This may come handy if you need to regenerate frequently:

`./network.sh -m generate && sleep 6m && beep && ./network.sh -m up`

Start the network, watch the logs, shutdown.

```
./network.sh -m up
./network.sh -m logs
./network.sh -m down
```

Navigate to web interfaces of respective organizations at ports 4000-4003:

1. [nsd](http://localhost:4000)
1. [issuer](http://localhost:4001)
1. [investor a](http://localhost:4002)
1. [investor b](http://localhost:4003)

