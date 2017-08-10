/**
 * @class BookController
 * @classdesc
 * @ngInject
 */
function BookController(BookService, ConfigLoader) {

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
    return BookService.list()
      .then(function(list){
        ctrl.list = list;
      })
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  }


  ctrl.init();
}

angular.module('nsd.controller.book', ['nsd.service.book'])
.controller('BookController', BookController);