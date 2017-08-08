/**
 * @class InstructionService
 * @classdesc
 * @ngInject
 */
function InstructionService(ApiService, ConfigLoader, $q, $log) {

  // jshint shadow: true
  var InstructionService = this;

  // TODO: moce to config/settings
  var chaincodeID = 'mycc';

  var ROOT_ENDORSER = 'nsd';

  /**
   */
  InstructionService.list = function() {
    // return $q.resolve(getSampleList());
    return ApiService.channels.list().then(function(list){
      var peer = InstructionService._getQueryPeer();

      return $q.all(
        list.map(function(channel){
          return ApiService.sc.query(channel.channel_id, chaincodeID, peer, 'query');
        })
      );
    }).then(function(results){
      console.log('results', results);
      return results.reduce(function(result, singleResult){
        try{
          var items = JSON.parse(singleResult.result);
          result.push.apply(result, items);
        }catch(e){
          $log.warn(e);
        }
        return result;
      }, []);
    });
  };

  /**
   */
  InstructionService.transfer = function(instruction) {
    $log.info('InstructionService.transfer', instruction);

    var channelID = InstructionService._getChannelForInstruction(instruction);
    var peers     = InstructionService._getPeersForInstruction(instruction);
    var args      = InstructionService._instructionToArguments(instruction);


    return ApiService.sc.invoke(channelID, chaincodeID, peers, 'transfer', args);
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
   * get name of bi-lateral channel for instruction.transfer and instruction.receiver
   */
  InstructionService._getChannelForInstruction = function(instruction) {
    var opponent = InstructionService._getOpponentID(instruction);
    return [ConfigLoader.getOrg(), opponent].sort().join('-');
  };

  /**
   *
   */
  InstructionService._getPeersForInstruction = function(instruction) {
    var config = ConfigLoader.get();

    var opponent = InstructionService._getOpponentID(instruction);
    var endorsersOrg = [config.org, opponent, ROOT_ENDORSER];

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




function getSampleList(){
  return [
    {
      transferer  : {
          dep: "ASDASDASD",
          acc: "123123213",
          div: "35"
      },
      receiver    : {
          dep: "zxczxczxc",
          acc: "798798798",
          div: "78"
      },
      security    : "US0378331005", // ISIN
      quantity    : 100500,
      reference   : "This is a first transaction",
      instruction_date     : Date.now(),
      trade_date  : Date.now(),
      status      : 'initiated',
      side        : 'transferer',

      authority : {
          document: "a-b/07",
          description: "authority description",
          created: Date.now()
      }
    }
  ];
}