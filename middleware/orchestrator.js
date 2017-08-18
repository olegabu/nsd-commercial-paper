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
  const USERNAME = config.user.username;


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

            if(event.event_name === 'Instruction.executed') {
              // instruction is executed, however stil has 'matched' status in ledger (but 'executed' in the event)
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







