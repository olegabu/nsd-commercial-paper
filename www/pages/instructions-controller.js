/**
 * @class InstructionsController
 * @classdesc
 * @ngInject
 */
function InstructionsController(InstructionService) {

  var ctrl = this;
  ctrl.list = [];

  /**
   *
   */
  ctrl.reload = function(){
    return InstructionService.list()
      .then(function(list){
        ctrl.list = list;
      });
  }


  // INIT
  ctrl.reload();

}

angular.module('nsd.controller.instructions', ['nsd.service.instructions'])
.controller('InstructionsController', InstructionsController);