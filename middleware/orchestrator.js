/**
 *
 */

const helper = require('./helper');
const ConfigHelper = require('./helper').ConfigHelper;

module.exports = function (require) {

  let log4js = require('log4js');
  let logger = log4js.getLogger('orchestrator');

  // TODO: move somewhere =)
  const ORG = process.env.ORG || null;
  if (ORG !== 'nsd') {
    logger.info('enabled for nsd only');
    return;
  }

  var tools = require('../lib/tools');
  let hfc = require('../lib-fabric/hfc.js');
  var config = hfc.getConfigSetting('config');
  var configHelper = new ConfigHelper(config);
  var networkConfig = configHelper.networkConfig;

  var endorsePeerId = Object.keys(networkConfig[ORG]||{}).filter(k=>k.startsWith('peer'))[0];
  var endorsePeerHost = tools.getHost(networkConfig[ORG][endorsePeerId].requests);

  //TODO user with login 'admin' has to be logged to the web gui, use admin identity that api server operates under
  var config = require('../config.json');
  const USERNAME = process.env.SERVICE_USER || 'service' /*config.user.username*/;


  logger.info('**************    ORCHESTRATOR     ******************');
  logger.info('Admin   \t: ' + USERNAME);
  logger.info('Org name\t: ' + ORG);
  logger.info('Endorse peer\t: (%s) %s', endorsePeerId, endorsePeerHost);
  logger.info('**************                     ******************');

  let invoke = require('../lib-fabric/invoke-transaction.js');
  let query  = require('../lib-fabric/query.js');
  let peerListener = require('../lib-fabric/peer-listener.js');


  const TYPE_ENDORSER_TRANSACTION = 'ENDORSER_TRANSACTION';

  ///////////////////////////////////////////
  /// main activity
  logger.info('registering for block events');
  peerListener.registerBlockEvent(function (block) {
    try {
      block.data.data.forEach(blockData => {

        let type = helper.getTransactionType(blockData);
        let channel = helper.getTransactionChannel(blockData);

        logger.info(`got block no. ${block.header.number}: ${type} on channel ${channel}`);

        if (type === TYPE_ENDORSER_TRANSACTION) {

          blockData.payload.data.actions.forEach(action => {
            let extension = action.payload.action.proposal_response_payload.extension;
            let event = extension.events;
            if(!event.event_name) {
              return;
            }
            logger.trace(`event ${event.event_name}`);

            if(event.event_name === 'Instruction.matched') {
              // instruction is matched, so we should move the values within 'book' cc
              var instruction = JSON.parse(event.payload.toString());
              logger.trace('Instruction.matched', JSON.stringify(instruction));
              instruction = helper.normalizeInstruction(instruction);
              moveBookByInstruction(instruction);
            }

            if(channel === 'depository' && event.event_name === 'Instruction.executed') {
              // instruction is executed, however still has 'matched' status in ledger (but 'executed' in the event)
              var instruction = JSON.parse(event.payload.toString());
              logger.trace('Instruction.executed', JSON.stringify(instruction));
              instruction = helper.normalizeInstruction(instruction);
              updateInstructionStatus(instruction, instruction.status /* 'executed' */);
            }
          }); // thru action elements


          //TODO this updates all positions on any new block on book channel. Better if this is done only on startup.
          // Book can emit move event with payload of updated Positions, then you don't have to query Book
          if(channel === 'depository') {
            updatePositionsFromBook();
          }
        }
      }); // thru block data elements
    }
    catch(e) {
      logger.error('Caught while processing block event', e);
    }
  });


  peerListener.eventHub.on('connected', function(){
    // run check on connect/reconnect, so we'll process all missed records
    _processMatchedInstructions();
  });



  // QUERY INSTRUCTIONS

  function _processMatchedInstructions(){
    logger.info('Process missed instructions');
    var INSTRUCTION_MATCHED_STATUS = 'matched';

    return _getAllInstructions(endorsePeerId/*, INSTRUCTION_MATCHED_STATUS*/) // TODO: uncomment this line when 'key' will be received
        .then(function(instructionInfoList){
          // typeof instructionInfoList is {Array<{channel_id:string, instruction:instruction}>}
          logger.debug('Got %s instruction(s) to process', instructionInfoList.length);

          return /*tools.*/chainPromise(instructionInfoList, function(instructionInfo){
            // var channelID = instructionInfo.channel_id;
            var instruction = instructionInfo.instruction;

            if(instruction.status !== INSTRUCTION_MATCHED_STATUS){
              logger.warn('Skip instruction with status "%s" (not "%s")', instruction.status, INSTRUCTION_MATCHED_STATUS);
              return;
            }

            return moveBookByInstruction(instruction)
              // already catched in 'moveBookByInstruction'
              // .catch(e=>{
              //   logger.error('_processInstruction failed:', e);
              // });
          });
        })
        .catch(e=>{
          logger.error('_processMatchedInstructions failed:', e);
        });
  }


  /**
   * @param {string} peer
   * @param {string} [status]
   * @return {Promise<Array<Instruction>>}
   */
  function _getAllInstructions(peer, status){
    // logger.trace('_getAllInstructions', peer, status);
    var self = this;
    return query.getChannels(peer, USERNAME, ORG)
        .then(result=>result.channels)
        // filter bilateral channels
        .then(channelList=>channelList.filter(channel=>helper.isBilateralChannel(channel.channel_id)))
        .then(function(channelList){
          // logger.trace('_getAllInstructions got channels:', JSON.stringify(channelList));
          return /*tools.*/chainPromise(channelList, function(channel){
              return _getInstructions(channel.channel_id, peer, status)
                .catch(e=>{
                  logger.warn(e);
                  return [];
                })
                .then(function(instructionList){
                  return instructionList.map(instruction=>{
                    return {
                      channel_id: channel.channel_id,
                      instruction : instruction
                    };
                  });
                });
          });
        })
        .then(function(dataArr){
          // join array of array into one array
          return dataArr.reduce(function(result, data){
            result.push.apply(result, data);
            return result;
          }, []);
        })
  }



  /**
   * @param {string} channelID
   * @param {string} peer - (orgPeerID)
   * @param {string} [status]
   * @return {Promise<Array<Instruction>>}
   */
  function _getInstructions(channelID, peer, status){
    // logger.trace('FabricRestClient.getInstructions', channelID, peer, status);
    var args = status ? [status] : [];
    var method = status ? 'queryByType' : 'query';

    //TODO peer0 is inconsistent with explicit peer url in invoke.invokeChaincode. This caused Oleg pain.
    return query.queryChaincode(peer, channelID, 'instruction', args, method, USERNAME, ORG)
      .then(function(response){ return response.result; })
      .then(function(results){
        // join key and value
        return results.map(function(singleResult){
          //logger.trace('FabricRestClient.getInstructions result', JSON.stringify(singleResult));
          return Object.assign({}, singleResult.key, singleResult.value);
        });
      });
  }



  /**
   *
   */
  function moveBookByInstruction(instruction) {
    logger.debug('invoking book move %s for %s', instruction.quantity, helper.instruction2string(instruction));

    //
    var args = helper.instructionArguments(instruction);
    return invoke.invokeChaincode([endorsePeerHost], 'depository', 'book', 'move', args, USERNAME, ORG)
      .then(function (/*transactionId*/) {
        logger.info('Move book record success', helper.instruction2string(instruction));
      })
      .catch(function(e) {
        logger.error('Move book record error', helper.instruction2string(instruction), e);
        return updateInstructionStatus(instruction, 'declined');
      });
  }

  /**
   *
   */
  function updateInstructionStatus(instruction, status) {
    logger.debug('set instruction status: %s for %s', status, helper.instruction2string(instruction));

    let channel = configHelper.getInstructionChannel(instruction);
    logger.debug('got channel %s for %s', channel, helper.instruction2string(instruction));

    //
    var args = helper.instructionArguments(instruction);
    args.push(status);
    return invoke.invokeChaincode([endorsePeerHost], channel, 'instruction', 'status', args, USERNAME, ORG)
      .then(function(/*transactionId*/) {
        logger.info('Update instruction status success', helper.instruction2string(instruction));
      })
      .catch(function (e) {
        logger.error('Cannot update instruction status', helper.instruction2string(instruction), e);
      });
  }

  /**
   * Copy balance from 'book' cc to 'position' cc, so it'll be visible for the owner, not only for nsd
   */
  function updatePositionsFromBook() {
    logger.debug('Query book to update all positions');

    //TODO peer0 is inconsistent with explicit peer url in invoke.invokeChaincode. This caused Oleg pain.
    return query.queryChaincode(endorsePeerId, 'depository', 'book', [], 'query', USERNAME, ORG)
      .then(response=>response.result)
      .then(function (result) {
        logger.debug('Query book success', JSON.stringify(result));

        result.forEach(position => {
          logger.trace('Update position', JSON.stringify(position));

          let org = configHelper.getOrgByAccount(position.balance.account, position.balance.division);
          if(!org) {
            logger.error('Cannot find org for position', JSON.stringify(position));
            return;
          }

          //  TODO: rename this bilateral channel
          let channel = 'nsd-' + org;
          logger.debug(`invoking position on ${channel} to put ${position.quantity} of ${position.security} to ${position.balance.account}/${position.balance.division}`);

          //
          var args = [
              position.balance.account,
              position.balance.division,
              position.security,
              '' + position.quantity
           ];
          return invoke.invokeChaincode([endorsePeerHost], channel, 'position', 'put', args, USERNAME, ORG)
            .then(function (/*transactionId*/) {
              logger.info('Put position success', helper.position2string(position));
            })
            .catch(function (e) {
              logger.error('Put position error', helper.position2string(position), e);
              throw e;
            });

        });
      })
      .catch(function (e) {
        logger.error('Cannot query books', e);
      });
  }

};












/**
 * Run {@link param promiseFn} across each element in array sequentially
 *
 * @param {Array} array
 * @param {object} opts
 * @param {object} opts.drop:boolean - don't save result for each promise
 * @param {function} promiseFn
 * @return {Promise}
 *
 * by preliminary estimation the recursive mode takes less memory than iterative,
 * because iterative one allocates memory for the function before any async operation run
 */
function chainPromise(array, opts, promiseFn){
    if(typeof opts === "function" && !promiseFn){
      promiseFn = opts;
      opts = {};
    }

    var i = 0;
    var result = [];

    var collectorFn = opts.drop ? nope : __collectResult;
    function __collectResult(res){
      result.push(res);
    }
    function nope(){}

    function __step(){
        if(i >= array.length){
            return Promise.resolve();
        }
        let item = array[i++];
        return Promise.resolve(promiseFn(item))
            .then(collectorFn)
            .then(__step);
    }

    return __step().then(function(){
      return opts.drop ? null : result;
    });
}
