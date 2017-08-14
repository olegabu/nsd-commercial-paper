/**
 * @class BalanceService
 * @classdesc
 * @ngInject
 */
function BalanceService(ApiService, ConfigLoader, $q, $log) {

  // jshint shadow: true
  var BalanceService = this;


  var CHAINCODE_ID = 'position';

  /**
   *
   */
  BalanceService._getChaincodeID = function() {
    return CHAINCODE_ID;
  };


  BalanceService._getChannelID = function() {
    // TODO: 'nsd' hardcoded
    return 'nsd-'+ConfigLoader.getOrg();
  };

  /**
   *
   */
  BalanceService.list = function() {
    $log.debug('BalanceService.list');

    var chaincodeID = BalanceService._getChaincodeID();
    var channelID = BalanceService._getChannelID();
    var peer = BalanceService._getQueryPeer();

    return ApiService.sc.query(channelID, chaincodeID, peer, 'query')
        .then(function(data){ return parseJson(data.result); });
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


  /**
   *
   */
  BalanceService._getQueryPeer = function() {
    var config = ConfigLoader.get();
    var peers = ConfigLoader.getOrgPeerIds(config.org);
    return config.org+'/'+peers[0];
  };


  return BalanceService;
}

angular.module('nsd.service.balance', ['nsd.service.api'])
  .service('BalanceService', BalanceService);
