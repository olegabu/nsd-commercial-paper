# Commercial Paper on Blockchain for NSD v2 

Decentralized application manages instructions to transfer securities between members of NSD.
See [Functional Specification Google Doc](https://docs.google.com/document/d/1N2PjBoSN_M2hXXtBFyUv9HACu0Q-6WWqCv_TRcdIS8Y/edit?usp=sharing).

# Deployment with dockers run on separate hosts


## Install prerequisites

-   Clone Nsd Commercial Paper delivery packages from github:  
`git clone -b 2018_03-PRE_RELEASE_01 --depth=1 https://github.com/olegabu/nsd-commercial-paper`  
`cd nsd-commercial-paper`  
`./prerequisites-deployment.sh`

-	Or download NsdCommercialPaper.zip from   
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
	`./main-register-new-org.sh $ORG2 $IP2`  
	`./main-register-new-org.sh $ORG3 $IP3`  

*Note, when new organization is registerd it's added to the list of existing organizations `env-external-orgs-list`. 
This list is used to automatically create tri-lateral channel with new organization being added.   
This list may be adjust manually to control trilateral channels creation.*


On next step the members start the network on their nodes:
  
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


# Adding new organization

To add new organization into the network the following steps need to performed:  

1) Download NSD Commercial source package(s) to new org's server. See prerequisites section.

2) Environment file is created on new org with environment variables adjusted:  
    edit `env-org-neworgname`:
    ```
    ...  
    THIS_ORG=neworgname
    ```

3) New organization generates crypto-material:  
    New org:  
    `source ./env-org-neworgname`  
    `./org-generate-crypto.sh`

4) On NSD server configure the initialization configuration for *instruction* chaincode:  
    edit `instruction_init.json`   (add new organization account information)
    
5) Register new organization in blockchain, automatically creating bi-lateral and tri-lateral 
channels with exisiting organizations (using organizations in `env-external-orgs-list`):  
    Nsd:  
    `./main-register-new-org.sh neworgname <neworg_ip>`

6) Start blockchain on new org:  
    New org:  
    `./org-start-node.sh`
    
7) Mutually join new org to each existing organization (except NSD) and vice-versa:  
   - Sberbank:  
        New org:  
        `./org-join-org.sh sberbank <sberbank_ip>`  
        Sberbank:
        `./org-join-org.sh neworgname <neworg_ip>`
    
   - Mts:  
        New org:  
        `./org-join-org.sh mts <mts_ip>`  
        Mts:
        `./org-join-org.sh neworgname <neworg_ip>`

   - Other org (if any):  
        New org:  
        `./org-join-org.sh <otherorg_name> <otherorg_ip>`  
        Other Org:
        `./org-join-org.sh neworgname <neworg_ip>`

    ...  
    Repeat for all necessary organizations
