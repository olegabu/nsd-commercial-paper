/* globals angular */
/* jshint -W069 */

/**
 * @typedef {object} Instruction
 *
 * key properties:
 *
 * @property {object} transferer
 * @property {string} transferer.deponent
 * @property {string} transferer.account
 * @property {string} transferer.division
 *
 * @property {object} receiver
 * @property {string} receiver.deponent
 * @property {string} receiver.account
 * @property {string} receiver.division
 *
 * @property {string} security
 * @property {number} quantity
 * @property {number} reference
 * @property {Date} instructionDate
 * @property {Date} tradeDate
 *
 * @property {InstructionService.type} type ('fop'|'dvp')

 * @property {object} [transfererRequisites]
 * @property {string} [transfererRequisites.account]
 * @property {string} [transfererRequisites.bic]
 * @property {object} [receiverRequisites]
 * @property {string} [receiverRequisites.account]
 * @property {string} [receiverRequisites.bic]
 *
 * @property {string} [paymentAmount]
 * @property {'RUB'}  [paymentCurrency]
 *
 *
 * extra properties:
 *
 * @property {InstructionService.status} status
 * @property {'transferer'|'receiver'} initiator
 *
 *
 * @property {string} memberInstructionId - ???
 * @property {string} memberInstructionIdFrom - ???
 * @property {string} memberInstructionIdTo - ???
 *
 * @property {string} deponentFrom
 * @property {string} deponentTo
 *
 * @property {string} alamedaFrom - xml
 * @property {string} alamedaSignatureFrom
 *
 * @property {string} alamedaTo - xml
 * @property {string} alamedaSignatureTo
 *
 *
 * @property {object} reasonFrom
 * @property {object} reasonFrom.created
 * @property {object} reasonFrom.description
 * @property {object} reasonFrom.document
 *
 * @property {string} reasonTo
 * @property {object} reasonTo.created
 * @property {object} reasonTo.description
 * @property {object} reasonTo.document
 *
 *
 * @property {string} [additionalInformation] for 16/3 only
 * @property {object} [additionalInformation.created]
 * @property {object} [additionalInformation.description]
 * @property {object} [additionalInformation.document]
 */

/**
 * @param {ApiService} ApiService
 * @param {ConfigLoader} ConfigLoader
 * @param $q
 * @param $log
 * @constructor
 *
 * @class InstructionService
 * @ngInject
 */
