/* globals angular */

/**
 * @typedef {object} Book
 * @property {object} balance
 * @property {string} balance.account
 * @property {string} balance.division
 * @property {string} security
 * @property {'money'|'paper'} type
 */



/**
 * @param {ApiService} ApiService
 * @param {ConfigLoader} ConfigLoader
 * @param {UserService} UserService
 * @param {InstructionService} InstructionService
 * @param $q
 * @param $log
 *
 * @return {BookService}
 * @constructor
 *
 * @class BookService
 * @ngInject
 */
function BookService(ApiService, ConfigLoader, UserService, InstructionService, $q, $log) {
  "use strict";

  var BookService = this;

  /**
   *
   */
  BookService._getChaincodeID = function() {
    return "book";
  };

  BookService.getChannelID = function() {
    return 'depository';
  };



  /**
   *
   */
  BookService.list = function() {
    $log.debug('BookService.list');

    var chaincodeID = BookService._getChaincodeID();
    var channelID = BookService.getChannelID();
    var peer = BookService._getQueryPeer();

    return ApiService.sc.query(channelID, chaincodeID, peer, 'query')
      .then(function(data){ return data.result; })
      .then(function(list){
        list.forEach(BookService._processBookItem);
        return list;
      })
      .then(function(list){
        return list.filter(function(book){ return book.quantity != 0; });
      });
  };


  // add 'org' and 'deponent' to the result, based on account+division
  BookService._processBookItem = function(book){
    book.org = ConfigLoader.getOrgByAccountDivision(book.balance.account, book.balance.division);
    book.deponent = (ConfigLoader.getAccount(book.org) || {}).dep;
    book.type = book.security.length > 3 ? 'paper' : 'money';
  };


  /**
   *
   */
  BookService._getQueryPeer = function() {
    var config = ConfigLoader.get();
    var peers = ConfigLoader.getOrgPeerIds(config.org);
    return config.org+'/'+peers[0];
  };


  /**
   *
   */
  BookService.put = function(book){
    $log.debug('BookService.put', book);

    var chaincodeID = BookService._getChaincodeID();
    var channelID = BookService.getChannelID();
    var peer = BookService._getQueryPeer();
    var args = BookService._arguments(book);
    args.push(book.quantity);

    // We can safely use here the result of _getQueryPeer() fn.
    return ApiService.sc.invoke(channelID, chaincodeID, [peer], 'put', args);

  };


  BookService.redeem = function(redemption){
    $log.debug('BookService.redeem', redemption);

    var chaincodeID = BookService._getChaincodeID();
    var channelID = BookService.getChannelID();
    var peer = BookService._getQueryPeer();
    var args = [
      redemption.security,
      JSON.stringify(redemption.reason||{})
    ];

    // We can safely use here the result of _getQueryPeer() fn.
    return ApiService.sc.invoke(channelID, chaincodeID, [peer], 'redeem', args);
  };


  /**
   *
   */
  BookService.history = function(book){
    $log.debug('BookService.history', book);

    var chaincodeID = BookService._getChaincodeID();
    var channelID = BookService.getChannelID();
    var peer = BookService._getQueryPeer();
    var args = BookService._arguments(book);
    var bookKey = BookService._bookKey(book);

    return ApiService.sc.query(channelID, chaincodeID, peer, 'history', args)
      .then(function(result){ return result.result; })
      .then(function(list){
        return list.map(function(singleValue){
          return Object.assign(singleValue.value, bookKey, {_created:parseDate(singleValue.timestamp) });
        });
      })
      .then(function(list){
        list.forEach(BookService._processBookItem);
        return list;
      });
  };

  /**
   * parse "2018-03-13 13:30:46.909727155 +0000 UTC" to date
   * @param datestr
   */
  function parseDate(datestr) {
    return new Date((datestr||'').replace(/\s*\+.+$/,'').replace(' ','T'));
  }

  /**
   * @param {string} [security]
   */
  BookService.redeemHistory = function(security){
    $log.debug('BookService.redeemHistory');

    if( UserService.getOrgRole() !== 'nsd'){
      // withRedeem = false;
      $log.warn('Role %s cannot fetch redeem history', UserService.getOrgRole());
      return $q.resolve(null);
    }

    var chaincodeID = BookService._getChaincodeID();
    var channelID = BookService.getChannelID();
    var peer = BookService._getQueryPeer();
    var args = security ? [ security ] : [];

    return ApiService.sc.query(channelID, chaincodeID, peer, 'redeemHistory', args)
      .then(function(result){ return result.result || []; }) // return null instead of empty string ""
      .then(function(redeemList){
        redeemList.forEach(function(redeem){
          redeem.instructions.forEach(InstructionService._processItem);
        });
        return redeemList;
      });
      // .then(function(list){
      //   return list.map(function(singleValue){
      //     return Object.assign(singleValue.value, bookKey, {_created:parseDate(singleValue.timestamp) });
      //   });
      // })
  };






  /**
   * return basic fields for any instruction request
   * @return {Array<string>}
   */
  BookService._bookKey = function(book) {
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
  BookService._arguments = function(book){
    return [
      book.balance.account,
      book.balance.division,
      book.security
    ];
  };

  //
  return BookService;
}

angular.module('nsd.service.book', ['nsd.service.api'])
  .service('BookService', BookService);
