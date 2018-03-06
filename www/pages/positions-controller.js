/* globals angular */
/**
 * @class PositionsController
 * @classdesc
 * @ngInject
 */
function PositionsController($scope, PositionsService, ConfigLoader, DialogService) {
  "use strict";

  var ctrl = this;

  ctrl.list = [];

  /**
   *
   */
  ctrl.init = function(){
      $scope.$on('chainblock-ch-'+ PositionsService.getChannelID(), ctrl.reload);
      ctrl.reload();
  };

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
  };


  /**
   * @param {Book} book
   */
  ctrl.showHistory = function(book){
    return PositionsService.history(book)
      .then(function(result){
        var scope = {history: result};
        return DialogService.dialog('book-history.html', scope);
      });
  };


  ctrl.init();
}

angular.module('nsd.controller.positions', ['nsd.service.positions'])
.controller('PositionsController', PositionsController);