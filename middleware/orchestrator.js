
module.exports = function(require) {

  let log4js  = require('log4js');
  let logger  = log4js.getLogger('orchestrator');

  const TYPE_ENDORSER_TRANSACTION = 'ENDORSER_TRANSACTION';

  // TODO: move somewhere =)
  const ORG = process.env.ORG || null;
  var config = require('../config.json');

  const USERNAME = config.user.username;

  if(ORG != 'nsd'){
    logger.info('Disabled for common members');
    return;
  }

  logger.info('**************    ORCHESTRATOR     ******************');
  logger.info('Admin     : ' + USERNAME);
  logger.info('Org name  : ' + ORG);



  let invoke = require('../lib-fabric/invoke-transaction.js');
  let peerListener = require('../lib-fabric/peer-listener.js');

  logger.info('registering for block events');

  peerListener.registerBlockEvent(function(block) {
    try {
      let type      = getTransactionType(block);
      let channelId = getTransactionChannel(block);
      logger.info('got block no. %d: event %s on channel %s', block.header.number, type, channelId);

      if(type === TYPE_ENDORSER_TRANSACTION) {

        //TODO collect from all data and actions elements
        let extension = block.data.data[0].payload.data.actions[0].payload.action.proposal_response_payload.extension;

        let event = extension.events;
        logger.debug('got event: ', event.event_name);

        if(event.event_name === 'Instruction.matched') {
          let instruction = JSON.parse(event.payload.toString());
          logger.info('Instruction.matched', instruction);

          //TODO get peer url from network config
          logger.info('Moving balance %s/%s => %s/%s',
              instruction.transferer.account,
              instruction.transferer.division,
              instruction.receiver.account,
              instruction.receiver.division
          );
          invoke.invokeChaincode([ 'peer0.nsd.nsd.ru:7051' ], 'depository', 'book', 'move',
            [
              instruction.transferer.account,
              instruction.transferer.division,
              instruction.receiver.account,
              instruction.receiver.division,
              instruction.security,
              instruction.quantity
            ],
            //TODO user with login 'admin' has to be logged to the web gui, use admin identity that api server operates under
            USERNAME, ORG)
            .then(function(transactionId) {
              return 'executed';
            })
            .catch(function(e) {
              return 'declined';
            })
            .then(function(status) {
              logger.debug('balance moved with result:', status);

              //TODO better way is to have Book raise an event that move was successful,
              // then the orchestrator reacts to it by calling instruction.status().
              // In the above scenario if the orchestrator dies while waiting for transactionId to commit it'll never get set to executed

              // update instruction status
              logger.info('Updating status for', JSON.stringify(instruction));
              return invoke.invokeChaincode([ 'peer0.nsd.nsd.ru:7051' ], channelId, 'instruction', 'status',
                [
                  instruction.deponentFrom,
                  instruction.transferer.account,
                  instruction.transferer.division,

                  instruction.deponentTo,
                  instruction.receiver.account,
                  instruction.receiver.division,

                  instruction.security,
                  instruction.quantity,
                  instruction.reference,
                  instruction.instructionDate,
                  instruction.tradeDate,

                  status
                ],
                USERNAME, ORG)
              .then(function() {
                logger.debug('status set for instruction', status);
              })

          })
        }

      }
    } catch(e) {
      logger.warn(e);
    }

  });

};





function getTransactionType(block){
  return block.data.data[0].payload.header.channel_header.type;
}

function getTransactionChannel(block){
  return block.data.data[0].payload.header.channel_header.channel_id;
}