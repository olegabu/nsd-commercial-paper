
module.exports = function(require) {
  let log4js  = require('log4js');
  let logger  = log4js.getLogger('orchestrator');

  const ORG = process.env.ORG || null;

  if(ORG != 'nsd'){
    logger.info('Disabled for common members');
    return;
  }

  let invoke = require('../lib-fabric/invoke-transaction.js');
  let peerListener = require('../lib-fabric/peer-listener.js');

  logger.info('registering for block events');

  peerListener.registerBlockEvent(function(block) {
    try {
      let channelId = block.data.data[0].payload.header.channel_header.channel_id;
      let type = block.data.data[0].payload.header.channel_header.type;

      logger.info('got block %d event %s on channel %s: ', block.header.number, type, channelId);

      if(type === 'ENDORSER_TRANSACTION') {
        //TODO collect from all data and actions elements
        let extension =
          block.data.data[0].payload.data.actions[0].payload.action.proposal_response_payload.extension;

        logger.debug('events', extension.events);

        if(extension.events.event_name === 'Instruction.matched') {
          let instruction = JSON.parse(extension.events.payload.toString());
          logger.info('Instruction.matched', instruction);

          //TODO get peer url from network config
          invoke.invokeChaincode([ 'peer0.nsd.nsd.ru:7051' ], 'depository', 'book', 'move',
            [instruction.transferer.account, instruction.transferer.division,
              instruction.receiver.account, instruction.receiver.division,
              instruction.security, instruction.quantity],
            //TODO user with login 'admin' has to be logged to the web gui, use admin identity that api server operates under
            'admin', 'nsd')
          .then(function(transactionId) {
            logger.debug('invoked', transactionId);
            //TODO register to wait for this transaction to commit which means the quantity moved from transferer to receiver
            // then call instruction.status() on channelId to raise the status to 'executed'
            // on error raise status to 'declined'

            //TODO better way is to have Book raise an event that move was successful,
            // then the orchestrator reacts to it by calling instruction.status().
            // In the above scenario if the orchestrator dies while waiting for transactionId to commit it'll never get set to executed
          })
        }

        /*logger.info('response: ', extension.response);

        if(extension.response.payload) {
          let o = JSON.parse(extension.response.payload.toString());
          logger.info('responsePayload: ', o);
        }*/
      }
    } catch(e) {
      logger.warn(e);
    }

  });

};