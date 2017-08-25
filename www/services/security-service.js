/**
 * @class SecurityService
 * @classdesc
 * @ngInject
 */
function SecurityService(ApiService, ConfigLoader, $q, $log) {

  // jshint shadow: true
  var SecurityService = this;


  SecurityService.STATUS_ACTIVE = 'active';

  /**
   *
   */
  SecurityService._getChaincodeID = function() {

    return 'security';

    // var chaincodeID = ConfigLoader.get()['contracts'].securityMaster;
    // if(!chaincodeID){
    //   // must be specified in network-config.json
    //   throw new Error("No chaincode name for 'securityMaster' contract");
    // }
    // return chaincodeID;
  };

  SecurityService.getChannelID = function() {
    return 'common';
  };



  /**
   * @param {string} [status] - security status to fetch
   */
  SecurityService.list = function(status) {
    $log.debug('SecurityService.list');

    var chaincodeID = SecurityService._getChaincodeID();
    var channelID = SecurityService.getChannelID();
    var peer = SecurityService._getQueryPeer();

    return ApiService.sc.query(channelID, chaincodeID, peer, 'query')
        .then(function(data){ return data.result; })
        .then(function(list){
          if(status){
            return list.filter(function(s){ return s.status == status; });
          } else {
            return list;
          }
        });
  };

  /**
   *
   */
  SecurityService.addCalendarEntry = function(cEntry) {
    $log.debug('SecurityService.addCalendarEntry');

    throw new Error("addCalendarEntry incomplete");
    // var chaincodeID = SecurityService._getChaincodeID();
    // var channelID = SecurityService.getChannelID();
    // var peer = SecurityService._getQueryPeer();

    // return ApiService.sc.query(channelID, chaincodeID, peer, 'query')
    //     .then(function(data){ return parseJson(data.result); });
  };

  SecurityService.sendSecurity = function(security){
    $log.debug('SecurityService.sendSecurity', security);

    var chaincodeID = SecurityService._getChaincodeID();
    var channelID = SecurityService.getChannelID();
    var peer = SecurityService._getQueryPeer();
    var args = [
      security.security,
      SecurityService.STATUS_ACTIVE,
      security.redeem.account,
      security.redeem.division
    ];

    // We can safely use here the result of _getQueryPeer() fn.
    return ApiService.sc.invoke(channelID, chaincodeID, [peer], 'put', args);
  };


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
