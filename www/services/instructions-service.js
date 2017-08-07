/**
 * @class InstructionService
 * @classdesc
 * @ngInject
 */
function InstructionService(ApiService, $q) {

  // jshint shadow: true
  var InstructionService = this;

  // TODO: moce to config/settings
  var channelID = 'mychannel';
  var chaincodeID = 'mycc';
  var peers = ['org1/peer0'];

  /**
   */
  InstructionService.list = function() {
    return $q.resolve(getSampleList());
    // return ApiService.sc.invoke(channelID, chaincodeID, peers, 'list');
  };

  /**
   */
  InstructionService.send = function() {
    return $q.resolve();
    // return ApiService.sc.invoke(channelID, chaincodeID, peers, 'list');
  };


  /**
   */
  InstructionService.receive = function() {
    return $q.resolve();
    // return ApiService.sc.invoke(channelID, chaincodeID, peers, 'list');
  };

}

angular.module('nsd.service.instructions', ['nsd.service.api'])
  .service('InstructionService', InstructionService);




function getSampleList(){
  return [
    {
      transferer  : "Org 1",
      receiver    : "Org 2",
      security    : "US0378331005",
      quantity    : 100500,
      reference   : "This is a first transaction",
      trade_date  : Date.now(),
      created     : Date.now(),
      status      : 'approved',
      side        : 'transferer'
    }

  ];
}