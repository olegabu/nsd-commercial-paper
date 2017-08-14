/**
 * @class PositionsController
 * @classdesc
 * @ngInject
 */
function PositionsController($scope, PositionsService, ConfigLoader) {

  var ctrl = this;

  ctrl.list = [];

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
    return PositionsService.list()
      .then(function(list){
        ctrl.list = list;
      })
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  }


  ctrl.init();
}

angular.module('nsd.controller.positions', ['nsd.service.positions'])
.controller('PositionsController', PositionsController);