module.exports = function (require) {

  let log4js = require('log4js');
  let logger = log4js.getLogger('orchestrator');

  const TYPE_ENDORSER_TRANSACTION = 'ENDORSER_TRANSACTION';

  // TODO: move somewhere =)
  const ORG = process.env.ORG || null;
  var config = require('../config.json');
  //TODO get peer url from network config
  const peers = ['peer0.nsd.nsd.ru:7051']

  //TODO user with login 'admin' has to be logged to the web gui, use admin identity that api server operates under
  const USERNAME = config.user.username;

  if (ORG !== 'nsd') {
    logger.info('enabled for nsd only');
    return;
  }

  logger.info('**************    ORCHESTRATOR     ******************');
  logger.info('Admin     : ' + USERNAME);
  logger.info('Org name  : ' + ORG);

  let invoke = require('../lib-fabric/invoke-transaction.js');
  let query = require('../lib-fabric/query.js');
  let peerListener = require('../lib-fabric/peer-listener.js');

  logger.info('registering for block events');

  peerListener.registerBlockEvent(function (block) {
    try {
      //TODO block can have many data elements, don't assume data[0], loop thru data
      block.data.data.forEach(blockData => {

      let type = getTransactionType(blockData);
      let channel = getTransactionChannel(blockData);

      logger.info(`got block no. ${block.header.number}: ${type} on channel ${channel}`);

        if (type === TYPE_ENDORSER_TRANSACTION) {

          blockData.payload.data.actions.forEach(action => {
            let extension = action.payload.action.proposal_response_payload.extension;
            let event = extension.events;
            logger.debug(`event ${event.event_name || 'none'}`);

            if(event.event_name === 'Instruction.matched') {
              let instruction = JSON.parse(event.payload.toString());
              logger.info(event.event_name, instruction);

              moveByInstruction(instruction);
            }

            //TODO event Instruction.declined doesn't seem to be raised by Book as no transaction is committed  
            if(event.event_name === 'Instruction.executed' || event.event_name === 'Instruction.declined') {
              let instruction = JSON.parse(event.payload.toString());
              logger.info(event.event_name, instruction);

              updateInstructionStatus(instruction);
            }

          }); // thru action elements


          if(channel === 'depository') {
            putPositionsFromBook();
          }

        }
      }); // thru block data elements
    }
    catch(e) {
      logger.error('caught while processing block event', e);
    }

  });

  function getTransactionType(blockData) {
    return blockData.payload.header.channel_header.type;
  }

  function getTransactionChannel(blockData) {
    return blockData.payload.header.channel_header.channel_id;
  }

  //TODO read account to org map from config
  function getOrg(o) {
    let ret = null;

    if (o.account === '902') {
      ret = 'a';
    }
    else if (o.account === '903') {
      ret = 'b';
    }

    return ret;
  }

  function moveByInstruction(instruction) {
    logger.info(`invoking book to move ${instruction.quantity} of ${instruction.security} from ${instruction.transferer.account}/${instruction.transferer.division} to ${instruction.receiver.account}/${instruction.receiver.division}`);

    invoke.invokeChaincode(peers, 'depository', 'book', 'move',
      [
        instruction.transferer.account,
        instruction.transferer.division,
        instruction.receiver.account,
        instruction.receiver.division,
        instruction.security,
        instruction.quantity,
        instruction.reference,
        instruction.instructionDate,
        instruction.tradeDate
      ],
      USERNAME, ORG)
    .then(function (transactionId) {
      logger.debug('move success', transactionId);

      //putPositionsFromBook();

      //return 'executed';
    })
    .catch(function(e) {
      logger.error('cannot move', e);
      //return 'declined';
    })
    /*.then(function (status) {
      updateInstructionStatus(instruction, status);
    });*/
  }

  function updateInstructionStatus(instruction) {
    let orgTransferer = getOrg(instruction.transferer);
    let orgReceiver = getOrg(instruction.receiver);

    if(!orgTransferer) {
      logger.error('cannot find orgTransferer', instruction)
    }

    if(!orgReceiver) {
      logger.error('cannot find orgReceiver', instruction)
    }

    let orgs = [orgTransferer, orgReceiver].sort();
    let channel = orgs[0] + '-' + orgs[1];

    logger.info(`invoking instruction on ${channel} to set status ${instruction.status}`, instruction);

    invoke.invokeChaincode(peers, channel, 'instruction', 'status',
      [
        instruction.transferer.account,
        instruction.transferer.division,
        instruction.receiver.account,
        instruction.receiver.division,
        instruction.security,
        instruction.quantity,
        instruction.reference,
        instruction.instructionDate,
        instruction.tradeDate,
        instruction.status
      ],
      USERNAME, ORG)
    .then(function (transactionId) {
      logger.debug('update instruction status success', transactionId);
    })
    .catch(function (e) {
      logger.error('cannot update instruction status', e);
    });
  }

  function putPositionsFromBook() {
    logger.info('querying book to update all positions');

    //TODO peer0 is inconsistent with explicit peer url in invoke.invokeChaincode. This caused Oleg pain.
    query.queryChaincode('peer0', 'depository', 'book', [], 'query', USERNAME, ORG)
    .then(function (res) {
      logger.debug('query success', res);

      res.result.forEach(position => {
        logger.debug('position', position);

        let org = getOrg(position.balance);

        if(!org) {
          logger.error('cannot find org for position', position);
          return;
        }

        let channel = ORG + '-' + org;

        logger.info(`invoking position on ${channel} to put ${position.quantity} of ${position.security} to ${position.balance.account}/${position.balance.division}`);

        invoke.invokeChaincode(peers, channel, 'position', 'put',
          [
            position.balance.account,
            position.balance.division,
            position.security,
            '' + position.quantity
          ],
          USERNAME, ORG)
        .then(function (transactionId) {
          logger.debug('put position success', transactionId);
        })
        .catch(function (e) {
          logger.error('cannot put position', e);
        });

      });
    })
    .catch(function (e) {
      logger.error('cannot query book', e);
    });
  }


};







