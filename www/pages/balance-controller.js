/**
 * @class BalanceController
 * @classdesc
 * @ngInject
 */
function BalanceController($scope, BalanceService, ConfigLoader) {

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
    return BalanceService.list()
      .then(function(list){
        ctrl.list = list;
      })
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  }


  ctrl.init();
}

angular.module('nsd.controller.balance', ['nsd.service.balance'])
.controller('BalanceController', BalanceController);