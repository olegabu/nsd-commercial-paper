/**
 * @class SecurityService
 * @classdesc
 * @ngInject
 */
function SecurityService(ApiService, ConfigLoader, $q, $log) {

  // jshint shadow: true
  var SecurityService = this;

  /**
   *
   */
  SecurityService._getChaincodeID = function() {

    return 'security';

    var chaincodeID = ConfigLoader.get()['contracts'].securityMaster;
    if(!chaincodeID){
      // must be specified in network-config.json
      throw new Error("No chaincode name for 'securityMaster' contract");
    }
    return chaincodeID;
  };

  SecurityService._getChannelID = function() {
    return 'depository';
  };



  /**
   *
   */
  SecurityService.list = function() {
    $log.debug('SecurityService.list');

    var chaincodeID = SecurityService._getChaincodeID();
    var channelID = SecurityService._getChannelID();
    var peer = SecurityService._getQueryPeer();

    return ApiService.sc.query(channelID, chaincodeID, peer, 'query')
        .then(function(data){ return parseJson(data.result); });
  };

  /**
   *
   */
  function parseJson(data){
    var parsed = null;
    try{
      parsed = JSON.parse(data);
    }catch(e){
      $log.warn(e, data);
    }
    return parsed;
  }


  /**
   *
   */
  SecurityService._getQueryPeer = function() {
    var config = ConfigLoader.get();
    var peers = ConfigLoader.getOrgPeerIds(config.org);
    return config.org+'/'+peers[0];
  };


  return SecurityService;
}

angular.module('nsd.service.security', ['nsd.service.api'])
  .service('SecurityService', SecurityService);
