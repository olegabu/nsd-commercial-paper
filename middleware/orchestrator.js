
module.exports = function(require){
  var log4js  = require('log4js');
  var logger  = log4js.getLogger('orchestrator');


  const ORG = process.env.ORG || null;

  if(ORG != 'nsd'){
    logger.info('Disabled for common members');
    return;
  }

  var peerListener = require('../lib-fabric/peer-listener.js');

  logger.info('registering for block events');

  peerListener.registerBlockEvent(function(block){
    var type = 'Unknown';
    try {
      type = block.data.data[0].payload.header.channel_header.type;
    } catch(e){
      logger.warn(e);
      return;
    }
    logger.info('got block event %s: ', type, block.header.data_hash);

    if(type === 'ENDORSER_TRANSACTION') {
      //TODO collect events from all data and actions elements
      let events = block.data.data[0].payload.data.actions[0].payload.action.proposal_response_payload.extension.events;
      logger.info('events: ', events);
    }

  });

};