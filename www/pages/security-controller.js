/**
 * @class SecurityController
 * @classdesc
 * @ngInject
 */
function SecurityController($scope, SecurityService) {

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
    return SecurityService.list()
      .then(function(list){
        ctrl.list = list;
      })
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  }


  ctrl.init();
}

angular.module('nsd.controller.security', ['nsd.service.security'])
.controller('SecurityController', SecurityController);