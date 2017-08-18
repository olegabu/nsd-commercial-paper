/**
 * @class BookService
 * @classdesc
 * @ngInject
 */
function BookService(ApiService, ConfigLoader, $q, $log) {

  // jshint shadow: true
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
        // add 'org' and 'deponent' to the result, based on account+division
        list.forEach(function(item){
          item.org = ConfigLoader.getOrgByAccountDivision(item.balance.account, item.balance.division);
          item.deponent = (ConfigLoader.getAccount(item.org) || {}).dep;
        })
        return list;
      });
  };



  /**
   *
   */
  BookService._getQueryPeer = function() {
    var config = ConfigLoader.get();
    var peers = ConfigLoader.getOrgPeerIds(config.org);
    return config.org+'/'+peers[0];
  };



  BookService.redeem = function(redemption){
    $log.debug('BookService.redeem', redemption);

    throw new Error('Redemption is incomplete');

    // var chaincodeID = BookService._getChaincodeID();
    // var channelID = BookService.getChannelID();
    // var peer = BookService._getQueryPeer();

    // // We can safely use here the result of _getQueryPeer() fn.
    // return ApiService.sc.invoke(channelID, chaincodeID, [peer], 'redeem', redemption);
  }


  /**
   *
   */
  BookService.history = function(book){
    $log.debug('BookService.history', book);
    $log.warn('BookService.history STUB!');
    // FIXME: this is a temp measure to test ui
    return BookService.list();
  }


  return BookService;
}

angular.module('nsd.service.book', ['nsd.service.api'])
  .service('BookService', BookService);
