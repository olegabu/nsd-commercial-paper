/**
 * @class InstructionsController
 * @classdesc
 * @ngInject
 */
function InstructionsController($scope, InstructionService, ConfigLoader /*, SocketService*/) {

  var ctrl = this;
  ctrl.list = [];

  var DATE_INPUT_FORMAT = 'dd/mm/yyyy';
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
      $scope.$on('chainblock', ctrl.reload);
      ctrl.reload();
  }

  /**
   *
   */
  ctrl.reload = function(){
    ctrl.invokeInProgress = true;
    return InstructionService.list()
      .then(function(list){
        ctrl.list = list;
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
        dep: ConfigLoader.getAccount( transferSide == TRANSFER_SIDE_TRANSFERER ? orgID : opponentID).dep
      },
      receiver:{
        dep: ConfigLoader.getAccount( transferSide == TRANSFER_SIDE_RECEIVER ? orgID : opponentID).dep
      },
      side: transferSide, // deprecate?
      initiator: transferSide,
      quantity: 0,
      trade_date    : new Date().format(DATE_INPUT_FORMAT),
      instruction_date : new Date().format(DATE_INPUT_FORMAT),
      reason:{
        created   : new Date().format(DATE_INPUT_FORMAT)
      }
    };
  }

  ctrl._fillAccount = function(transferSide, opponentID){
      if(transferSide == TRANSFER_SIDE_TRANSFERER){
        ctrl.accountFrom = ConfigLoader.getAccount();
        ctrl.accountTo = opponentID ? ConfigLoader.getAccount(opponentID) : null;
      } else {
        ctrl.accountFrom = opponentID ? ConfigLoader.getAccount(opponentID) : null;
        ctrl.accountTo = ConfigLoader.getAccount();
      }
  };


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


  /**
   *
   */
  ctrl._getOrgIDByChannel = function(channelID){
    return channelID.split('-').filter(function(org){ return org != ctrl.org; })[0];
  }

  /**
   *
   */
  ctrl.sendTransfer = function(){
    var instruction = $scope.inst;
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
   *
   */
  ctrl.cancelTransfer = function(){
    $scope.inst = null;
  };

  //////////////

  // INIT
  ctrl.init();

}

angular.module('nsd.controller.instructions', ['nsd.service.instructions'])
.controller('InstructionsController', InstructionsController);