/**
 * @class InstructionService
 * @classdesc
 * @ngInject
 */
function InstructionService(ApiService, ConfigLoader, $q, $log) {

  // jshint shadow: true
  var InstructionService = this;

  InstructionService.status = {
      MATCHED : 'matched',
      DECLINED: 'declined',
      EXECUTED: 'executed',
      CANCELED: 'canceled'
  };

  /**
   *
   */
  InstructionService._getChaincodeID = function() {
    var chaincodeID = ConfigLoader.get()['contracts'].instruction;
    if(!chaincodeID){
      // must be specified in network-config.json
      throw new Error("No chaincode name for 'instruction' contract");
    }
    return chaincodeID;
  };


  /**
   *
   */
  InstructionService.listAll = function() {
    $log.debug('InstructionService.listAll');

    var chaincodeID = InstructionService._getChaincodeID();
    var peer = InstructionService._getQueryPeer();

    return ApiService.channels.list().then(function(channelList){
      return $q.all( channelList
        .map(function(channel){ return channel.channel_id; })
        .filter(function(channelID){ return InstructionService.isBilateralChannel(channelID); })
        .sort()
        .map(function(channelID){
          // promise for each channel:
          return ApiService.sc.query(channelID, chaincodeID, peer, 'query')
            .then(function(data){ return {
                channel: channelID,
                result: parseJson(data.result)
              };
            }).catch(function(){
              return {
                channel: channelID,
                result: []
              };
            });

      }));
    }).then(function(results){
      // join array of results into one array (groupedList)
      return results.reduce(function(result, singleResult){
        result[singleResult.channel] = singleResult.result;
        return result;
      }, {});
    });
  };

  /**
   *
   */
  function parseJson(data){
    if(typeof data == "string"){
      try{
        data = JSON.parse(data);
      }catch(e){
        $log.warn(e, data);
      }
    }
    return data;
  }

  InstructionService.isBilateralChannel = function(channelID){
    return channelID.indexOf('-') > 0 && !channelID.startsWith('nsd-');
  }



  /**
   *
   */
  InstructionService.transfer = function(instruction) {
    $log.debug('InstructionService.transfer', instruction);

    var chaincodeID = InstructionService._getChaincodeID();
    var channelID   = InstructionService._getInstructionChannel(instruction);
    var peers       = InstructionService._getEndorsePeers(instruction);
    var args        = InstructionService._instructionArguments(instruction);

    args.push(
      instruction.deponentFrom,
      instruction.deponentTo,
      JSON.stringify(instruction.reason)
    );

    return ApiService.sc.invoke(channelID, chaincodeID, peers, 'transfer', args);
  };

  /**
   *
   */
  InstructionService.receive = function(instruction) {
    $log.debug('InstructionService.receive', instruction);

    var chaincodeID = InstructionService._getChaincodeID();
    var channelID   = InstructionService._getInstructionChannel(instruction);
    var peers       = InstructionService._getEndorsePeers(instruction);
    var args        = InstructionService._instructionArguments(instruction);

    args.push(
      instruction.deponentFrom,
      instruction.deponentTo,
      JSON.stringify(instruction.reason)
    );

    return ApiService.sc.invoke(channelID, chaincodeID, peers, 'receive', args);
  };

  /**
   *
   */
  InstructionService.cancelInstruction = function(instruction) {
    $log.debug('InstructionService.cancelInstruction', instruction);

    var chaincodeID = InstructionService._getChaincodeID();
    var channelID   = InstructionService._getInstructionChannel(instruction);
    var peers       = InstructionService._getEndorsePeers(instruction);
    var args        = InstructionService._instructionArguments(instruction);

    args.push(InstructionService.status.CANCELED);

    return ApiService.sc.invoke(channelID, chaincodeID, peers, 'status', args);
  };




  /**
   * Expecting deponentFrom, accountFrom, divisionFrom, deponentTo, accountTo, divisionTo, security, quantity, reference, instructionDate, tradeDate)
   */
  InstructionService.history = function(instruction){
    $log.debug('InstructionService.history', instruction);

    var chaincodeID = InstructionService._getChaincodeID();
    var channelID   = InstructionService._getInstructionChannel(instruction);
    var peer        = InstructionService._getQueryPeer();
    var args        = InstructionService._instructionArguments(instruction);

    return ApiService.sc.query(channelID, chaincodeID, peer, 'history', args);
  }


  /**
   * return basic fields for any instruction request
   * @return {Array<string>}
   */
  InstructionService._instructionArguments = function(instruction) {
    var args = [
      instruction.transferer.account,  // 0: accountFrom
      instruction.transferer.division, // 1: divisionFrom

      instruction.receiver.account,    // 2: accountTo
      instruction.receiver.division,   // 3: divisionTo

      instruction.security,            // 4: security
      ''+instruction.quantity,         // 5: quantity // TODO: fix: string parameters
      instruction.reference,           // 6: reference
      instruction.instructionDate,     // 7: instructionDate  (ISO)
      instruction.tradeDate,           // 8: tradeDate  (ISO)
    ];

    return args;
  }


  /**
   * get instruction opponents ID.
   * @param {Instruction} instruction
   * @return {Array<string>} multiple (two actually) orgID
   */
  InstructionService._getInstructionOrgs = function(instruction) {
    var org1 = ConfigLoader.getOrgByDepcode(instruction.deponentFrom);
    if(!org1){
      throw new Error("Deponent owner not found: " + instruction.deponentFrom);
    }
    var org2 = ConfigLoader.getOrgByDepcode(instruction.deponentTo);
    if(!org2){
      throw new Error("Deponent owner not found: " + instruction.deponentTo);
    }
    return [org1, org2];
  }

  /**
   * get name of bi-lateral channel for opponent and the organisation
   */
  InstructionService._getInstructionChannel = function(instruction) {
    var orgArr = InstructionService._getInstructionOrgs(instruction);
    // make channel name as '<org1_ID>-<org2_ID>'.
    // Please, pay attention to the orgs order - ot should be sorted
    return orgArr.sort().join('-');
  };

  /**
   * get orgPeerIDs of endorsers, which should endose the transaction
   * @return {Array<string>}
   */
  InstructionService._getEndorsePeers = function(instruction) {
    var endorserOrgs = InstructionService._getInstructionOrgs(instruction);

    // root endorser
    var config = ConfigLoader.get();
    var rootEndorsers = config.endorsers || [];
    endorserOrgs.push.apply(endorserOrgs, rootEndorsers);

    //
    var peers = endorserOrgs.reduce(function(result, org){
      var peers = ConfigLoader.getOrgPeerIds(org);
      result.push( org+'/'+peers[0] ); // orgPeerID  // endorse by the first peer
      return result;
    }, []);

    return peers;
  };


  /**
   *
   */
  InstructionService._getQueryPeer = function() {
    var config = ConfigLoader.get();
    var peers = ConfigLoader.getOrgPeerIds(config.org);
    return config.org+'/'+peers[0];
  };


}

angular.module('nsd.service.instructions', ['nsd.service.api'])
  .service('InstructionService', InstructionService);
