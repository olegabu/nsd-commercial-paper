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
        list.forEach(BookService._processBookItem);
        return list;
      });
  };


  // add 'org' and 'deponent' to the result, based on account+division
  BookService._processBookItem = function(book){
    book.org = ConfigLoader.getOrgByAccountDivision(book.balance.account, book.balance.division);
    book.deponent = (ConfigLoader.getAccount(book.org) || {}).dep;
  }


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

  }


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

    var chaincodeID = BookService._getChaincodeID();
    var channelID = BookService.getChannelID();
    var peer = BookService._getQueryPeer();
    var args = BookService._arguments(book);
    var bookKey = BookService._bookKey(book);

    return ApiService.sc.query(channelID, chaincodeID, peer, 'history', args)
      .then(function(result){ return result.result; })
      .then(function(list){
        return list.map(function(singleValue){
          return Object.assign(singleValue.value, bookKey, {_created:new Date(singleValue.timestamp) });
        });
      })
      .then(function(list){
        list.forEach(BookService._processBookItem);
        return list;
      });
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
