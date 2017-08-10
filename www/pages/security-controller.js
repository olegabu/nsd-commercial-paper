/**
 * @class SecurityController
 * @classdesc
 * @ngInject
 */
function SecurityController(SecurityService) {

  var ctrl = this;

  ctrl.list = [];

  /**
   *
   */
  ctrl.init = function(){
      // var socket = SocketService.getSocket();
      // socket.on('chainblock', ctrl.reload);
      // $scope.$on("$destroy", function handler() {
      //     // destruction code here
      //     socket.off('chainblock', ctrl.reload);
      // });
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