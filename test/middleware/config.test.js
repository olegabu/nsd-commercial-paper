
var assert = require('assert');
/**
 *
 */
describe('config converter', function(){


  var INSTRUCTION_INIT_EXAMPLE='{"Args":["init","[{\\"organization\\":\\"megafon.nsd.ru\\",\\"deponent\\":\\"CA9861913023\\",\\"balances\\":[{\\"account\\":\\"MZ130605006C\\",\\"division\\":\\"19000000000000000\\"},{\\"account\\":\\"MZ130605006C\\",\\"division\\":\\"22000000000000000\\"}]},{\\"organization\\":\\"raiffeisen.nsd.ru\\",\\"deponent\\":\\"DE000DB7HWY7\\",\\"balances\\":[{\\"account\\":\\"MS980129006C\\",\\"division\\":\\"00000000000000000\\"}]}]"]}';
  /*/ /**/


  var result1 = {
    "megafon":{
      "dep":"CA9861913023",
      "role":"investor",
      "acc":{
        "MZ130605006C":["19000000000000000", "22000000000000000"]
      }
    },
    "raiffeisen":{
      "dep":"DE000DB7HWY7",
      "role":"investor",
      "acc":{
        "MS980129006C":["00000000000000000"]
      }
    },
    // "role":'investor'
  };
  var result2 = clone(result1);
  result2["megafon"].role = 'nsd';
  result2["raiffeisen"].role = 'nsd';
  // result2.role = 'nsd';

  it('sample', function(){

    assert.deepEqual(convertAccountConfig(INSTRUCTION_INIT_EXAMPLE),  result1);
    assert.deepEqual(convertAccountConfig(INSTRUCTION_INIT_EXAMPLE, 'nsd'),  result2);

  });

});



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


function clone(obj){
  return JSON.parse(JSON.stringify(obj));
}