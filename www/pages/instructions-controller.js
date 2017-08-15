/**
 * @class InstructionsController
 * @classdesc
 * @ngInject
 */
function InstructionsController($scope, InstructionService, ConfigLoader /*, SocketService*/) {

  var ctrl = this;
  ctrl.list = [];

  // var DATE_INPUT_FORMAT = 'dd/mm/yyyy';
  var DATE_FABRIC_FORMAT = 'yyyy-mm-dd'; // ISO
  var TRANSFER_SIDE_TRANSFERER = 'transferer';
  var TRANSFER_SIDE_RECEIVER = 'receiver';


  ctrl.org = ConfigLoader.get().org;

  // ConfigLoader.getAccount(orgID)
  ctrl.accountFrom = null;
  ctrl.accountTo   = null;


  /**
   *
   */
  ctrl.init = function(){
      $scope.$on('chainblock', function(e, block){
        if( InstructionService.isBilateralChannel(block.getChannel()) ){
          ctrl.reload();
        }
      });
      ctrl.reload();
  }

  /**
   *
   */
  ctrl.reload = function(){
    ctrl.invokeInProgress = true;
    return InstructionService.listAll()
      .then(function(groupedList){

        // flattern: combine all group element into single array
        ctrl.list = Object.keys(groupedList).reduce(function(result, channel){
          result.push.apply(result, groupedList[channel]);
          return result;
        }, []);

      })
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  }


  /**
   * @return {Instruction}
   */
  ctrl._getDefaultinstruction = function(transferSide, opponentID){
    var orgID = ctrl.org;
    return {
      transferer:{
        dep: ctrl._getDeponentCode(transferSide == TRANSFER_SIDE_TRANSFERER ? orgID : opponentID)
      },
      receiver:{
        dep: ctrl._getDeponentCode(transferSide == TRANSFER_SIDE_RECEIVER ? orgID : opponentID)
      },
      side: transferSide, // deprecate?
      initiator: transferSide,
      // quantity: 0, // TODO: cause ui bug with overlapping label and input field with value
      trade_date    : new Date(),//.format(DATE_INPUT_FORMAT),
      instruction_date : new Date(),//.format(DATE_INPUT_FORMAT),
      reason:{
        created   : new Date()//.format(DATE_INPUT_FORMAT)
      }
    };
  }

  ctrl._getDeponentCode = function(orgID){
    var account = ConfigLoader.getAccount(orgID) || {};
    return account.dep;
  }

  /**
   *
   */
  ctrl.getStatusClass = function(status){
    switch(status){
      case 'declined': return 'red-text darken-4';
      case 'matched': return 'deep-purple-text';
      case 'executed': return 'green-text darken-4';
      default: return '';
    }

  }

  /**
   *
   */
  ctrl.newInstructionTransfer = function(transferSide, _channel){
    if(!$scope.inst || $scope.inst.side != transferSide){
        // preset values

        var opponentOrgID = ctrl._getOrgIDByChannel(_channel);
        $scope.inst = ctrl._getDefaultinstruction(transferSide, opponentOrgID);

        // preset
        ctrl._fillAccount(transferSide, opponentOrgID);
    }
  };

  ctrl._fillAccount = function(transferSide, opponentID){
    if(transferSide == TRANSFER_SIDE_TRANSFERER){
      ctrl.accountFrom = ConfigLoader.getAccount(ctrl.org);
      ctrl.accountTo = opponentID ? ConfigLoader.getAccount(opponentID) : null;
    } else {
      ctrl.accountFrom = opponentID ? ConfigLoader.getAccount(opponentID) : null;
      ctrl.accountTo = ConfigLoader.getAccount(ctrl.org);
    }
  };



  /**
   *
   */
  ctrl._getOrgIDByChannel = function(channelID){
    if(!channelID) return null;
    return channelID.split('-').filter(function(org){ return org != ctrl.org; })[0];
  }

  /**
   *
   */
  ctrl.sendInstruction = function(){
    var instruction = $scope.inst;

    // FIXME here date can come in two different formats:
    //  Date object when we change form value
    //  String (like '1 August, 2017') when we not change form value
    // Now we use formatDate() to transform both of it into ISO
    instruction.trade_date        = formatDate(instruction.trade_date);
    instruction.instruction_date  = formatDate(instruction.instruction_date);
    instruction.reason.created    = formatDate(instruction.reason.created);

    var p;
    switch(instruction.side){
      case TRANSFER_SIDE_TRANSFERER:
        p = InstructionService.transfer(instruction);
        break;
      case TRANSFER_SIDE_RECEIVER:
        p = InstructionService.receive(instruction);
        break;
      default:
        throw new Error('Unknpown transfer side: ' + instruction.side);
    }


    ctrl.invokeInProgress = true;
    return p.then(function(){
      $scope.inst = null;
    })
    .finally(function(){
      ctrl.invokeInProgress = false;
    });
  };

  /**
   * Parse date in format dd/mm/yyyy
   * @param {string|Date} dateStr
   * @return {Date}
   */
  function formatDate(date){
    if(!date) return null;

    if(!(date instanceof Date)){
      // assumind date is a string: '1 August, 2017'
      // TODO: we shouldn't rely on this
      date = new Date(date)
    }
    return date.format(DATE_FABRIC_FORMAT);
  }

  /**
   *
   */
  ctrl.cancelInstruction = function(){
    $scope.inst = null;
  };

  //////////////

  // INIT
  ctrl.init();

}

angular.module('nsd.controller.instructions', ['nsd.service.instructions'])
.controller('InstructionsController', InstructionsController);