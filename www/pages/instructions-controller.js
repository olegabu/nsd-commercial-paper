/**
 * @class InstructionsController
 * @classdesc
 * @ngInject
 */
function InstructionsController($scope, InstructionService, ConfigLoader) {

  var ctrl = this;
  ctrl.list = [];

  var DATE_INPUT_FORMAT = 'dd/mm/yyyy';
  var TRANSFER_SIDE_TRANSFERER = 'transferer';
  var TRANSFER_SIDE_RECEIVER = 'receiver';


  ctrl.org = ConfigLoader.get().org;
  ctrl.account = ConfigLoader.getAccount();

  /**
   *
   */
  ctrl.reload = function(){
    return InstructionService.list()
      .then(function(list){
        ctrl.list = list;
      });
  }


  /**
   * @return {Instruction}
   */
  ctrl._getDefaultinstruction = function(transferSide){
    return {
      transferer:{
        dep: transferSide == TRANSFER_SIDE_TRANSFERER ? ctrl.account.dep : null
      },
      receiver:{
        dep: transferSide == TRANSFER_SIDE_RECEIVER ? ctrl.account.dep : null
      },
      side: transferSide, // deprecate?
      initiator: transferSide,
      trade_date    : new Date().format(DATE_INPUT_FORMAT),
      instruction_date : new Date().format(DATE_INPUT_FORMAT),
      reason:{
        created   : new Date().format(DATE_INPUT_FORMAT)
      }
    };
  }

  /**
   *
   */
  ctrl.newInstructionTransfer = function(transferSide){
    if(!$scope.inst || $scope.inst.side != transferSide){
        // preset values
        $scope.inst = ctrl._getDefaultinstruction(transferSide);
    }
  };

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

    return p.then(function(){
      $scope.inst = null;
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
  ctrl.reload();

}

angular.module('nsd.controller.instructions', ['nsd.service.instructions'])
.controller('InstructionsController', InstructionsController);