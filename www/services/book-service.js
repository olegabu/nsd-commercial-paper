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
    var chaincodeID = ConfigLoader.get()['contracts'].book;
    if(!chaincodeID){
      // must be specified in network-config.json
      throw new Error("No chaincode name for 'book' contract");
    }
    return chaincodeID;
  };

  BookService._getChannelID = function() {
    return 'depository';
  };



  /**
   *
   */
  BookService.list = function() {
    $log.debug('BookService.list');

    var chaincodeID = BookService._getChaincodeID();
    var channelID = BookService._getChannelID();
    var peer = BookService._getQueryPeer();

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
  BookService._getQueryPeer = function() {
    var config = ConfigLoader.get();
    var peers = ConfigLoader.getOrgPeerIds(config.org);
    return config.org+'/'+peers[0];
  };


  return BookService;
}

angular.module('nsd.service.book', ['nsd.service.api'])
  .service('BookService', BookService);
