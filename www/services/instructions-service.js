/**
 * @class InstructionService
 * @classdesc
 * @ngInject
 */
function InstructionService(ApiService, ConfigLoader, $q, $log) {

  // jshint shadow: true
  var InstructionService = this;

  // TODO: moce to config/settings
  var channelID = 'mychannel';
  var chaincodeID = 'mycc';
  var peers = ['org1/peer0'];

  var ROOT_ENDORSER = 'nsd';

  /**
   */
  InstructionService.list = function() {
    return $q.resolve(getSampleList());
    // return ApiService.sc.invoke(channelID, chaincodeID, peers, 'list');
  };

  /**
   */
  InstructionService.transfer = function(instruction) {
    $log.info('InstructionService.transfer', instruction);

    var channelID = InstructionService._getChannelForInstruction(instruction);
    var peers = InstructionService._getPeersForInstruction(instruction);

    var args = [
      // no deponentFrom here
      instruction.transferer.acc,     // 0: accountFrom
      instruction.transferer.div,     // 1: divisionFrom

      instruction.receiver.dep,     // 2: deponentTo
      instruction.receiver.acc,     // 3: accountTo
      instruction.receiver.div,     // 4: divisionTo

      instruction.security,           // 5: security
      instruction.quantity,           // 6: quantity
      instruction.reference,          // 7: reference
      instruction.instruction_date,   // 8: instructionDate  (date format?)
      instruction.trade_date,         // 9: tradeDate  (date format?)
      instruction.reason              // 10: reason (TODO: complex field)
    ];


    return ApiService.sc.invoke(channelID, chaincodeID, peers, 'transfer', args);
  };


  InstructionService._getOpponentID = function(instruction) {
    var opponentDep = instruction.transferer.dep || instruction.receiver.dep;
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
    return [config.org, opponent].sort().join('-');
  };

  /**
   *
   */
  InstructionService._getPeersForInstruction = function(instruction) {
    var config = ConfigLoader.get();

    var opponent = InstructionService._getOpponentID(instruction);
    var endorsersOrg = [config.org, opponent, ROOT_ENDORSER];

    var peers = endorsersOrg.reduce(function(result, org){
      var peers = ConfigLoader.getOrgPeerIds(org);
      result.push( org+'/'+peers[0] ); // orgPeerID  // endorse by the first peer
    }, []);

    return peers;

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