function InstructionService(ApiService, ConfigLoader, $q, $log) {
  "use strict";

  var InstructionService = this;

  /**
   * Enum instruction statuses
   * @enum {string}
   */
  InstructionService.status = {
      MATCHED : 'matched',
      DECLINED: 'declined',
      EXECUTED: 'executed',
      CANCELED: 'canceled',
      DOWNLOADED: 'downloaded',
      ROLLBACK_INITIATED: 'rollbackInitiated',
      ROLLBACK_DONE: 'rollbackDone',
      ROLLBACK_FAILED: 'rollbackDeclined'
      // 'transferer-signed'
      // 'receiver-signed'
  };

  /**
   * Enum instruction types
   * @enum {string}
   */
  InstructionService.type = {
    /**
     * free of payment
     */
    FOP: 'fop',
    /**
     * Delivery versus payment
     */
    DVP: 'dvp'
  };

  /**
   *
   */
  InstructionService._getChaincodeID = function() {
    var chaincodeID = ConfigLoader.get()['contracts'].instruction;
    if(!chaincodeID){
      // must be specified in network-config.json
      throw new Error("No chaincode name for 'instruction' contract");
    }
    return chaincodeID;
  };

  /**
   *
   */
  InstructionService.listAll = function() {
    $log.debug('InstructionService.listAll');

    var chaincodeID = InstructionService._getChaincodeID();
    var peer = InstructionService._getQueryPeer();

    return ApiService.channels.list().then(function(channelList){
      return $q.all( channelList
        .map(function(channel){ return channel.channel_id; })
        .filter(function(channelID){ return InstructionService.isBilateralChannel(channelID); })
        .sort()
        .map(function(channelID){
          // promise for each channel:
          return ApiService.sc.query(channelID, chaincodeID, peer, 'query')
            .then(function(data){ return {
                channel: channelID,
                result: data.result
              };
            }).catch(function(){
              return {
                channel: channelID,
                result: []
              };
            });

      }));
    })
    .then(function(results){
      // join array of results into one array (groupedList)
      return results.reduce(function(result, singleResult){
        result[singleResult.channel] = singleResult.result;
        return result;
      }, {});
    })
    .then(function(groupedList){
      // flattern: combine all group element into single array
      return Object.keys(groupedList).reduce(function(result, channel){
        result.push.apply(result, groupedList[channel]);
        return result;
      }, []);
    })
    .then(function(results){
      // join key and value
      return results.map(function(singleResult){
        return Object.assign({}, _fixStatus(singleResult.value), singleResult.key);
      });
    });
  };

  /**
   *
   */
  function _fixStatus(instruction){
    var signedFrom = (instruction.alamedaSignatureFrom && instruction.alamedaSignatureFrom.length > 0 );
    var signedTo = (instruction.alamedaSignatureTo && instruction.alamedaSignatureTo.length > 0 );

    // jshint -W016
    if( signedFrom ^ signedTo ){ // xor
      instruction.status = signedFrom ? 'transferer-signed' : 'receiver-signed';
    }
    return instruction;
  }


  /**
   * Determine whether it's a channel between two members (and nsd is always here).
   * Actually, should be called "threeLateral"
   * @return {boolean}
   */
  InstructionService.isBilateralChannel = function(channelID){
    return channelID.indexOf('-') > 0 && !channelID.startsWith('nsd-');
  };



  /**
   *
   */
  InstructionService.transfer = function(instruction) {
    $log.debug('InstructionService.transfer', instruction);

    var chaincodeID = InstructionService._getChaincodeID();
    var channelID   = InstructionService._getInstructionChannel(instruction);
    var peers       = InstructionService._getEndorsePeers(instruction);
    var args        = InstructionService._instructionArguments(instruction);

    args.push(
      instruction.deponentFrom,
      instruction.deponentTo,
      instruction.memberInstructionId,
      JSON.stringify(instruction.reason||{})
    );

    return ApiService.sc.invoke(channelID, chaincodeID, peers, 'transfer', args);
  };

  /**
   * @param {Instruction} instruction
   */
  InstructionService.receive = function(instruction) {
    $log.debug('InstructionService.receive', instruction);

    var chaincodeID = InstructionService._getChaincodeID();
    var channelID   = InstructionService._getInstructionChannel(instruction);
    var peers       = InstructionService._getEndorsePeers(instruction);
    var args        = InstructionService._instructionArguments(instruction);

    args.push(
      instruction.deponentFrom,
      instruction.deponentTo,
      instruction.memberInstructionId,
      JSON.stringify(instruction.reason||{})
    );

    // only for receiver!
    if (instruction.type === 'dvp') {
      args.push(
       JSON.stringify(instruction.additionalInformation||{})
      );
    }

    return ApiService.sc.invoke(channelID, chaincodeID, peers, 'receive', args);
  };

  /**
   *
   */
  InstructionService.rollbackInstruction = function(instruction, reason) {
    return this.updateStatus(instruction, InstructionService.status.ROLLBACK_INITIATED, reason);
  };

  InstructionService.cancelInstruction = function(instruction) {
    return this.updateStatus(instruction, InstructionService.status.CANCELED);
  };

  InstructionService.setDownloaded = function(instruction) {
    return this.updateStatus(instruction, InstructionService.status.DOWNLOADED);
  };

  /**
   * @param {Instruction} instruction
   * @param {string} status
   * @param {string} [reason]
   */
  InstructionService.updateStatus = function(instruction, status, reason) {
    reason = reason || '';
    $log.debug('InstructionService.updateStatus', instruction, status, reason);

    var chaincodeID = InstructionService._getChaincodeID();
    var channelID   = InstructionService._getInstructionChannel(instruction);
    var peers       = InstructionService._getEndorsePeers(instruction);
    var args        = InstructionService._instructionArguments(instruction);

    args.push(reason, status);

    return ApiService.sc.invoke(channelID, chaincodeID, peers, 'status', args);
  };

  /**
   * @param {Instruction} instruction
   * @param {'receiver'|'transferer'} transfererOrReceiver
   */
  InstructionService.updateDownloadFlags = function(instruction, transfererOrReceiver) {
    $log.debug('InstructionService.updateDownloadFlags', instruction, transfererOrReceiver);

    var chaincodeID = InstructionService._getChaincodeID();
    var channelID   = InstructionService._getInstructionChannel(instruction);
    var peers       = InstructionService._getEndorsePeers(instruction);
    var args        = InstructionService._instructionArguments(instruction);

    args.push(transfererOrReceiver);

    return ApiService.sc.invoke(channelID, chaincodeID, peers, 'updateDownloadFlags', args);
  };




  /**
   * @param {Instruction} instruction
   * @param {string} signature
   */
  InstructionService.sign = function(instruction, signature) {
    $log.debug('InstructionService.sign', instruction, signature);

    var chaincodeID = InstructionService._getChaincodeID();
    var channelID   = InstructionService._getInstructionChannel(instruction);
    var peers       = InstructionService._getEndorsePeers(instruction);
    var args        = InstructionService._instructionArguments(instruction);

    args.push(signature);

    return ApiService.sc.invoke(channelID, chaincodeID, peers, 'sign', args);
  };




  /**
   *
   */
  InstructionService.history = function(instruction){
    $log.debug('InstructionService.history', instruction);

    var chaincodeID = InstructionService._getChaincodeID();
    var channelID   = InstructionService._getInstructionChannel(instruction);
    var peer        = InstructionService._getQueryPeer();
    var args        = InstructionService._instructionArguments(instruction);
    var instructionKey = InstructionService._instructionKey(instruction);

    return ApiService.sc.query(channelID, chaincodeID, peer, 'history', args)
      .then(function(result){ return result.result; })
      .then(function(result){
        // get pure value
        return result.map(function(singleValue){
          return Object.assign( _fixStatus(singleValue.value), instructionKey, {_created: parseDate(singleValue.timestamp) });
        });
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
   * return basic fields for any instruction request
   * @param {Instruction} instruction
   * @return {Array<string>}
   */
  InstructionService._instructionArguments = function(instruction) {
    var args = [
      instruction.transferer.account,  // accountFrom
      instruction.transferer.division, // divisionFrom

      instruction.receiver.account,    // accountTo
      instruction.receiver.division,   // divisionTo

      instruction.security,            // security
      instruction.quantity,            // quantity // TODO: fix: string parameters
      instruction.reference,           // reference
      instruction.instructionDate,     // instructionDate  (ISO)
      instruction.tradeDate,           // tradeDate  (ISO)

      instruction.type                 // instruction type
    ];

    if (instruction.type === 'dvp') {
      args.push.apply(args, [
        instruction.transfererRequisites.account,
        instruction.transfererRequisites.bic,
        instruction.receiverRequisites.account,
        instruction.receiverRequisites.bic,
        instruction.paymentAmount,
        instruction.paymentCurrency
      ]);
    }
    return args;
  };


  /**
   * return basic fields for any instruction request
   * @return {Array<string>}
   */
  InstructionService._instructionKey = function(instruction) {
    return {
      transferer:{
        account  : instruction.transferer.account,
        division : instruction.transferer.division
      },
      receiver:{
        account  : instruction.receiver.account,
        division : instruction.receiver.division
      },
      security  : instruction.security,
      quantity  : instruction.quantity,
      reference : instruction.reference,
      instructionDate : instruction.instructionDate,
      tradeDate : instruction.tradeDate
    };
  };


  /**
   * get instruction opponents ID.
   * @param {Instruction} instruction
   * @return {Array<string>} multiple (two actually) orgID
   */
  InstructionService._getInstructionOrgs = function(instruction) {
    var org1 = ConfigLoader.getOrgByDepcode(instruction.deponentFrom);
    if(!org1){
      throw new Error("Deponent owner not found: " + instruction.deponentFrom);
    }
    var org2 = ConfigLoader.getOrgByDepcode(instruction.deponentTo);
    if(!org2){
      throw new Error("Deponent owner not found: " + instruction.deponentTo);
    }
    return [org1, org2];
  };

  /**
   * get name of bi-lateral channel for opponent and the organisation
   */
  InstructionService._getInstructionChannel = function(instruction) {
    var orgArr = InstructionService._getInstructionOrgs(instruction);
    // make channel name as '<org1_ID>-<org2_ID>'.
    // Please, pay attention to the orgs order - ot should be sorted
    return orgArr.sort().join('-');
  };

  /**
   * get orgPeerIDs of endorsers, which should endose the transaction
   * @return {Array<string>}
   */
  InstructionService._getEndorsePeers = function(instruction) {
    var endorserOrgs = InstructionService._getInstructionOrgs(instruction);

    // root endorser
    var config = ConfigLoader.get();
    var rootEndorsers = config.endorsers || [];
    endorserOrgs.push.apply(endorserOrgs, rootEndorsers);

    //
    var peers = endorserOrgs.reduce(function(result, org){
      var peers = ConfigLoader.getOrgPeerIds(org);
      result.push( org+'/'+peers[0] ); // orgPeerID  // endorse by the first peer
      return result;
    }, []);

    return peers;
  };


  /**
   *
   */
  InstructionService._getQueryPeer = function() {
    var config = ConfigLoader.get();
    var peers = ConfigLoader.getOrgPeerIds(config.org);
    return config.org+'/'+peers[0];
  };


  // add 'org' and 'deponent' to the result, based on account+division
  InstructionService._processItem = function(instruction){
    instruction._orgFrom = ConfigLoader.getOrgByAccountDivision(instruction.transferer.account, instruction.transferer.division);
    instruction._orgTo = ConfigLoader.getOrgByAccountDivision(instruction.receiver.account, instruction.receiver.division);
    instruction.deponentFrom = (ConfigLoader.getAccount(instruction._orgFrom) || {}).dep;
    instruction.deponentTo = (ConfigLoader.getAccount(instruction._orgTo) || {}).dep;

    if(instruction.reason){
      instruction.reason     = parseJsonSafe(instruction.reason); // for redeem instruction
    }
  };

  function parseJsonSafe(str){
    try{
      return JSON.parse(str);
    }catch(e){
      return str;
    }
  }

}

angular.module('nsd.service.instructions', ['nsd.service.api'])
  .service('InstructionService', InstructionService);
