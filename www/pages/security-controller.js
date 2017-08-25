/**
 * @class SecurityController
 * @classdesc
 * @ngInject
 */
function SecurityController($scope, SecurityService, ConfigLoader) {

  var ctrl = this;

  ctrl.list = [];

  ctrl.accounts = ConfigLoader.getAllAccounts();

  /**
   *
   */
  ctrl.init = function(){
      // $scope.$on('chainblock', ctrl.reload);
      $scope.$on('chainblock-ch-'+ SecurityService.getChannelID(), ctrl.reload);
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


  ctrl.newCalendarEntry = function(){
    $scope.centry = $scope.centry || {
      date: new Date()
    };
  }

  ctrl.sendCEntry = function(centry){
    ctrl.invokeInProgress = true;
    return SecurityService.addCalendarEntry(centry)
      .then(function(){
        $scope.centry = null;
      })
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  }

  ctrl.sendSecurity = function(security){
    ctrl.invokeInProgress = true;
    return SecurityService.sendSecurity(security)
      .then(function(){
        $scope.security = null;
      })
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  }



  ctrl.init();
}

angular.module('nsd.controller.security', ['nsd.service.security'])
.controller('SecurityController', SecurityController);