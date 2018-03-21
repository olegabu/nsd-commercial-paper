/**
 *
 */

var ConfigHelper = require('./helper.js').ConfigHelper;


var INSTRUCTION_INIT = require('../instruction_init.json');
var ROLE = process.env.ROLE || 'investor';
var ORG  = process.env.ORG || process.env.THIS_ORG;

if(!INSTRUCTION_INIT){
  throw new Error('INSTRUCTION_INIT environment is not set');
}

if(!ROLE){
  throw new Error('ROLE environment is not set');
}

console.log('INSTRUCTION_INIT', INSTRUCTION_INIT);

// part of original config function with 'account-config' is added
module.exports = function(require, app){

  var packageInfo = require('../package.json');
  var expressEnv  = require('../lib/express-env-middleware');
  var hfc = require('../lib-fabric/hfc.js');

  var clientConfig = hfc.getConfigSetting('config');

  var accountConfig = ConfigHelper.convertAccountConfig(INSTRUCTION_INIT, ROLE);
  clientConfig['account-config'] = accountConfig;
  clientConfig.org = ORG;

  hfc.setConfigSetting('config', clientConfig);


  // get public config
  var clientConfig = hfc.getConfigSetting('config');
  //
  app.get('/config.js', expressEnv('__config', clientConfig));
  app.get('/config', function(req, res) {
      res.setHeader('X-Api-Version', packageInfo.version);
      res.setHeader('X-Api-Middleware', 'Config 0.1');
      res.send(clientConfig);
  });
};




