/**
 * @class SecurityService
 * @classdesc
 * @ngInject
 */
function SecurityService(ApiService, ConfigLoader, BookService, $q, $log) {
  "use strict";

  /**
   * @typedef {object} Security
   * @property {string} security
   * @property {'active'} status
   * @property {'money'|'paper'} type
   *
   * @property {any[]} entries
   * @property {object} redeem
   * @property {string} redeem.account
   * @property {string} redeem.division
   */

  // jshint shadow: true
  var SecurityService = this;

  /**
   * @memberOf SecurityService
   */
  SecurityService.STATUS_ACTIVE = 'active';


  /**
   * @memberOf SecurityService
   */
  SecurityService.TYPE_PAPER = 'paper';
  /**
   * @memberOf SecurityService
   */
  SecurityService.TYPE_MONEY = 'money';

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
   * @param {boolean} [withRedeem] - fetch redeem instructions for each security
   * @return {Security[]}
   */
  SecurityService.list = function(status, withRedeem) {
    $log.debug('SecurityService.list');

    var chaincodeID = SecurityService._getChaincodeID();
    var channelID = SecurityService.getChannelID();
    var peer = SecurityService._getQueryPeer();

    return ApiService.sc.query(channelID, chaincodeID, peer, 'query')
        .then(function(data){ return data.result; })
        .then(function(list){
          if (status) {
            return list.filter(function(s){ return s.status === status; });
          } else {
            return list;
          }
        })
        .then(function(list){
          // fill type
          list.forEach(function(item){
            item.type = item.security.length > 3 ? SecurityService.TYPE_PAPER : SecurityService.TYPE_MONEY;
          });
          return list;
        })
        .then(function(list){
          if(!withRedeem) {
            return list;
          }

          return $q.all(list.map(function(security){
            return BookService.redeemHistory(security.security)
              .then(function(redeemList){
                security.redeem = redeemList;
                return security;
              });
          }))
        });
  };

  /**
   *
   */
  SecurityService.addCalendarEntry = function(cEntry) {
    $log.debug('SecurityService.addCalendarEntry');

    var chaincodeID = SecurityService._getChaincodeID();
    var channelID = SecurityService.getChannelID();
    var peer = SecurityService._getQueryPeer();
    var args = [
      cEntry.security,
      cEntry.type,
      cEntry.date,
      cEntry.description||'',
      cEntry.reference
    ];

    // We can safely use here the result of _getQueryPeer() fn.
    return ApiService.sc.invoke(channelID, chaincodeID, [peer], 'addEntry', args);
  };

  SecurityService.sendSecurity = function(security){
    $log.debug('SecurityService.sendSecurity', security);

    var chaincodeID = SecurityService._getChaincodeID();
    var channelID = SecurityService.getChannelID();
    var peer = SecurityService._getQueryPeer();
    var args = [
      security.security.toUpperCase(),
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
