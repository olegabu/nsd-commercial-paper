# Commercial Paper on Blockchain Pilot for NSD

Decentralized application manages instructions to transfer securities between members of NSD.
See [Functional Specification Google Doc](https://docs.google.com/document/d/1N2PjBoSN_M2hXXtBFyUv9HACu0Q-6WWqCv_TRcdIS8Y/edit?usp=sharing).

# Deployment with dockers run on separate hosts


## Install prerequisites
-	Download NsdCommercialPaper.zip  
https://drive.google.com/file/d/18VFq9qxVdZIiKII2zbTY_MBFcQWiPJQ-/view?usp=sharing 


On other Linux distros make sure these versions or higher are installed:  

*Docker version 17.12.1*  
*docker-compose version 1.8.0*  
*jq*  

To install them on Ubuntu 16.04 you nay use the following commands:  

`cd fabric-starter`  
`./init-docker.sh`  


**Now re-login to have user applied into docker group.**  


Next execute in console:   
`cd fabric-starter`  
`./init-fabric.sh`    


##Configuration

For initial deployment the following organizations are used:
- ORG1 – nsd
- ORG2 – sberbank 
- ORG3 – mts

and the corresponded IP addresses:
- IP1=91.208.232.164 - NSD node's IP
- IP2=193.232.123.109 - Sberbank node's IP
- IP3=213.87.44.178 - MTS node's IP
 

In Commercial Paper v2 installation NSD serves as MAIN_NODE which is configured as environment variable exported in files *env-common*.
Other memebers are defined as THIS_ORG variable set correspondingly in *env-org-<org-name>* files.

Check initial configuration or reconfigure organization names, and IP-addresses in configuration files: 

Folder **nsd-commercial-paper**:
-	*env-common*
-	*env-org-sberbank*
-	*env-org-mts*  

as well as initialization arguments for blockhains :
-	*instruction_init.json*
-	*book_init.json*
-	*security_init.json*

##Deployment:

At first each member has to generate their crypto material; it then will be exposed by http interface on port 8080 to be accessible by the other organizations: 


1.	Sberbank:  
	`cd nsd-commercial-paper`  
	`source ./env-org-sberbank`  
	`./org-generate-crypto.sh`

2.	Mts:
	`cd nsd-commercial-paper`  
	`source ./env-org-mts`  
	`./org-generate-crypto.sh`

After that the main org (NSD) starts the blockchain network, adds the members one by one and creates *common*, *depository* and bilateral and trilateral channels:


3.	Nsd:  
	`cd nsd-commercial-paper`  
	`source ./env-org-nsd`  
	`./main-start-org.sh`  
	`./main-register-org $ORG2 $IP2`  
	`./main-register-org $ORG3 $IP3`

Then the members start the network on their nodes:
  
4.	Sberbank:  
	`./org-start-node.sh`

5.	Mts (after Sberbank's run is finished):  
	`./org-start-node.sh`

Now newly started members join each other:

6.	Sberbank:  
	`./org-join-org.sh $ORG3 $IP3`

7.	Mts (after Sberbank's joining is finished):  
	`./ org-join-org.sh $ORG2 $IP2`


Next start Commercial paper client:

8.	On all orgs:  
	`cd nsd-commercial-paper-client`  
	`./network.sh –m install`  
	`./network.sh –m up`
