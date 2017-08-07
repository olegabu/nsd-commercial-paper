/**
 * @class InstructionsController
 * @classdesc
 * @ngInject
 */
function InstructionsController($scope, InstructionService) {

  var ctrl = this;
  ctrl.list = [];

  var DATE_INPUT_FORMAT = 'dd/mm/yyyy';

  /**
   *
   */
  ctrl.reload = function(){
    return InstructionService.list()
      .then(function(list){
        ctrl.list = list;
      });
  }

  ctrl.newInstructionTransfer = function(){
    if(!$scope.inst){
        $scope.inst = {
          trade_date  : new Date().format(DATE_INPUT_FORMAT),
          created     : new Date().format(DATE_INPUT_FORMAT),
          authority:{
            created   : new Date().format(DATE_INPUT_FORMAT)
          }
        };
    }
  };

  ctrl.sendTransfer = function(){
    var instruction = $scope.inst;
    // InstructionService.send();
    // InstructionService.receive();

    $scope.inst = null;
  };
  ctrl.cancelTransfer = function(){
    $scope.inst = null;
  };

  // INIT
  ctrl.reload();

}

angular.module('nsd.controller.instructions', ['nsd.service.instructions'])
.controller('InstructionsController', InstructionsController);