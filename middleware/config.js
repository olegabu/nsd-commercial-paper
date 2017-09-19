/**
 *
 */
var INSTRUCTION_INIT = process.env.INSTRUCTION_INIT;
var ROLE = process.env.ROLE;
var ORG  = process.env.ORG;

if(!INSTRUCTION_INIT){
  throw new Error('INSTRUCTION_INIT environment is not set');
}

if(!ROLE){
  throw new Error('ROLE environment is not set');
}


// part of original config function with 'account-config' is added
module.exports = function(require, app){

  var packageInfo = require('../package.json');
  var expressEnv  = require('../lib/express-env-middleware');
  var hfc = require('../lib-fabric/hfc.js');

  var clientConfig = hfc.getConfigSetting('config');

  var accountConfig = convertAccountConfig(INSTRUCTION_INIT, ROLE);
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
}







function convertAccountConfig(instructionInit, role){
  role = role || 'investor';
  var obj = JSON.parse(instructionInit);
  var accData = JSON.parse(obj.Args[1]);
  return accData.reduce(function(result, item){

    var account = item.balances.reduce(function(res, it){
      res[it.account] = res[it.account] || [];
      res[it.account].push(it.division);
      return res;
    }, {});

    // HOTFIX: remove domain
    var org = (item.organization.match(/^[\w]+/)||[])[0];

    result[org] = {
      dep  : item.deponent,
      role : role, // value is valid here only for the organisation!
      acc  : account
    };
    return result;
  }, {/*role : role*/});

}
