/* globals angular */

/**
 * @class PositionsService
 * @classdesc
 * @ngInject
 */
function PositionsService(ApiService, ConfigLoader, $q, $log) {
  "use strict";

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
        .then(function(data){ return data.result; })
        .then(function(list){
          list.forEach(PositionsService._processBookItem);
          return list;
        })
        .then(function(list){
          return list.filter(function(book){ return book.quantity != 0; });
        });

  };


  // add 'org' and 'deponent' to the result, based on account+division
  PositionsService._processBookItem = function(book){
    book.org = ConfigLoader.getOrgByAccountDivision(book.balance.account, book.balance.division);
    book.deponent = (ConfigLoader.getAccount(book.org) || {}).dep;
    book.type = book.security.length > 3 ? 'paper' : 'money';
  };

  /**
   *
   */
  PositionsService._getQueryPeer = function() {
    var config = ConfigLoader.get();
    var peers = ConfigLoader.getOrgPeerIds(config.org);
    return config.org+'/'+peers[0];
  };


  /**
   *
   */
  PositionsService.history = function(book){
    $log.debug('PositionsService.history', book);

    var chaincodeID = PositionsService._getChaincodeID();
    var channelID = PositionsService.getChannelID();
    var peer = PositionsService._getQueryPeer();
    var args = PositionsService._arguments(book);
    var bookKey = PositionsService._bookKey(book);

    return ApiService.sc.query(channelID, chaincodeID, peer, 'history', args)
      .then(function(result){ return result.result; })
      .then(function(list){
        return list.map(function(singleValue){
          return Object.assign(singleValue.value, bookKey, {_created: parseDate(singleValue.timestamp) });
        });
      })
      .then(function(list){
        list.forEach(PositionsService._processBookItem);
        return list;
      });
  };

  function parseDate(datestr){
    return new Date((datestr||'').replace(/\s*\+.+$/,''));
  }


  /**
   * return basic fields for any instruction request
   * @return {Array<string>}
   */
  PositionsService._bookKey = function(book) {
    return {
      balance:{
        account  : book.balance.account,
        division : book.balance.division
      },
      security : book.security
    };
  };


  /**
   * Get basc book arguments for all book qury/invoke requests
   * @retutn (Array<string>)
   */
  PositionsService._arguments = function(book){
    return [
      book.balance.account,
      book.balance.division,
      book.security
    ];
  };

  //
  return PositionsService;
}

angular.module('nsd.service.positions', ['nsd.service.api'])
  .service('PositionsService', PositionsService);
