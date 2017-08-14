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
    $log.debug('InstructionService.list');

    var chaincodeID = InstructionService._getChaincodeID();
    var peer = InstructionService._getQueryPeer();

    return ApiService.channels.list().then(function(channelList){
      return $q.all( channelList
        .map(function(channel){ return channel.channel_id; })
        .filter(function(channelID){ return _isBilateralChannel(channelID); })
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
            })

      }));
    }).then(function(results){
      // join array of array into one array (flatten)
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

  function _isBilateralChannel(channelID){
    return channelID.indexOf('-') > 0 && !channelID.startsWith('nsd-');
  }



  /**
   *
   */
  InstructionService.transfer = function(instruction) {
    $log.debug('InstructionService.transfer', instruction);

    var chaincodeID = InstructionService._getChaincodeID();
    var opponent    = InstructionService._getOpponentID(instruction);
    var channelID   = InstructionService._getOpponentChannel(opponent);
    var peers       = InstructionService._getEndorsePeers(opponent);
    var args        = InstructionService._instructionArguments(instruction);


    return ApiService.sc.invoke(channelID, chaincodeID, peers, 'transfer', args);
  };

  /**
   *
   */
  InstructionService.receive = function(instruction) {
    $log.debug('InstructionService.receive', instruction);

    var chaincodeID = InstructionService._getChaincodeID();
    var opponent    = InstructionService._getOpponentID(instruction);
    var channelID   = InstructionService._getOpponentChannel(opponent);
    var peers       = InstructionService._getEndorsePeers(opponent);
    var args        = InstructionService._instructionArguments(instruction);


    return ApiService.sc.invoke(channelID, chaincodeID, peers, 'receive', args);
  };


  /**
   *
   */
  InstructionService._instructionArguments = function(instruction) {
    return [
      instruction.transferer.dep,     // 0: deponentFrom
      instruction.transferer.acc,     // 1: accountFrom
      instruction.transferer.div,     // 2: divisionFrom

      instruction.receiver.dep,       // 3: deponentTo
      instruction.receiver.acc,       // 4: accountTo
      instruction.receiver.div,       // 5: divisionTo

      instruction.security,           // 6: security
      ''+instruction.quantity,        // 7: quantity // TODO: fix: string parameters
      instruction.reference,          // 8: reference
      instruction.instruction_date,   // 9: instructionDate  (date format?)
      instruction.trade_date,         // 10: tradeDate  (date format?)
      JSON.stringify(instruction.reason)  // 11: reason (TODO: complex field)
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


}

angular.module('nsd.service.instructions', ['nsd.service.api'])
  .service('InstructionService', InstructionService);
