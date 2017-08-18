/**
 * @class PositionsService
 * @classdesc
 * @ngInject
 */
function PositionsService(ApiService, ConfigLoader, $q, $log) {

  // jshint shadow: true
  var PositionsService = this;


  var CHAINCODE_ID = 'position';

  /**
   *
   */
  PositionsService._getChaincodeID = function() {
    return CHAINCODE_ID;
  };


  PositionsService.getChannelID = function() {
    // TODO: 'nsd' hardcoded
    return 'nsd-'+ConfigLoader.getOrg();
  };

  /**
   *
   */
  PositionsService.list = function() {
    $log.debug('PositionsService.list');

    var chaincodeID = PositionsService._getChaincodeID();
    var channelID = PositionsService.getChannelID();
    var peer = PositionsService._getQueryPeer();

    return ApiService.sc.query(channelID, chaincodeID, peer, 'query')
        .then(function(data){ return data.result; });
  };


  /**
   *
   */
  PositionsService.history = function(book){
    $log.debug('PositionsService.history', book);
    $log.warn('PositionsService.history STUB!');
    // FIXME: this is a temp measure to test ui
    return PositionsService.list();
  }

  /**
   *
   */
  PositionsService._getQueryPeer = function() {
    var config = ConfigLoader.get();
    var peers = ConfigLoader.getOrgPeerIds(config.org);
    return config.org+'/'+peers[0];
  };


  return PositionsService;
}

angular.module('nsd.service.positions', ['nsd.service.api'])
  .service('PositionsService', PositionsService);
