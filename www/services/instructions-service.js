/**
 * @class InstructionService
 * @classdesc
 * @ngInject
 */
function InstructionService(ApiService, ConfigLoader, $q, $log) {

  // jshint shadow: true
  var InstructionService = this;

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
  InstructionService.list = function() {

    var chaincodeID = InstructionService._getChaincodeID();
    var peer = InstructionService._getQueryPeer();

    return ApiService.channels.list().then(function(list){
      return $q.all( list.map(function(channel){
        // promise for each channel:
        return ApiService.sc.query(channel.channel_id, chaincodeID, peer, 'query')
            .then(function(data){ return parseJson(data.result); });

      }));
    }).then(function(results){
      // join array of array into one array (flatten)
      return results.reduce(function(result, singleResult){
        result.push.apply(result, singleResult);
        return result;
      }, []);
    });
  };

  /**
   *
   */
  function parseJson(data){
    var parsed = null;
    try{
      parsed = JSON.parse(data);
    }catch(e){
      $log.warn(e);
    }
    return parsed;
  }



  /**
   *
   */
  InstructionService.transfer = function(instruction) {
    $log.info('InstructionService.transfer', instruction);

    var chaincodeID = InstructionService._getChaincodeID();
    var opponent    = InstructionService._getOpponentID(instruction);
    var channelID   = InstructionService._getOpponentChannel(opponent);
    var peers       = InstructionService._getEndorsePeers(opponent);
    var args        = InstructionService._instructionToArguments(instruction);


    return ApiService.sc.invoke(channelID, chaincodeID, peers, 'transfer', args);
  };

  /**
   *
   */
  InstructionService.receive = function(instruction) {
    $log.info('InstructionService.receive', instruction);

    var chaincodeID = InstructionService._getChaincodeID();
    var opponent    = InstructionService._getOpponentID(instruction);
    var channelID   = InstructionService._getOpponentChannel(opponent);
    var peers       = InstructionService._getEndorsePeers(opponent);
    var args        = InstructionService._instructionFromArguments(instruction);


    return ApiService.sc.invoke(channelID, chaincodeID, peers, 'receive', args);
  };


  /**
   *
   */
  InstructionService._instructionToArguments = function(instruction) {
    return [
      // no deponentFrom here
      instruction.transferer.acc,     // 0: accountFrom
      instruction.transferer.div,     // 1: divisionFrom

      instruction.receiver.dep,       // 2: deponentTo
      instruction.receiver.acc,       // 3: accountTo
      instruction.receiver.div,       // 4: divisionTo

      instruction.security,           // 5: security
      ''+instruction.quantity,           // 6: quantity // TODO: fix: string parameters
      instruction.reference,          // 7: reference
      instruction.instruction_date,   // 8: instructionDate  (date format?)
      instruction.trade_date,         // 9: tradeDate  (date format?)
      JSON.stringify(instruction.reason)              // 10: reason (TODO: complex field)
    ];
  }
  /**
   *
   */
  InstructionService._instructionFromArguments = function(instruction) {
    return [
      // no deponentFrom here
      instruction.transferer.dep,     // 0: deponentFrom
      instruction.transferer.acc,     // 1: accountFrom
      instruction.transferer.div,     // 2: divisionFrom

      instruction.receiver.acc,       // 3: accountTo
      instruction.receiver.div,       // 4: divisionTo

      instruction.security,           // 5: security
      ''+instruction.quantity,           // 6: quantity // TODO: fix: string parameters
      instruction.reference,          // 7: reference
      instruction.instruction_date,   // 8: instructionDate  (date format?)
      instruction.trade_date,         // 9: tradeDate  (date format?)
      JSON.stringify(instruction.reason)              // 10: reason (TODO: complex field)
    ];
  }



  /**
   * get instruction opponent ID.
   * It would be transfererID when you are receiver and vise a versa
   * @param {Instruction} instruction
   * @return {string} orgID
   */
  InstructionService._getOpponentID = function(instruction) {
    var opponentDep = instruction.side == 'transferer' ? instruction.receiver.dep : instruction.transferer.dep;
    if(!opponentDep){
      throw new Error("Deponent not set");
    }
    var opponent = ConfigLoader.getOrgByDepcode(opponentDep);
    if(!opponent){
      throw new Error("Deponent owner not found: "+opponentDep);
    }
    return opponent;
  }

  /**
   * get name of bi-lateral channel for opponent and the organisation
   */
  InstructionService._getOpponentChannel = function(opponent) {
    // make channel name as '<org1_ID>-<org2_ID>'.
    // Please, pay attention to the orgs order - ot should be sorted
    return [ConfigLoader.getOrg(), opponent].sort().join('-');
  };

  /**
   * get orgPeerIDs of endorsers, which should endose the transaction
   */
  InstructionService._getEndorsePeers = function(opponent) {
    var config = ConfigLoader.get();

    var endorsersOrg = [config.org, opponent];

    var rootEndorsers = config.endorsers || [];
    endorsersOrg.push.apply(endorsersOrg, rootEndorsers);

    //
    var peers = endorsersOrg.reduce(function(result, org){
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


  /**
   */
  InstructionService.receive = function(instruction) {
    $log.info('InstructionService.receive', instruction);
    return $q.resolve();
    // return ApiService.sc.invoke(channelID, chaincodeID, peers, 'list');
  };

}

angular.module('nsd.service.instructions', ['nsd.service.api'])
  .service('InstructionService', InstructionService);
